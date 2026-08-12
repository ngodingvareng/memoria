package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// --- Repository interfaces, defined here per the Dependency Rule ---

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	// GetByUsername backs @mention resolution and Circle Invite's
	// "invite by username" path. Callers doing the latter must
	// additionally check DiscoverableByUsername — see the
	// GetUserByUsername query comment.
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	// SetUsername claims a username for an account that doesn't have one
	// yet (the post-register onboarding step). Returns
	// errs.ErrUsernameAlreadyExists on a uniqueness conflict.
	SetUsername(ctx context.Context, id uuid.UUID, username string) (*entity.User, error)
	// UpdateImagePath sets (or clears, with nil) the profile photo URL.
	UpdateImagePath(ctx context.Context, id uuid.UUID, imagePath *string) (*entity.User, error)
	// UpdatePrivacySettings overwrites MentionPolicy/CircleInvitePolicy/
	// DiscoverableByUsername/StripPhotoMetadata for user.ID — the
	// Social Interaction + Data Controls toggles (FEATURES.md, Privacy &
	// Control), separate from profile content so a settings screen can
	// save them independently. Other fields on user are ignored.
	UpdatePrivacySettings(ctx context.Context, user *entity.User) (*entity.User, error)
	// SoftDelete starts the recovery grace period (see SoftDeleteUser's
	// query comment) — idempotent, a WHERE clause matching zero rows
	// (already deleted) is a silent no-op. Used by
	// UserUsecase.DeleteAccount.
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type UserAccountRepository interface {
	CreateCredential(ctx context.Context, userID uuid.UUID, accountID, passwordHash string) (*entity.UserAccount, error)
	GetCredentialByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserAccount, error)
	// IncrementFailedLoginAttempts records one more failed password
	// attempt against the credential account and returns the row with
	// its updated count, so the caller can decide whether it just
	// crossed the lockout threshold.
	IncrementFailedLoginAttempts(ctx context.Context, userID uuid.UUID) (*entity.UserAccount, error)
	LockCredentialAccount(ctx context.Context, userID uuid.UUID, until time.Time) error
	ResetFailedLoginAttempts(ctx context.Context, userID uuid.UUID) error
	// UpdatePasswordHash overwrites the credential account's password —
	// the write side of the reset-password flow.
	UpdatePasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) error
	// GetByProvider looks up an OAuth-linked account by provider + the
	// provider's own account id (e.g. Google's `sub` claim). Returns
	// errs.ErrNotFound if this provider identity hasn't been linked to
	// any user yet — the caller decides find-vs-create from there.
	GetByProvider(ctx context.Context, provider enum.AuthProvider, accountID string) (*entity.UserAccount, error)
	// CreateOAuth links a non-credential provider identity to userID.
	// Token/scope fields aren't part of this signature — flows that only
	// verify an ID token (rather than completing an authorization-code
	// exchange) never obtain a provider access/refresh token to store.
	CreateOAuth(
		ctx context.Context,
		userID uuid.UUID,
		provider enum.AuthProvider,
		accountID string,
	) (*entity.UserAccount, error)
}

// UserVerificationRepository backs single-use, expiring, identifier/
// value tokens — currently just the password-reset link, though the
// underlying user_verifications table is generic enough to also serve
// email verification later without a schema change.
type UserVerificationRepository interface {
	Create(ctx context.Context, identifier, value string, expiresAt time.Time) (*entity.UserVerification, error)
	// GetValid returns errs.ErrNotFound if no unexpired row matches —
	// deliberately not distinguishing "wrong value" from "expired" to
	// the caller, since ResetPassword folds both into the same generic
	// "invalid or expired" response.
	GetValid(ctx context.Context, identifier, value string) (*entity.UserVerification, error)
	// Delete is how a verification is consumed — GetValidUserVerification's
	// own doc comment: "the app should still delete it... to prevent
	// reuse." There's no separate consumed_at flow.
	Delete(ctx context.Context, id uuid.UUID) error
	// DeleteByIdentifier invalidates any previously-issued pending
	// tokens for the same identifier before issuing a new one.
	DeleteByIdentifier(ctx context.Context, identifier string) error
}

