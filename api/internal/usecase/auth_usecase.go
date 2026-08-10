package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
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
	}
}

func (u *authUsecase) Register(ctx context.Context, input RegisterInput) (*AuthTokens, error) {
	existing, err := u.users.GetByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return nil, fmt.Errorf("checking existing email: %w", err)
	}
	if existing != nil {
		return nil, errs.ErrEmailAlreadyExists
	}

	hashedPassword, err := u.hasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	var tokens *AuthTokens
	err = u.uow.WithTransaction(ctx, func(repos AuthRepositories) error {
		user, err := repos.User.Create(ctx, &entity.User{
			Name: input.Name, Email: input.Email, Timezone: "UTC",
		})
		if err != nil {
			return fmt.Errorf("creating user: %w", err)
		}
		if _, err := repos.UserAccount.CreateCredential(ctx, user.ID, user.ID.String(), hashedPassword); err != nil {
			return fmt.Errorf("creating credential account: %w", err)
		}
		// Register doubles as login — the account is unusable for the
		// welcome/username-claim step otherwise, and there's no reason
		// to make the frontend log in a second time right after signup.
		t, err := u.issueSession(ctx, repos.RefreshToken, user, uuid.New(), input.IPAddress, input.UserAgent)
		if err != nil {
			return fmt.Errorf("issuing session: %w", err)
		}
		tokens = t
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("registering user: %w", err)
	}
	return tokens, nil
}

func (u *authUsecase) Login(ctx context.Context, input LoginInput) (*AuthTokens, error) {
	user, err := u.users.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("looking up user: %w", err)
	}

	account, err := u.userAccounts.GetCredentialByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("looking up credential account: %w", err)
	}
	if account.PasswordHash == nil {
		return nil, errs.ErrInvalidCredentials
	}

	if account.LockedUntil != nil && account.LockedUntil.After(time.Now()) {
		return nil, errs.ErrAccountLocked
	}

	if err := u.hasher.Compare(*account.PasswordHash, input.Password); err != nil {
		if regErr := u.registerFailedLogin(ctx, user.ID); regErr != nil {
			return nil, fmt.Errorf("registering failed login: %w", regErr)
		}
		return nil, errs.ErrInvalidCredentials
	}

	if account.FailedLoginAttempts > 0 || account.LockedUntil != nil {
		if err := u.userAccounts.ResetFailedLoginAttempts(ctx, user.ID); err != nil {
			return nil, fmt.Errorf("resetting failed login attempts: %w", err)
		}
	}

	// Login always starts a fresh chain.
	return u.issueSession(ctx, u.refreshTokens, user, uuid.New(), input.IPAddress, input.UserAgent)
}

// registerFailedLogin records one more failed password attempt and, once
// the count reaches maxFailedLoginAttempts, locks the account for
// loginLockoutDuration. Locking re-triggers on every failed attempt made
// after the count is already at/above the threshold (including right
// after a previous lockout expires), which is deliberately more
// aggressive than resetting the count on expiry alone.
func (u *authUsecase) registerFailedLogin(ctx context.Context, userID uuid.UUID) error {
	updated, err := u.userAccounts.IncrementFailedLoginAttempts(ctx, userID)
	if err != nil {
		return fmt.Errorf("incrementing failed login attempts: %w", err)
	}
	if int(updated.FailedLoginAttempts) >= u.maxFailedLoginAttempts {
		if err := u.userAccounts.LockCredentialAccount(ctx, userID, time.Now().Add(u.loginLockoutDuration)); err != nil {
			return fmt.Errorf("locking account: %w", err)
		}
	}
	return nil
}

func (u *authUsecase) issueSession(ctx context.Context, repo RefreshTokenRepository, user *entity.User, familyID uuid.UUID, ip, ua *string) (*AuthTokens, error) {
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

func (u *authUsecase) Refresh(ctx context.Context, input RefreshInput) (*AuthTokens, error) {
	tokenHash := u.refreshTokenGen.Hash(input.RefreshToken)

	existing, err := u.refreshTokens.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.ErrUnauthorized
		}
		return nil, fmt.Errorf("looking up refresh token: %w", err)
	}

	// revoked_at already set means this token is no longer the tip
	// of its chain, either because it was already rotated, or
	// because it's being reused. The row alone can't tell those two
	// cases apart, so treat it as potential theft: revoke the whole
	// chain and force a fresh login.
	if existing.RevokedAt != nil {
		if err := u.refreshTokens.RevokeFamily(ctx, existing.FamilyID); err != nil {
			return nil, fmt.Errorf("revoking compromised token family: %w", err)
		}
		return nil, errs.ErrUnauthorized
	}
	if existing.ExpiresAt.Before(time.Now()) {
		return nil, errs.ErrTokenExpired
	}

	user, err := u.users.GetByID(ctx, existing.UserID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.ErrUnauthorized
		}
		return nil, fmt.Errorf("looking up refresh token user: %w", err)
	}

	var tokens *AuthTokens
	err = u.uow.WithTransaction(ctx, func(repos AuthRepositories) error {
		accessToken, accessExpiresAt, err := u.accessTokens.Generate(user.ID)
		if err != nil {
			return fmt.Errorf("generating access token: %w", err)
		}
		rawRefresh, err := u.refreshTokenGen.Generate()
		if err != nil {
			return fmt.Errorf("generating refresh token: %w", err)
		}
		refreshExpiresAt := time.Now().Add(u.refreshTokenTTL)

		// Insert the new row first, then rotate the old row to point at
		// this new one — the order matters because replaced_by_id is a
		// FK into refresh_tokens(id).
		newToken, err := repos.RefreshToken.Create(ctx, &entity.RefreshToken{
			UserID: user.ID, FamilyID: existing.FamilyID,
			TokenHash: u.refreshTokenGen.Hash(rawRefresh),
			ExpiresAt: refreshExpiresAt, IPAddress: input.IPAddress, UserAgent: input.UserAgent,
		})
		if err != nil {
			return fmt.Errorf("creating rotated refresh token: %w", err)
		}

		if err := repos.RefreshToken.Rotate(ctx, existing.ID, newToken.ID); err != nil {
			// errs.ErrConflict here means another (concurrent) Refresh
			// request already rotated this token first, we lost the
			// race, and the whole transaction (including newToken
			// above) gets rolled back.
			return fmt.Errorf("rotating refresh token: %w", err)
		}

		tokens = &AuthTokens{
			User: user, AccessToken: accessToken, AccessTokenExpiresAt: accessExpiresAt,
			RefreshToken: rawRefresh, RefreshTokenExpiresAt: refreshExpiresAt,
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errs.ErrConflict) {
			return nil, errs.ErrUnauthorized
		}
		return nil, fmt.Errorf("refreshing session: %w", err)
	}
	return tokens, nil
}

