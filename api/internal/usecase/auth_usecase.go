package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/pkg/errs"
)

const (
	sessionTTL          = 30 * 24 * time.Hour // 30 days
	defaultUserTimezone = "UTC"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}
type UserAccountRepository interface {
	CreateCredential(ctx context.Context, userID uuid.UUID, accountID, passwordHash string) (*entity.UserAccount, error)
	GetCredentialByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserAccount, error)
}
type UserSessionRepository interface {
	Create(ctx context.Context, session *entity.UserSession) (*entity.UserSession, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.UserSession, error)
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}

// AuthRepositories groups the repositories a single atomic auth
// operation might need. Register has to insert a users row AND a
// user_accounts row as one all-or-nothing unit — neither UserRepository
// nor UserAccountRepository alone can guarantee that.
type AuthRepositories struct {
	User        UserRepository
	UserAccount UserAccountRepository
}

// AuthUnitOfWork runs fn with repositories all bound to the SAME
// transaction. This is the multi-repository counterpart to the
// single-repository WithTransaction pattern used in activity_usecase.go
// — that pattern only coordinates one repository at a time, which isn't
// enough here since Register spans two.
type AuthUnitOfWork interface {
	WithTransaction(ctx context.Context, fn func(AuthRepositories) error) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type TokenGenerator interface {
	GenerateSessionToken() (string, error)
	HashToken(rawToken string) string
}

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

type LoginResult struct {
	User         *entity.User
	SessionToken string
	ExpiresAt    time.Time
}

type AuthUsecase interface {
	Register(ctx context.Context, input RegisterInput) (*entity.User, error)
	Login(ctx context.Context, input LoginInput) (*LoginResult, error)
	Logout(ctx context.Context, sessionToken string) error
	ValidateSession(ctx context.Context, sessionToken string) (*entity.User, error)
}

type authUsecase struct {
	uow          AuthUnitOfWork
	users        UserRepository
	userAccounts UserAccountRepository
	userSessions UserSessionRepository
	hasher       PasswordHasher
	tokens       TokenGenerator
}

func NewAuthUsecase(
	uow AuthUnitOfWork,
	users UserRepository,
	userAccounts UserAccountRepository,
	userSessions UserSessionRepository,
	hasher PasswordHasher,
	tokens TokenGenerator,
) AuthUsecase {
	return &authUsecase{
		uow:          uow,
		users:        users,
		userAccounts: userAccounts,
		userSessions: userSessions,
		hasher:       hasher,
		tokens:       tokens,
	}
}

// Register implements [AuthUsecase].
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
			Name:     input.Name,
			Email:    input.Email,
			Timezone: defaultUserTimezone,
		})
		if err != nil {
			return fmt.Errorf("creating user: %w", err)
		}

		// account_id for the 'credential' provider is just the user's
		// own id — there's no external account identifier to store here,
		// unlike OAuth providers (google/github).
		if _, err := repos.UserAccount.CreateCredential(ctx, user.ID, user.ID.String(), hashedPassword); err != nil {
			return fmt.Errorf("creating credential account: %w", err)
		}

		created = user
		return nil
	})
	if err != nil {
		// The GetByEmail check above is a nicer-error-message convenience,
		// not the real guard against a race (two concurrent registrations
		// with the same email) — the repository's Create translates the
		// DB's unique constraint violation into errs.ErrEmailAlreadyExists
		// too, so that race is still caught correctly here.
		return nil, fmt.Errorf("registering user: %w", err)
	}

	return created, nil
}

// Login implements [AuthUsecase].
func (u *authUsecase) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	panic("unimplemented")
}

// Logout implements [AuthUsecase].
func (u *authUsecase) Logout(ctx context.Context, sessionToken string) error {
	panic("unimplemented")
}

// ValidateSession implements [AuthUsecase].
func (u *authUsecase) ValidateSession(ctx context.Context, sessionToken string) (*entity.User, error) {
	panic("unimplemented")
}
