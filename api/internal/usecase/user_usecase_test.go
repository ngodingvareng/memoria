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

func newUserUsecase(t *testing.T) (usecase.UserUsecase, *mocks.MockUserRepository) {
	users := mocks.NewMockUserRepository(t)
	return usecase.NewUserUsecase(users), users
}

func TestUserUsecase_CheckUsernameAvailability_Available(t *testing.T) {
	uc, users := newUserUsecase(t)

	users.EXPECT().GetByUsername(mock.Anything, "budisantoso").Return(nil, errs.ErrNotFound)

	available, err := uc.CheckUsernameAvailability(context.Background(), "budisantoso")

	require.NoError(t, err)
	assert.True(t, available)
}

func TestUserUsecase_CheckUsernameAvailability_Taken(t *testing.T) {
	uc, users := newUserUsecase(t)

	users.EXPECT().GetByUsername(mock.Anything, "budisantoso").
		Return(&entity.User{ID: uuid.New(), Username: strPtr("budisantoso")}, nil)

	available, err := uc.CheckUsernameAvailability(context.Background(), "budisantoso")

	require.NoError(t, err)
	assert.False(t, available)
}

func TestUserUsecase_CheckUsernameAvailability_RepositoryError(t *testing.T) {
	uc, users := newUserUsecase(t)

	wantErr := errors.New("db exploded")
	users.EXPECT().GetByUsername(mock.Anything, "budisantoso").Return(nil, wantErr)

	_, err := uc.CheckUsernameAvailability(context.Background(), "budisantoso")

	assert.ErrorIs(t, err, wantErr)
}

func TestUserUsecase_SetUsername_Success(t *testing.T) {
	uc, users := newUserUsecase(t)

	userID := uuid.New()
	updated := &entity.User{ID: userID, Username: strPtr("budisantoso")}
	users.EXPECT().SetUsername(mock.Anything, userID, "budisantoso").Return(updated, nil)

	result, err := uc.SetUsername(context.Background(), userID, "budisantoso")

	require.NoError(t, err)
	assert.Equal(t, updated, result)
}

func TestUserUsecase_SetUsername_AlreadyTaken(t *testing.T) {
	uc, users := newUserUsecase(t)

	userID := uuid.New()
	users.EXPECT().SetUsername(mock.Anything, userID, "budisantoso").
		Return(nil, errs.ErrUsernameAlreadyExists)

	_, err := uc.SetUsername(context.Background(), userID, "budisantoso")

	assert.ErrorIs(t, err, errs.ErrUsernameAlreadyExists)
}