// Mailer sends the password-reset link. Declared here (consumer-side)
// per this codebase's interface convention — satisfied by
// mailer.SMTPMailer.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *entity.RefreshToken) (*entity.RefreshToken, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	// Rotate atomically marks the id as revoked+replaced, but
	// ONLY if it hasn't been revoked. errs.ErrConflict means it lost the race
	// (concurrent refresh) or the token is no longer the tip of the chain.
	Rotate(ctx context.Context, id, replacedByID uuid.UUID) error
	RevokeFamily(ctx context.Context, familyID uuid.UUID) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
}

// AuthRepositories groups the repositories a single atomic auth
// operation might need. Register has to insert a users row AND a
// user_accounts row as one all-or-nothing unit — neither UserRepository
// nor UserAccountRepository alone can guarantee that.
type AuthRepositories struct {
	User         UserRepository
	UserAccount  UserAccountRepository
	RefreshToken RefreshTokenRepository
}

// AuthUnitOfWork runs fn with repositories all bound to the SAME
// transaction. This is the multi-repository counterpart to the
// single-repository WithTransaction pattern used in thread_usecase.go
// — that pattern only coordinates one repository at a time, which isn't
// enough here since Register spans two.
type AuthUnitOfWork interface {
	WithTransaction(ctx context.Context, fn func(AuthRepositories) error) error
}

// --- Supporting interfaces ---

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

// AccessTokenIssuer mints/verifies the stateless JWT access token.
// Middleware depends directly on this interface (not through
// AuthUsecase) because verifying an access token doesn't need anything
// from the rest of the auth flow — it just needs to parse & validate
// the signature/exp.
type AccessTokenIssuer interface {
	Generate(userID uuid.UUID) (token string, expiresAt time.Time, err error)
	Parse(token string) (uuid.UUID, error)
}

type RefreshTokenGenerator interface {
	Generate() (raw string, err error)
	Hash(raw string) string
}

// GoogleIDTokenVerifier validates a Google Identity Services ID token
// (a signed JWT the frontend obtains directly from Google — not an
// OAuth access token or authorization code) and extracts the identity
// claims LoginWithGoogle needs. Any invalid token collapses to
// errs.ErrInvalidToken.
type GoogleIDTokenVerifier interface {
	Verify(ctx context.Context, idToken string) (*GoogleIdentity, error)
}

// GoogleIdentity is the subset of a verified Google ID token's claims
// LoginWithGoogle acts on.
type GoogleIdentity struct {
	// Subject is Google's `sub` claim — the stable, provider-scoped
	// account id used as UserAccountRepository's accountID, since email
	// alone isn't a safe join key (it can change on Google's side).
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
}

// --- Inputs / outputs ---

type RegisterInput struct {
	Name      string
	Email     string
	Password  string
	IPAddress *string
	UserAgent *string
}

type LoginInput struct {
	Email     string
	Password  string
	IPAddress *string
	UserAgent *string
}

type RefreshInput struct {
	RefreshToken string
	IPAddress    *string
	UserAgent    *string
}

type ResetPasswordInput struct {
	Email       string
	Token       string
	NewPassword string
}

type GoogleLoginInput struct {
	IDToken   string
	IPAddress *string
	UserAgent *string
}

// passwordResetTokenTTL is how long a requested reset link stays valid.
const passwordResetTokenTTL = 30 * time.Minute

// passwordResetIdentifier namespaces user_verifications rows by purpose
// (identifier alone isn't unique to password reset — the table is
// shared with whatever else ends up using it, e.g. email verification
// later) and by email rather than user id, since ForgotPassword only
// has an email to go on before it knows whether an account exists.
func passwordResetIdentifier(email string) string {
	return "password_reset:" + strings.ToLower(email)
}

