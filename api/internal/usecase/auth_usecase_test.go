//go:build unit

package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/usecase/mocks"
)

// testRefreshTokenTTL is an arbitrary fixed duration passed into every
// usecase instance under test — the exact value doesn't matter for
// these tests, only that AuthTokens.RefreshTokenExpiresAt lands roughly
// time.Now()+ttl, which the affected tests assert with WithinDuration
// rather than exact equality (constructing the token happens a few
// nanoseconds after time.Now() is read here).
const testRefreshTokenTTL = 30 * 24 * time.Hour

// testMaxFailedLoginAttempts/testLoginLockoutDuration are the lockout
// parameters every test usecase instance is built with.
const (
	testMaxFailedLoginAttempts = 5
	testLoginLockoutDuration   = 15 * time.Minute
)

const testWebBaseURL = "http://localhost:5173"

func newAuthUsecase(t *testing.T) (
	usecase.AuthUsecase,
	*mocks.MockUserRepository,
	*mocks.MockUserAccountRepository,
	*mocks.MockRefreshTokenRepository,
	*mocks.MockAuthUnitOfWork,
	*mocks.MockPasswordHasher,
	*mocks.MockAccessTokenIssuer,
	*mocks.MockRefreshTokenGenerator,
	*mocks.MockUserVerificationRepository,
	*mocks.MockMailer,
) {
	users := mocks.NewMockUserRepository(t)
	userAccounts := mocks.NewMockUserAccountRepository(t)
	refreshTokens := mocks.NewMockRefreshTokenRepository(t)
	uow := mocks.NewMockAuthUnitOfWork(t)
	hasher := mocks.NewMockPasswordHasher(t)
	accessTokens := mocks.NewMockAccessTokenIssuer(t)
	refreshTokenGen := mocks.NewMockRefreshTokenGenerator(t)
	verifications := mocks.NewMockUserVerificationRepository(t)
	mailer := mocks.NewMockMailer(t)

	uc := usecase.NewAuthUsecase(uow, users, userAccounts, refreshTokens, verifications, hasher, accessTokens, refreshTokenGen, mailer, testWebBaseURL, testRefreshTokenTTL, testMaxFailedLoginAttempts, testLoginLockoutDuration)
	return uc, users, userAccounts, refreshTokens, uow, hasher, accessTokens, refreshTokenGen, verifications, mailer
}

// --- Register ---
// Register creates the account WITHOUT a username (claimed later via
// UserUsecase.SetUsername in the welcome/onboarding step) and doubles as
// login — it issues a session in the same transaction as the user +
// credential rows, exactly like Login does.

func TestAuthUsecase_Register_Success(t *testing.T) {
	uc, users, userAccounts, refreshTokens, uow, hasher, accessTokens, refreshTokenGen, _, _ := newAuthUsecase(t)

	users.EXPECT().GetByEmail(mock.Anything, "budi@example.com").Return(nil, errs.ErrNotFound)
	hasher.EXPECT().Hash("s3cur3-password").Return("hashed-password", nil)

	created := &entity.User{ID: uuid.New(), Name: "Budi", Email: "budi@example.com"}
	accessExpiresAt := time.Now().Add(15 * time.Minute)
	wantRefreshExpiresAt := time.Now().Add(testRefreshTokenTTL)

	uow.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(usecase.AuthRepositories) error) error {
			return fn(usecase.AuthRepositories{User: users, UserAccount: userAccounts, RefreshToken: refreshTokens})
		})
	users.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.Name == "Budi" && u.Email == "budi@example.com" && u.Timezone == "UTC" && u.Username == nil
		})).
		Return(created, nil)
	userAccounts.EXPECT().
		CreateCredential(mock.Anything, created.ID, created.ID.String(), "hashed-password").
		Return(&entity.UserAccount{}, nil)
	accessTokens.EXPECT().Generate(created.ID).Return("access-token", accessExpiresAt, nil)
	refreshTokenGen.EXPECT().Generate().Return("raw-refresh-token", nil)
	refreshTokenGen.EXPECT().Hash("raw-refresh-token").Return("hashed-refresh-token")
	refreshTokens.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(rt *entity.RefreshToken) bool {
			return rt.UserID == created.ID && rt.TokenHash == "hashed-refresh-token" && rt.FamilyID != uuid.Nil
		})).
		Return(&entity.RefreshToken{ID: uuid.New()}, nil)

	result, err := uc.Register(context.Background(), usecase.RegisterInput{
		Name:     "Budi",
		Email:    "budi@example.com",
		Password: "s3cur3-password",
	})

	require.NoError(t, err)
	assert.Equal(t, created, result.User)
	assert.Equal(t, "access-token", result.AccessToken)
	assert.Equal(t, accessExpiresAt, result.AccessTokenExpiresAt)
	assert.Equal(t, "raw-refresh-token", result.RefreshToken)
	assert.WithinDuration(t, wantRefreshExpiresAt, result.RefreshTokenExpiresAt, time.Second)
}

