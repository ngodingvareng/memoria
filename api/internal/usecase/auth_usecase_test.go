//go:build unit

package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/usecase/mocks"
)

func newAuthUsecase(t *testing.T) (
	usecase.AuthUsecase,
	*mocks.MockUserRepository,
	*mocks.MockUserAccountRepository,
	*mocks.MockUserSessionRepository,
	*mocks.MockAuthUnitOfWork,
	*mocks.MockPasswordHasher,
	*mocks.MockTokenGenerator,
) {
	users := mocks.NewMockUserRepository(t)
	userAccounts := mocks.NewMockUserAccountRepository(t)
	sessions := mocks.NewMockUserSessionRepository(t)
	uow := mocks.NewMockAuthUnitOfWork(t)
	hasher := mocks.NewMockPasswordHasher(t)
	tokens := mocks.NewMockTokenGenerator(t)

	uc := usecase.NewAuthUsecase(uow, users, userAccounts, sessions, hasher, tokens)
	return uc, users, userAccounts, sessions, uow, hasher, tokens
}

// --- Register ---

func TestAuthUsecase_Register_Success(t *testing.T) {
	uc, users, userAccounts, _, uow, hasher, _ := newAuthUsecase(t)

	users.EXPECT().GetByEmail(mock.Anything, "budi@example.com").Return(nil, errs.ErrNotFound)
	hasher.EXPECT().Hash("s3cur3-password").Return("hashed-password", nil)

	created := &entity.User{ID: uuid.New(), Name: "Budi", Email: "budi@example.com"}
	uow.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(usecase.AuthRepositories) error) error {
			return fn(usecase.AuthRepositories{User: users, UserAccount: userAccounts})
		})
	users.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.Name == "Budi" && u.Email == "budi@example.com" && u.Timezone == "UTC"
		})).
		Return(created, nil)
	userAccounts.EXPECT().
		CreateCredential(mock.Anything, created.ID, created.ID.String(), "hashed-password").
		Return(&entity.UserAccount{}, nil)

	result, err := uc.Register(context.Background(), usecase.RegisterInput{
		Name:     "Budi",
		Email:    "budi@example.com",
		Password: "s3cur3-password",
	})

	assert.NoError(t, err)
	assert.Equal(t, created, result)
}

