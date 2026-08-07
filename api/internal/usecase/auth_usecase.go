package usecase

import (
	"context"
	"errors"
	"fmt"
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
	Name     string
	Email    string
	Password string
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
	Register(ctx context.Context, input RegisterInput) (*entity.User, error)
	Login(ctx context.Context, input LoginInput) (*AuthTokens, error)
	Refresh(ctx context.Context, input RefreshInput) (*AuthTokens, error)
	Logout(ctx context.Context, refreshToken string) error
}

type authUsecase struct {
	uow             AuthUnitOfWork
	users           UserRepository
	userAccounts    UserAccountRepository
	refreshTokens   RefreshTokenRepository
	hasher          PasswordHasher
	accessTokens    AccessTokenIssuer
	refreshTokenGen RefreshTokenGenerator
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
	hasher PasswordHasher,
	accessTokens AccessTokenIssuer,
	refreshTokenGen RefreshTokenGenerator,
	refreshTokenTTL time.Duration,
	maxFailedLoginAttempts int,
	loginLockoutDuration time.Duration,
) AuthUsecase {
	return &authUsecase{
		uow: uow, users: users, userAccounts: userAccounts, refreshTokens: refreshTokens,
		hasher: hasher, accessTokens: accessTokens, refreshTokenGen: refreshTokenGen,
		refreshTokenTTL:        refreshTokenTTL,
		maxFailedLoginAttempts: maxFailedLoginAttempts,
		loginLockoutDuration:   loginLockoutDuration,
	}
}

func (u *authUsecase) Register(ctx context.Context, input RegisterInput) (*entity.User, error) {
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

	var created *entity.User
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
		created = user
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("registering user: %w", err)
	}
	return created, nil
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