func TestAuthUsecase_Register_EmailAlreadyExists(t *testing.T) {
	uc, users, _, _, _, _, _, _, _, _ := newAuthUsecase(t)

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
	uc, users, _, _, _, hasher, _, _, _, _ := newAuthUsecase(t)

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
	uc, users, userAccounts, _, uow, hasher, _, _, _, _ := newAuthUsecase(t)

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

func TestAuthUsecase_Register_IssueSessionFails(t *testing.T) {
	uc, users, userAccounts, refreshTokens, uow, hasher, accessTokens, _, _, _ := newAuthUsecase(t)

	users.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(nil, errs.ErrNotFound)
	hasher.EXPECT().Hash(mock.Anything).Return("hashed", nil)

	created := &entity.User{ID: uuid.New(), Name: "Budi", Email: "budi@example.com"}
	uow.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(usecase.AuthRepositories) error) error {
			return fn(usecase.AuthRepositories{User: users, UserAccount: userAccounts, RefreshToken: refreshTokens})
		})
	users.EXPECT().Create(mock.Anything, mock.Anything).Return(created, nil)
	userAccounts.EXPECT().
		CreateCredential(mock.Anything, created.ID, created.ID.String(), "hashed").
		Return(&entity.UserAccount{}, nil)
	wantErr := errors.New("token signing exploded")
	accessTokens.EXPECT().Generate(created.ID).Return("", time.Time{}, wantErr)
	// refreshTokenGen/refreshTokens.Create deliberately not stubbed — a
	// failed access token mint must short-circuit before minting/storing
	// a refresh token too, and the whole transaction rolls back.

	result, err := uc.Register(context.Background(), usecase.RegisterInput{
		Name: "Budi", Email: "budi@example.com", Password: "whatever",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}

// --- Login ---

func TestAuthUsecase_Login_Success(t *testing.T) {
	uc, users, userAccounts, refreshTokens, _, hasher, accessTokens, refreshTokenGen, _, _ := newAuthUsecase(t)

	user := &entity.User{ID: uuid.New(), Email: "budi@example.com"}
	passwordHash := "scrypt-hash"
	accessExpiresAt := time.Now().Add(15 * time.Minute)
	wantRefreshExpiresAt := time.Now().Add(testRefreshTokenTTL)

	users.EXPECT().GetByEmail(mock.Anything, "budi@example.com").Return(user, nil)
	userAccounts.EXPECT().GetCredentialByUserID(mock.Anything, user.ID).
		Return(&entity.UserAccount{PasswordHash: &passwordHash}, nil)
	hasher.EXPECT().Compare(passwordHash, "correct-password").Return(nil)
	accessTokens.EXPECT().Generate(user.ID).Return("access-token", accessExpiresAt, nil)
	refreshTokenGen.EXPECT().Generate().Return("raw-refresh-token", nil)
	refreshTokenGen.EXPECT().Hash("raw-refresh-token").Return("hashed-refresh-token")
	refreshTokens.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(rt *entity.RefreshToken) bool {
			// FamilyID is generated fresh (uuid.New()) inside Login for
			// every new session, so the test can only assert it's set,
			// not its exact value.
			return rt.UserID == user.ID &&
				rt.TokenHash == "hashed-refresh-token" &&
				rt.FamilyID != uuid.Nil
		})).
		Return(&entity.RefreshToken{ID: uuid.New()}, nil)

	result, err := uc.Login(context.Background(), usecase.LoginInput{
		Email: "budi@example.com", Password: "correct-password",
	})

	require.NoError(t, err)
	assert.Equal(t, user, result.User)
	assert.Equal(t, "access-token", result.AccessToken)
	assert.Equal(t, accessExpiresAt, result.AccessTokenExpiresAt)
	assert.Equal(t, "raw-refresh-token", result.RefreshToken)
	assert.WithinDuration(t, wantRefreshExpiresAt, result.RefreshTokenExpiresAt, time.Second)
}