func (u *authUsecase) Logout(ctx context.Context, refreshToken string) error {
	row, err := u.refreshTokens.GetByTokenHash(ctx, u.refreshTokenGen.Hash(refreshToken))
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil // idempotent
		}
		return fmt.Errorf("looking up refresh token: %w", err)
	}
	if row.RevokedAt != nil {
		return nil
	}
	if err := u.refreshTokens.Revoke(ctx, row.ID); err != nil {
		return fmt.Errorf("revoking refresh token: %w", err)
	}
	return nil
}

// ForgotPassword implements [AuthUsecase]. Never reveals whether email
// belongs to an account — the "user not found" path and the "sent"
// path both return nil, on purpose.
func (u *authUsecase) ForgotPassword(ctx context.Context, email string) error {
	user, err := u.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("resolving email: %w", err)
	}

	identifier := passwordResetIdentifier(email)
	// Invalidate any earlier pending reset before issuing a new one —
	// same "resend OTP" reasoning as DeleteUserVerificationsByIdentifier's
	// own doc comment.
	if err := u.verifications.DeleteByIdentifier(ctx, identifier); err != nil {
		return fmt.Errorf("clearing previous reset tokens: %w", err)
	}

	rawToken, err := u.refreshTokenGen.Generate()
	if err != nil {
		return fmt.Errorf("generating reset token: %w", err)
	}
	if _, err := u.verifications.Create(ctx, identifier, u.refreshTokenGen.Hash(rawToken), time.Now().Add(passwordResetTokenTTL)); err != nil {
		return fmt.Errorf("storing reset token: %w", err)
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s&email=%s",
		strings.TrimSuffix(u.webBaseURL, "/"), url.QueryEscape(rawToken), url.QueryEscape(email))
	body := fmt.Sprintf(
		"Hi %s,\n\nSomeone requested a password reset for your Memoria account. If this was you, use the link below within 30 minutes to set a new password:\n\n%s\n\nIf you didn't request this, you can safely ignore this email.",
		user.Name, resetLink,
	)
	if err := u.mailer.Send(ctx, email, "Reset your Memoria password", body); err != nil {
		return fmt.Errorf("sending reset email: %w", err)
	}
	return nil
}

// ResetPassword implements [AuthUsecase].
func (u *authUsecase) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	identifier := passwordResetIdentifier(input.Email)
	verification, err := u.verifications.GetValid(ctx, identifier, u.refreshTokenGen.Hash(input.Token))
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrInvalidToken
		}
		return fmt.Errorf("validating reset token: %w", err)
	}

	user, err := u.users.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrInvalidToken
		}
		return fmt.Errorf("resolving email: %w", err)
	}

	passwordHash, err := u.hasher.Hash(input.NewPassword)
	if err != nil {
		return fmt.Errorf("hashing new password: %w", err)
	}
	if err := u.userAccounts.UpdatePasswordHash(ctx, user.ID, passwordHash); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	// Best-effort from here: the password change already succeeded,
	// none of this should turn a successful reset into an error
	// response.
	if err := u.verifications.Delete(context.WithoutCancel(ctx), verification.ID); err != nil {
		slog.WarnContext(ctx, "failed to consume password reset token", "error", err)
	}
	if err := u.userAccounts.ResetFailedLoginAttempts(context.WithoutCancel(ctx), user.ID); err != nil {
		slog.WarnContext(ctx, "failed to clear login lockout after password reset", "error", err)
	}
	// Force re-login everywhere — a password reset is exactly the
	// moment a stolen session should stop working.
	if err := u.refreshTokens.RevokeAllByUserID(context.WithoutCancel(ctx), user.ID); err != nil {
		slog.WarnContext(ctx, "failed to revoke sessions after password reset", "error", err)
	}

	return nil
}
