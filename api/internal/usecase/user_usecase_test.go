//go:build unit

package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/usecase/mocks"
)

func newUserUsecase(t *testing.T) (usecase.UserUsecase, *mocks.MockUserRepository, *mocks.MockUserKnownRepository) {
	users := mocks.NewMockUserRepository(t)
	knowns := mocks.NewMockUserKnownRepository(t)
	return usecase.NewUserUsecase(users, knowns), users, knowns
}

func TestUserUsecase_CheckUsernameAvailability_Available(t *testing.T) {
	uc, users, _ := newUserUsecase(t)

	users.EXPECT().GetByUsername(mock.Anything, "budisantoso").Return(nil, errs.ErrNotFound)

	available, err := uc.CheckUsernameAvailability(context.Background(), "budisantoso")

	require.NoError(t, err)
	assert.True(t, available)
}

func TestUserUsecase_CheckUsernameAvailability_Taken(t *testing.T) {
	uc, users, _ := newUserUsecase(t)

	users.EXPECT().GetByUsername(mock.Anything, "budisantoso").
		Return(&entity.User{ID: uuid.New(), Username: strPtr("budisantoso")}, nil)

	available, err := uc.CheckUsernameAvailability(context.Background(), "budisantoso")

	require.NoError(t, err)
	assert.False(t, available)
}

func TestUserUsecase_CheckUsernameAvailability_RepositoryError(t *testing.T) {
	uc, users, _ := newUserUsecase(t)

	wantErr := errors.New("db exploded")
	users.EXPECT().GetByUsername(mock.Anything, "budisantoso").Return(nil, wantErr)

	_, err := uc.CheckUsernameAvailability(context.Background(), "budisantoso")

	assert.ErrorIs(t, err, wantErr)
}

func TestUserUsecase_SetUsername_Success(t *testing.T) {
	uc, users, _ := newUserUsecase(t)

	userID := uuid.New()
	updated := &entity.User{ID: userID, Username: strPtr("budisantoso")}
	users.EXPECT().SetUsername(mock.Anything, userID, "budisantoso").Return(updated, nil)

	result, err := uc.SetUsername(context.Background(), userID, "budisantoso")

	require.NoError(t, err)
	assert.Equal(t, updated, result)
}

func TestUserUsecase_SetUsername_AlreadyTaken(t *testing.T) {
	uc, users, _ := newUserUsecase(t)

	userID := uuid.New()
	users.EXPECT().SetUsername(mock.Anything, userID, "budisantoso").
		Return(nil, errs.ErrUsernameAlreadyExists)

	_, err := uc.SetUsername(context.Background(), userID, "budisantoso")

	assert.ErrorIs(t, err, errs.ErrUsernameAlreadyExists)
}

func TestUserUsecase_GetPublicProfileByUsername_Success(t *testing.T) {
	uc, users, _ := newUserUsecase(t)

	found := &entity.User{ID: uuid.New(), Name: "Gede", Username: strPtr("gede")}
	users.EXPECT().GetByUsername(mock.Anything, "gede").Return(found, nil)

	result, err := uc.GetPublicProfileByUsername(context.Background(), "gede")

	require.NoError(t, err)
	assert.Equal(t, found, result)
}

func TestUserUsecase_GetPublicProfileByUsername_NotFound(t *testing.T) {
	uc, users, _ := newUserUsecase(t)

	users.EXPECT().GetByUsername(mock.Anything, "ghost").Return(nil, errs.ErrNotFound)

	_, err := uc.GetPublicProfileByUsername(context.Background(), "ghost")

	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestUserUsecase_MarkUserKnown_Success(t *testing.T) {
	uc, users, knowns := newUserUsecase(t)

	knowerID := uuid.New()
	target := &entity.User{ID: uuid.New(), Username: strPtr("gede")}
	users.EXPECT().GetByUsername(mock.Anything, "gede").Return(target, nil)
	knowns.EXPECT().MarkKnown(mock.Anything, knowerID, target.ID).Return(nil)

	err := uc.MarkUserKnown(context.Background(), knowerID, "gede")

	assert.NoError(t, err)
}

func TestUserUsecase_MarkUserKnown_UsernameNotFound(t *testing.T) {
	uc, users, _ := newUserUsecase(t)

	users.EXPECT().GetByUsername(mock.Anything, "ghost").Return(nil, errs.ErrNotFound)

	// knowns.MarkKnown must never be reached — no expectation set on it
	// at all, so mockery's t.Cleanup assertion fails the test if it is.
	err := uc.MarkUserKnown(context.Background(), uuid.New(), "ghost")

	assert.ErrorIs(t, err, errs.ErrNotFound)
}