func TestAuthUsecase_Login_EmailNotFound(t *testing.T) {
	uc, users, _, _, _, _, _, _, _, _ := newAuthUsecase(t)

	users.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(nil, errs.ErrNotFound)

	result, err := uc.Login(context.Background(), usecase.LoginInput{
		Email: "nobody@example.com", Password: "whatever",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrInvalidCredentials,
		"must NOT leak errs.ErrNotFound here — that would let an attacker enumerate registered emails")
}

func TestAuthUsecase_Login_NoCredentialAccount(t *testing.T) {
	uc, users, userAccounts, _, _, _, _, _, _, _ := newAuthUsecase(t)

	user := &entity.User{ID: uuid.New()}
	users.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(user, nil)
	userAccounts.EXPECT().GetCredentialByUserID(mock.Anything, user.ID).Return(nil, errs.ErrNotFound)

	_, err := uc.Login(context.Background(), usecase.LoginInput{Email: "x@example.com", Password: "y"})

	assert.ErrorIs(t, err, errs.ErrInvalidCredentials)
}

func TestAuthUsecase_Login_WrongPassword(t *testing.T) {
	uc, users, userAccounts, _, _, hasher, _, _, _, _ := newAuthUsecase(t)

	user := &entity.User{ID: uuid.New()}
	hash := "some-hash"
	users.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(user, nil)
	userAccounts.EXPECT().GetCredentialByUserID(mock.Anything, user.ID).
		Return(&entity.UserAccount{PasswordHash: &hash}, nil)
	hasher.EXPECT().Compare(hash, "wrong-password").Return(errors.New("mismatch"))
	// Below the lockout threshold — only the counter increments, no lock.
	userAccounts.EXPECT().IncrementFailedLoginAttempts(mock.Anything, user.ID).
		Return(&entity.UserAccount{FailedLoginAttempts: 1}, nil)

	_, err := uc.Login(context.Background(), usecase.LoginInput{Email: "x@example.com", Password: "wrong-password"})

	assert.ErrorIs(t, err, errs.ErrInvalidCredentials)
}

func TestAuthUsecase_Login_WrongPassword_LocksAccountAtThreshold(t *testing.T) {
	uc, users, userAccounts, _, _, hasher, _, _, _, _ := newAuthUsecase(t)

	user := &entity.User{ID: uuid.New()}
	hash := "some-hash"
	users.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(user, nil)
	userAccounts.EXPECT().GetCredentialByUserID(mock.Anything, user.ID).
		Return(&entity.UserAccount{PasswordHash: &hash}, nil)
	hasher.EXPECT().Compare(hash, "wrong-password").Return(errors.New("mismatch"))
	userAccounts.EXPECT().IncrementFailedLoginAttempts(mock.Anything, user.ID).
		Return(&entity.UserAccount{FailedLoginAttempts: testMaxFailedLoginAttempts}, nil)
	userAccounts.EXPECT().
		LockCredentialAccount(mock.Anything, user.ID, mock.MatchedBy(func(until time.Time) bool {
			return until.After(time.Now())
		})).
		Return(nil)

	_, err := uc.Login(context.Background(), usecase.LoginInput{Email: "x@example.com", Password: "wrong-password"})

	assert.ErrorIs(t, err, errs.ErrInvalidCredentials)
}

func TestAuthUsecase_Login_AccountLocked_RejectsBeforePasswordCheck(t *testing.T) {
	uc, users, userAccounts, _, _, hasher, _, _, _, _ := newAuthUsecase(t)

	user := &entity.User{ID: uuid.New()}
	hash := "some-hash"
	lockedUntil := time.Now().Add(5 * time.Minute)
	users.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(user, nil)
	userAccounts.EXPECT().GetCredentialByUserID(mock.Anything, user.ID).
		Return(&entity.UserAccount{PasswordHash: &hash, LockedUntil: &lockedUntil}, nil)
	// hasher.Compare deliberately not stubbed — a locked account must be
	// rejected before ever touching the password.
	_ = hasher

	_, err := uc.Login(context.Background(), usecase.LoginInput{Email: "x@example.com", Password: "whatever"})

	assert.ErrorIs(t, err, errs.ErrAccountLocked)
}

func TestAuthUsecase_Login_LockExpired_AllowsPasswordCheck(t *testing.T) {
	uc, users, userAccounts, refreshTokens, _, hasher, accessTokens, refreshTokenGen, _, _ := newAuthUsecase(t)

	user := &entity.User{ID: uuid.New()}
	passwordHash := "scrypt-hash"
	expiredLock := time.Now().Add(-time.Minute)
	users.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(user, nil)
	userAccounts.EXPECT().GetCredentialByUserID(mock.Anything, user.ID).
		Return(&entity.UserAccount{PasswordHash: &passwordHash, FailedLoginAttempts: testMaxFailedLoginAttempts, LockedUntil: &expiredLock}, nil)
	hasher.EXPECT().Compare(passwordHash, "correct-password").Return(nil)
	userAccounts.EXPECT().ResetFailedLoginAttempts(mock.Anything, user.ID).Return(nil)
	accessTokens.EXPECT().Generate(user.ID).Return("access-token", time.Now().Add(15*time.Minute), nil)
	refreshTokenGen.EXPECT().Generate().Return("raw-refresh-token", nil)
	refreshTokenGen.EXPECT().Hash("raw-refresh-token").Return("hashed-refresh-token")
	refreshTokens.EXPECT().Create(mock.Anything, mock.Anything).Return(&entity.RefreshToken{ID: uuid.New()}, nil)

	_, err := uc.Login(context.Background(), usecase.LoginInput{Email: "x@example.com", Password: "correct-password"})

	assert.NoError(t, err)
}

func TestAuthUsecase_Login_NilPasswordHash_OAuthOnlyAccount(t *testing.T) {
	uc, users, userAccounts, _, _, _, _, _, _, _ := newAuthUsecase(t)

	user := &entity.User{ID: uuid.New()}
	users.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(user, nil)
	userAccounts.EXPECT().GetCredentialByUserID(mock.Anything, user.ID).
		Return(&entity.UserAccount{PasswordHash: nil}, nil)
	// hasher.Compare deliberately not stubbed — must never be called
	// against a nil hash.

	_, err := uc.Login(context.Background(), usecase.LoginInput{Email: "x@example.com", Password: "y"})

	assert.ErrorIs(t, err, errs.ErrInvalidCredentials)
}

// --- Refresh ---

func TestAuthUsecase_Refresh_Success_RotatesToken(t *testing.T) {
	uc, users, _, refreshTokens, uow, _, accessTokens, refreshTokenGen, _, _ := newAuthUsecase(t)

	user := &entity.User{ID: uuid.New()}
	familyID := uuid.New()
	existing := &entity.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	refreshTokenGen.EXPECT().Hash("old-raw-token").Return("old-hash")
	refreshTokens.EXPECT().GetByTokenHash(mock.Anything, "old-hash").Return(existing, nil)
	users.EXPECT().GetByID(mock.Anything, user.ID).Return(user, nil)

	newAccessExpiresAt := time.Now().Add(15 * time.Minute)
	newTokenRow := &entity.RefreshToken{ID: uuid.New()}

	uow.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(usecase.AuthRepositories) error) error {
			return fn(usecase.AuthRepositories{RefreshToken: refreshTokens})
		})
	accessTokens.EXPECT().Generate(user.ID).Return("new-access-token", newAccessExpiresAt, nil)
	refreshTokenGen.EXPECT().Generate().Return("new-raw-token", nil)
	refreshTokenGen.EXPECT().Hash("new-raw-token").Return("new-hash")
	refreshTokens.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(rt *entity.RefreshToken) bool {
			// The rotated token must stay in the SAME family as the one
			// it replaces — that's what lets RevokeFamily catch reuse
			// later even after several rotations.
			return rt.UserID == user.ID && rt.FamilyID == familyID && rt.TokenHash == "new-hash"
		})).
		Return(newTokenRow, nil)
	refreshTokens.EXPECT().Rotate(mock.Anything, existing.ID, newTokenRow.ID).Return(nil)

	result, err := uc.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: "old-raw-token"})

	require.NoError(t, err)
	assert.Equal(t, user, result.User)
	assert.Equal(t, "new-access-token", result.AccessToken)
	assert.Equal(t, "new-raw-token", result.RefreshToken)
}