// AuthTokens is returned by Login & Refresh — both mint a new
// access/refresh pair. RefreshToken here is the raw value — the only
// place the raw value ever leaves; only its hash gets stored.
type AuthTokens struct {
	User                  *entity.User
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

// --- Usecase ---

type AuthUsecase interface {
	Register(ctx context.Context, input RegisterInput) (*AuthTokens, error)
	Login(ctx context.Context, input LoginInput) (*AuthTokens, error)
	Refresh(ctx context.Context, input RefreshInput) (*AuthTokens, error)
	Logout(ctx context.Context, refreshToken string) error
	// ForgotPassword emails a reset link if email belongs to an account
	// — always returns nil either way, silently, so the caller can never
	// use this to probe which emails are registered.
	ForgotPassword(ctx context.Context, email string) error
	// ResetPassword returns errs.ErrInvalidToken for a wrong or expired
	// token — deliberately not distinguishing the two.
	ResetPassword(ctx context.Context, input ResetPasswordInput) error
	// LoginWithGoogle verifies a Google ID token and either logs into
	// the account already linked to it, or — if this Google identity has
	// never signed in before AND its email isn't already registered
	// under a different provider — creates a new account and logs into
	// it. If the email IS already registered elsewhere, this returns
	// errs.ErrEmailAlreadyExists rather than auto-linking; the user
	// should log in with their existing method instead.
	LoginWithGoogle(ctx context.Context, input GoogleLoginInput) (*AuthTokens, error)
}

type authUsecase struct {
	uow             AuthUnitOfWork
	users           UserRepository
	userAccounts    UserAccountRepository
	refreshTokens   RefreshTokenRepository
	verifications   UserVerificationRepository
	hasher          PasswordHasher
	accessTokens    AccessTokenIssuer
	refreshTokenGen RefreshTokenGenerator
	mailer          Mailer
	webBaseURL      string
	refreshTokenTTL time.Duration
	// maxFailedLoginAttempts/loginLockoutDuration implement per-account
	// lockout (see Login below) — independent of any IP-based rate
	// limiting applied at the HTTP layer.
	maxFailedLoginAttempts int
	loginLockoutDuration   time.Duration
	googleVerifier         GoogleIDTokenVerifier
}

func NewAuthUsecase(
	uow AuthUnitOfWork,
	users UserRepository,
	userAccounts UserAccountRepository,
	refreshTokens RefreshTokenRepository,
	verifications UserVerificationRepository,
	hasher PasswordHasher,
	accessTokens AccessTokenIssuer,
	refreshTokenGen RefreshTokenGenerator,
	mailer Mailer,
	webBaseURL string,
	refreshTokenTTL time.Duration,
	maxFailedLoginAttempts int,
	loginLockoutDuration time.Duration,
	googleVerifier GoogleIDTokenVerifier,
) AuthUsecase {
	return &authUsecase{
		uow: uow, users: users, userAccounts: userAccounts, refreshTokens: refreshTokens,
		verifications: verifications,
		hasher:        hasher, accessTokens: accessTokens, refreshTokenGen: refreshTokenGen,
		mailer:                 mailer,
		webBaseURL:             webBaseURL,
		refreshTokenTTL:        refreshTokenTTL,
		maxFailedLoginAttempts: maxFailedLoginAttempts,
		loginLockoutDuration:   loginLockoutDuration,
		googleVerifier:         googleVerifier,
	}
}

// issueSession mints a fresh access/refresh pair and persists the
// refresh token — the shared final step of Register, Login, Refresh,
// and LoginWithGoogle.
func (u *authUsecase) issueSession(
	ctx context.Context,
	repo RefreshTokenRepository,
	user *entity.User,
	familyID uuid.UUID,
	ip, ua *string,
) (*AuthTokens, error) {
	accessToken, accessExpiresAt, err := u.accessTokens.Generate(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}
	rawRefresh, err := u.refreshTokenGen.Generate()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}
	refreshExpiresAt := time.Now().Add(u.refreshTokenTTL)

	if _, err := repo.Create(ctx, &entity.RefreshToken{
		UserID: user.ID, FamilyID: familyID,
		TokenHash: u.refreshTokenGen.Hash(rawRefresh),
		ExpiresAt: refreshExpiresAt, IPAddress: ip, UserAgent: ua,
	}); err != nil {
		return nil, fmt.Errorf("creating refresh token: %w", err)
	}

	return &AuthTokens{
		User: user, AccessToken: accessToken, AccessTokenExpiresAt: accessExpiresAt,
		RefreshToken: rawRefresh, RefreshTokenExpiresAt: refreshExpiresAt,
	}, nil
}
