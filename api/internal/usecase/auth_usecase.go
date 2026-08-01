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
// single-repository WithTransaction pattern used in activity_usecase.go
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
// Middleware bergantung langsung ke interface ini (bukan lewat
// AuthUsecase) karena verifikasi access token tidak butuh apa pun dari
// alur auth lain — cuma butuh parse & validasi signature/exp.
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

// AuthTokens dikembalikan Login & Refresh — dua-duanya mencetak
// pasangan access/refresh baru. RefreshToken di sini raw — satu-
// satunya tempat nilai mentahnya keluar; yang disimpan cuma hash-nya.
type AuthTokens struct {
	User                  *entity.User
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

// LoginResult carries the raw session token — the only place it exists
// outside the client's cookie. It's never persisted; only its hash
// (TokenGenerator.HashToken) is, via UserSessionRepository.Create.
type LoginResult struct {
	User         *entity.User
	SessionToken string
	ExpiresAt    time.Time
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
) AuthUsecase {
	return &authUsecase{
		uow: uow, users: users, userAccounts: userAccounts, refreshTokens: refreshTokens,
		hasher: hasher, accessTokens: accessTokens, refreshTokenGen: refreshTokenGen,
		refreshTokenTTL: refreshTokenTTL,
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
	if err := u.hasher.Compare(*account.PasswordHash, input.Password); err != nil {
		return nil, errs.ErrInvalidCredentials
	}

	// Login selalu memulai chain baru.
	return u.issueSession(ctx, u.refreshTokens, user, uuid.New(), input.IPAddress, input.UserAgent)
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

	// revoked_at sudah terisi berarti token ini bukan lagi ujung
	// chain-nya — entah karena sudah dirotasi sebelumnya, atau memang
	// sedang dipakai ulang (reuse). Dari baris saja dua kasus itu tidak
	// bisa dibedakan, jadi diperlakukan sebagai potensi pencurian: revoke
	// seluruh chain, paksa login ulang.
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

		// Insert dulu baris baru, baru rotate baris lama menunjuk ke
		// baris baru ini — urutan ini penting karena replaced_by_id
		// adalah FK ke refresh_tokens(id).
		newToken, err := repos.RefreshToken.Create(ctx, &entity.RefreshToken{
			UserID: user.ID, FamilyID: existing.FamilyID,
			TokenHash: u.refreshTokenGen.Hash(rawRefresh),
			ExpiresAt: refreshExpiresAt, IPAddress: input.IPAddress, UserAgent: input.UserAgent,
		})
		if err != nil {
			return fmt.Errorf("creating rotated refresh token: %w", err)
		}

		if err := repos.RefreshToken.Rotate(ctx, existing.ID, newToken.ID); err != nil {
			// errs.ErrConflict di sini berarti request Refresh lain
			// (concurrent) sudah lebih dulu merotasi token ini — kita
			// kalah race, seluruh transaksi (termasuk newToken di atas)
			// di-rollback.
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