func TestAuthUsecase_Refresh_TokenNotFound(t *testing.T) {
	uc, _, _, refreshTokens, _, _, _, refreshTokenGen, _, _ := newAuthUsecase(t)

	refreshTokenGen.EXPECT().Hash("unknown-token").Return("unknown-hash")
	refreshTokens.EXPECT().GetByTokenHash(mock.Anything, "unknown-hash").Return(nil, errs.ErrNotFound)

	_, err := uc.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: "unknown-token"})

	assert.ErrorIs(t, err, errs.ErrUnauthorized)
}

func TestAuthUsecase_Refresh_Expired(t *testing.T) {
	uc, users, _, refreshTokens, _, _, _, refreshTokenGen, _, _ := newAuthUsecase(t)

	existing := &entity.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		FamilyID:  uuid.New(),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	refreshTokenGen.EXPECT().Hash("expired-token").Return("expired-hash")
	refreshTokens.EXPECT().GetByTokenHash(mock.Anything, "expired-hash").Return(existing, nil)
	// users.GetByID deliberately not stubbed — an expired token should
	// short-circuit before ever looking up the user.
	_ = users

	_, err := uc.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: "expired-token"})

	assert.ErrorIs(t, err, errs.ErrTokenExpired)
}

func TestAuthUsecase_Refresh_ReuseDetected_RevokesEntireFamily(t *testing.T) {
	uc, _, _, refreshTokens, _, _, _, refreshTokenGen, _, _ := newAuthUsecase(t)

	familyID := uuid.New()
	revokedAt := time.Now().Add(-1 * time.Minute)
	existing := &entity.RefreshToken{
		ID:        uuid.New(),
		FamilyID:  familyID,
		RevokedAt: &revokedAt, // already rotated out once before
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	refreshTokenGen.EXPECT().Hash("stolen-token").Return("stolen-hash")
	refreshTokens.EXPECT().GetByTokenHash(mock.Anything, "stolen-hash").Return(existing, nil)
	refreshTokens.EXPECT().RevokeFamily(mock.Anything, familyID).Return(nil)
	// users.GetByID and uow.WithTransaction deliberately not stubbed —
	// a reused token must be rejected before ever minting new tokens.

	_, err := uc.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: "stolen-token"})

	assert.ErrorIs(t, err, errs.ErrUnauthorized)
}