func TestAuthUsecase_Register_EmailAlreadyExists(t *testing.T) {
	uc, users, _, _, _, _, _ := newAuthUsecase(t)

	users.EXPECT().GetByEmail(mock.Anything, "budi@example.com").
		Return(&entity.User{ID: uuid.New(), Email: "budi@example.com"}, nil)
	// hasher.Hash and uow.WithTransaction are deliberately never stubbed —
	// if Register called either after finding an existing email, this
	// test fails with "unexpected call" instead of silently passing.

	result, err := uc.Register(context.Background(), usecase.RegisterInput{
		Name:     "Budi",
		Email:    "budi@example.com",
		Password: "whatever",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrEmailAlreadyExists)
}

func TestAuthUsecase_Register_HashingFails(t *testing.T) {
	uc, users, _, _, _, hasher, _ := newAuthUsecase(t)

	users.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(nil, errs.ErrNotFound)
	wantErr := errors.New("hashing exploded")
	hasher.EXPECT().Hash(mock.Anything).Return("", wantErr)

	result, err := uc.Register(context.Background(), usecase.RegisterInput{
		Name: "Budi", Email: "budi@example.com", Password: "whatever",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}

func TestAuthUsecase_Register_TransactionFails(t *testing.T) {
	uc, users, userAccounts, _, uow, hasher, _ := newAuthUsecase(t)

	users.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(nil, errs.ErrNotFound)
	hasher.EXPECT().Hash(mock.Anything).Return("hashed", nil)

	wantErr := errors.New("db exploded mid-transaction")
	uow.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(usecase.AuthRepositories) error) error {
			return fn(usecase.AuthRepositories{User: users, UserAccount: userAccounts})
		})
	users.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, wantErr)
	// userAccounts.CreateCredential must NOT be called if creating the
	// user itself already failed — not stubbing it turns that into a
	// hard test failure instead of silently passing.

	result, err := uc.Register(context.Background(), usecase.RegisterInput{
		Name: "Budi", Email: "budi@example.com", Password: "whatever",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}

// --- Login ---

func TestAuthUsecase_Login_Success(t *testing.T) {
	uc, users, userAccounts, sessions, _, hasher, tokens := newAuthUsecase(t)

	user := &entity.User{ID: uuid.New(), Email: "budi@example.com"}
	passwordHash := "scrypt-hash"
	users.EXPECT().GetByEmail(mock.Anything, "budi@example.com").Return(user, nil)
	userAccounts.EXPECT().GetCredentialByUserID(mock.Anything, user.ID).
		Return(&entity.UserAccount{PasswordHash: &passwordHash}, nil)
	hasher.EXPECT().Compare(passwordHash, "correct-password").Return(nil)
	tokens.EXPECT().GenerateSessionToken().Return("raw-token", nil)
	tokens.EXPECT().HashToken("raw-token").Return("hashed-token")
	sessions.EXPECT().Create(mock.Anything, mock.MatchedBy(func(s *entity.UserSession) bool {
		return s.UserID == user.ID && s.TokenHash == "hashed-token"
	})).Return(&entity.UserSession{}, nil)

	result, err := uc.Login(context.Background(), usecase.LoginInput{
		Email: "budi@example.com", Password: "correct-password",
	})

	assert.NoError(t, err)
	assert.Equal(t, user, result.User)
	assert.Equal(t, "raw-token", result.SessionToken)
}

func TestAuthUsecase_Login_EmailNotFound(t *testing.T) {
	uc, users, _, _, _, _, _ := newAuthUsecase(t)

	users.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(nil, errs.ErrNotFound)

	result, err := uc.Login(context.Background(), usecase.LoginInput{
		Email: "nobody@example.com", Password: "whatever",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrInvalidCredentials,
		"must NOT leak errs.ErrNotFound here — that would let an attacker enumerate registered emails")
}

func TestAuthUsecase_Login_NoCredentialAccount(t *testing.T) {
	uc, users, userAccounts, _, _, _, _ := newAuthUsecase(t)

	user := &entity.User{ID: uuid.New()}
	users.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(user, nil)
	userAccounts.EXPECT().GetCredentialByUserID(mock.Anything, user.ID).Return(nil, errs.ErrNotFound)

	_, err := uc.Login(context.Background(), usecase.LoginInput{Email: "x@example.com", Password: "y"})

	assert.ErrorIs(t, err, errs.ErrInvalidCredentials)
}

func TestAuthUsecase_Login_WrongPassword(t *testing.T) {
	uc, users, userAccounts, _, _, hasher, _ := newAuthUsecase(t)

	user := &entity.User{ID: uuid.New()}
	hash := "some-hash"
	users.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(user, nil)
	userAccounts.EXPECT().GetCredentialByUserID(mock.Anything, user.ID).
		Return(&entity.UserAccount{PasswordHash: &hash}, nil)
	hasher.EXPECT().Compare(hash, "wrong-password").Return(errors.New("mismatch"))

	_, err := uc.Login(context.Background(), usecase.LoginInput{Email: "x@example.com", Password: "wrong-password"})

	assert.ErrorIs(t, err, errs.ErrInvalidCredentials)
}

func TestAuthUsecase_Login_NilPasswordHash_OAuthOnlyAccount(t *testing.T) {
	uc, users, userAccounts, _, _, _, _ := newAuthUsecase(t)

	user := &entity.User{ID: uuid.New()}
	users.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(user, nil)
	userAccounts.EXPECT().GetCredentialByUserID(mock.Anything, user.ID).
		Return(&entity.UserAccount{PasswordHash: nil}, nil)
	// hasher.Compare deliberately not stubbed — must never be called
	// against a nil hash.

	_, err := uc.Login(context.Background(), usecase.LoginInput{Email: "x@example.com", Password: "y"})

	assert.ErrorIs(t, err, errs.ErrInvalidCredentials)
}

// --- Logout ---

func TestAuthUsecase_Logout_Success(t *testing.T) {
	uc, _, _, sessions, _, _, tokens := newAuthUsecase(t)

	tokens.EXPECT().HashToken("raw-token").Return("hashed-token")
	sessions.EXPECT().DeleteByTokenHash(mock.Anything, "hashed-token").Return(nil)

	err := uc.Logout(context.Background(), "raw-token")

	assert.NoError(t, err)
}

func TestAuthUsecase_Logout_Error(t *testing.T) {
	uc, _, _, sessions, _, _, tokens := newAuthUsecase(t)

	wantErr := errors.New("db exploded")
	tokens.EXPECT().HashToken(mock.Anything).Return("hashed-token")
	sessions.EXPECT().DeleteByTokenHash(mock.Anything, mock.Anything).Return(wantErr)

	err := uc.Logout(context.Background(), "raw-token")

	assert.ErrorIs(t, err, wantErr)
}

// --- ValidateSession ---

func TestAuthUsecase_ValidateSession_Success(t *testing.T) {
	uc, users, _, sessions, _, _, tokens := newAuthUsecase(t)

	user := &entity.User{ID: uuid.New()}
	session := &entity.UserSession{UserID: user.ID, ExpiresAt: time.Now().Add(1 * time.Hour)}

	tokens.EXPECT().HashToken("raw-token").Return("hashed-token")
	sessions.EXPECT().GetByTokenHash(mock.Anything, "hashed-token").Return(session, nil)
	users.EXPECT().GetByID(mock.Anything, user.ID).Return(user, nil)

	result, err := uc.ValidateSession(context.Background(), "raw-token")

	assert.NoError(t, err)
	assert.Equal(t, user, result)
}

func TestAuthUsecase_ValidateSession_SessionNotFound(t *testing.T) {
	uc, _, _, sessions, _, _, tokens := newAuthUsecase(t)

	tokens.EXPECT().HashToken(mock.Anything).Return("hashed-token")
	sessions.EXPECT().GetByTokenHash(mock.Anything, mock.Anything).Return(nil, errs.ErrNotFound)

	_, err := uc.ValidateSession(context.Background(), "raw-token")

	assert.ErrorIs(t, err, errs.ErrUnauthorized)
}

func TestAuthUsecase_ValidateSession_Expired(t *testing.T) {
	uc, _, _, sessions, _, _, tokens := newAuthUsecase(t)

	session := &entity.UserSession{UserID: uuid.New(), ExpiresAt: time.Now().Add(-1 * time.Hour)}
	tokens.EXPECT().HashToken(mock.Anything).Return("hashed-token")
	sessions.EXPECT().GetByTokenHash(mock.Anything, mock.Anything).Return(session, nil)
	// users.GetByID deliberately not stubbed — an expired session should
	// short-circuit before ever looking up the user.

	_, err := uc.ValidateSession(context.Background(), "raw-token")

	assert.ErrorIs(t, err, errs.ErrTokenExpired)
}

func TestAuthUsecase_ValidateSession_UserNotFound(t *testing.T) {
	uc, users, _, sessions, _, _, tokens := newAuthUsecase(t)

	session := &entity.UserSession{UserID: uuid.New(), ExpiresAt: time.Now().Add(1 * time.Hour)}
	tokens.EXPECT().HashToken(mock.Anything).Return("hashed-token")
	sessions.EXPECT().GetByTokenHash(mock.Anything, mock.Anything).Return(session, nil)
	users.EXPECT().GetByID(mock.Anything, session.UserID).Return(nil, errs.ErrNotFound)

	_, err := uc.ValidateSession(context.Background(), "raw-token")

	assert.ErrorIs(t, err, errs.ErrUnauthorized)
}