func TestAuthUsecase_Refresh_UserNotFound(t *testing.T) {
	uc, users, _, refreshTokens, _, _, _, refreshTokenGen, _, _ := newAuthUsecase(t)

	existing := &entity.RefreshToken{
		ID: uuid.New(), UserID: uuid.New(), FamilyID: uuid.New(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	refreshTokenGen.EXPECT().Hash("orphaned-token").Return("orphaned-hash")
	refreshTokens.EXPECT().GetByTokenHash(mock.Anything, "orphaned-hash").Return(existing, nil)
	users.EXPECT().GetByID(mock.Anything, existing.UserID).Return(nil, errs.ErrNotFound)

	_, err := uc.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: "orphaned-token"})

	assert.ErrorIs(t, err, errs.ErrUnauthorized)
}

func TestAuthUsecase_Refresh_ConflictOnRotate_TreatedAsUnauthorized(t *testing.T) {
	uc, users, _, refreshTokens, uow, _, accessTokens, refreshTokenGen, _, _ := newAuthUsecase(t)

	user := &entity.User{ID: uuid.New()}
	existing := &entity.RefreshToken{
		ID: uuid.New(), UserID: user.ID, FamilyID: uuid.New(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	refreshTokenGen.EXPECT().Hash("raced-token").Return("raced-hash")
	refreshTokens.EXPECT().GetByTokenHash(mock.Anything, "raced-hash").Return(existing, nil)
	users.EXPECT().GetByID(mock.Anything, user.ID).Return(user, nil)

	newTokenRow := &entity.RefreshToken{ID: uuid.New()}
	uow.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(usecase.AuthRepositories) error) error {
			return fn(usecase.AuthRepositories{RefreshToken: refreshTokens})
		})
	accessTokens.EXPECT().Generate(user.ID).Return("new-access-token", time.Now().Add(15*time.Minute), nil)
	refreshTokenGen.EXPECT().Generate().Return("new-raw-token", nil)
	refreshTokenGen.EXPECT().Hash("new-raw-token").Return("new-hash")
	refreshTokens.EXPECT().Create(mock.Anything, mock.Anything).Return(newTokenRow, nil)
	// Simulates another concurrent Refresh request winning the race and
	// rotating this same token first.
	refreshTokens.EXPECT().Rotate(mock.Anything, existing.ID, newTokenRow.ID).Return(errs.ErrConflict)

	_, err := uc.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: "raced-token"})

	assert.ErrorIs(t, err, errs.ErrUnauthorized)
}

// --- Logout ---

func TestAuthUsecase_Logout_Success(t *testing.T) {
	uc, _, _, refreshTokens, _, _, _, refreshTokenGen, _, _ := newAuthUsecase(t)

	row := &entity.RefreshToken{ID: uuid.New()}
	refreshTokenGen.EXPECT().Hash("raw-token").Return("hashed-token")
	refreshTokens.EXPECT().GetByTokenHash(mock.Anything, "hashed-token").Return(row, nil)
	refreshTokens.EXPECT().Revoke(mock.Anything, row.ID).Return(nil)

	err := uc.Logout(context.Background(), "raw-token")

	assert.NoError(t, err)
}

func TestAuthUsecase_Logout_TokenNotFound_IsIdempotent(t *testing.T) {
	uc, _, _, refreshTokens, _, _, _, refreshTokenGen, _, _ := newAuthUsecase(t)

	refreshTokenGen.EXPECT().Hash("already-gone").Return("gone-hash")
	refreshTokens.EXPECT().GetByTokenHash(mock.Anything, "gone-hash").Return(nil, errs.ErrNotFound)
	// Revoke deliberately not stubbed — logging out a token that's
	// already gone should be a no-op, not an error.

	err := uc.Logout(context.Background(), "already-gone")

	assert.NoError(t, err)
}

func TestAuthUsecase_Logout_AlreadyRevoked_IsIdempotent(t *testing.T) {
	uc, _, _, refreshTokens, _, _, _, refreshTokenGen, _, _ := newAuthUsecase(t)

	revokedAt := time.Now().Add(-time.Minute)
	row := &entity.RefreshToken{ID: uuid.New(), RevokedAt: &revokedAt}
	refreshTokenGen.EXPECT().Hash("already-revoked").Return("revoked-hash")
	refreshTokens.EXPECT().GetByTokenHash(mock.Anything, "revoked-hash").Return(row, nil)
	// Revoke deliberately not stubbed — calling Revoke again on an
	// already-revoked row would be redundant.

	err := uc.Logout(context.Background(), "already-revoked")

	assert.NoError(t, err)
}

func TestAuthUsecase_Logout_RevokeError(t *testing.T) {
	uc, _, _, refreshTokens, _, _, _, refreshTokenGen, _, _ := newAuthUsecase(t)

	row := &entity.RefreshToken{ID: uuid.New()}
	wantErr := errors.New("db exploded")
	refreshTokenGen.EXPECT().Hash(mock.Anything).Return("hashed-token")
	refreshTokens.EXPECT().GetByTokenHash(mock.Anything, mock.Anything).Return(row, nil)
	refreshTokens.EXPECT().Revoke(mock.Anything, row.ID).Return(wantErr)

	err := uc.Logout(context.Background(), "raw-token")

	assert.ErrorIs(t, err, wantErr)
}

// --- ForgotPassword ---

func TestAuthUsecase_ForgotPassword_UnknownEmail_SilentSuccess(t *testing.T) {
	uc, users, _, _, _, _, _, _, verifications, mailer := newAuthUsecase(t)

	users.EXPECT().GetByEmail(mock.Anything, "ghost@example.com").Return(nil, errs.ErrNotFound)
	// Nothing else must be reached — no expectations set on verifications
	// or mailer at all, so mockery's t.Cleanup assertion fails the test
	// if either is called, which is exactly the "never leak whether the
	// email exists" property this is testing.

	err := uc.ForgotPassword(context.Background(), "ghost@example.com")

	assert.NoError(t, err)
	verifications.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mailer.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthUsecase_ForgotPassword_KnownEmail_SendsResetLink(t *testing.T) {
	uc, users, _, _, _, _, _, refreshTokenGen, verifications, mailer := newAuthUsecase(t)

	user := &entity.User{ID: uuid.New(), Name: "Budi", Email: "budi@example.com"}
	users.EXPECT().GetByEmail(mock.Anything, "budi@example.com").Return(user, nil)
	verifications.EXPECT().DeleteByIdentifier(mock.Anything, "password_reset:budi@example.com").Return(nil)
	refreshTokenGen.EXPECT().Generate().Return("raw-reset-token", nil)
	refreshTokenGen.EXPECT().Hash("raw-reset-token").Return("hashed-reset-token")
	verifications.EXPECT().
		Create(mock.Anything, "password_reset:budi@example.com", "hashed-reset-token", mock.Anything).
		Return(&entity.UserVerification{ID: uuid.New()}, nil)
	mailer.EXPECT().
		Send(mock.Anything, "budi@example.com", mock.Anything, mock.MatchedBy(func(body string) bool {
			return strings.Contains(body, "raw-reset-token") && strings.Contains(body, testWebBaseURL)
		})).
		Return(nil)

	err := uc.ForgotPassword(context.Background(), "budi@example.com")

	assert.NoError(t, err)
}

// --- ResetPassword ---

func TestAuthUsecase_ResetPassword_InvalidOrExpiredToken(t *testing.T) {
	uc, _, _, _, _, _, _, refreshTokenGen, verifications, _ := newAuthUsecase(t)

	refreshTokenGen.EXPECT().Hash("bad-token").Return("hashed-bad-token")
	verifications.EXPECT().
		GetValid(mock.Anything, "password_reset:budi@example.com", "hashed-bad-token").
		Return(nil, errs.ErrNotFound)

	err := uc.ResetPassword(context.Background(), usecase.ResetPasswordInput{
		Email: "budi@example.com", Token: "bad-token", NewPassword: "a-new-password",
	})

	assert.ErrorIs(t, err, errs.ErrInvalidToken)
}

func TestAuthUsecase_ResetPassword_Success(t *testing.T) {
	uc, users, userAccounts, refreshTokens, _, hasher, _, refreshTokenGen, verifications, _ := newAuthUsecase(t)

	userID := uuid.New()
	user := &entity.User{ID: userID, Email: "budi@example.com"}
	verification := &entity.UserVerification{ID: uuid.New()}

	refreshTokenGen.EXPECT().Hash("good-token").Return("hashed-good-token")
	verifications.EXPECT().
		GetValid(mock.Anything, "password_reset:budi@example.com", "hashed-good-token").
		Return(verification, nil)
	users.EXPECT().GetByEmail(mock.Anything, "budi@example.com").Return(user, nil)
	hasher.EXPECT().Hash("a-new-password").Return("hashed-new-password", nil)
	userAccounts.EXPECT().UpdatePasswordHash(mock.Anything, userID, "hashed-new-password").Return(nil)
	verifications.EXPECT().Delete(mock.Anything, verification.ID).Return(nil)
	userAccounts.EXPECT().ResetFailedLoginAttempts(mock.Anything, userID).Return(nil)
	refreshTokens.EXPECT().RevokeAllByUserID(mock.Anything, userID).Return(nil)

	err := uc.ResetPassword(context.Background(), usecase.ResetPasswordInput{
		Email: "budi@example.com", Token: "good-token", NewPassword: "a-new-password",
	})

	assert.NoError(t, err)
}
