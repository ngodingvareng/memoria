//go:build unit

package usecase_test

import (
	"context"
	"errors"
	"strings"
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

func newUserUsecase(t *testing.T) (usecase.UserUsecase, *mocks.MockUserRepository, *mocks.MockUserKnownRepository, *mocks.MockProfileImageStorage) {
	users := mocks.NewMockUserRepository(t)
	knowns := mocks.NewMockUserKnownRepository(t)
	storage := mocks.NewMockProfileImageStorage(t)
	return usecase.NewUserUsecase(users, knowns, storage), users, knowns, storage
}

func TestUserUsecase_CheckUsernameAvailability_Available(t *testing.T) {
	uc, users, _, _ := newUserUsecase(t)

	users.EXPECT().GetByUsername(mock.Anything, "budisantoso").Return(nil, errs.ErrNotFound)

	available, err := uc.CheckUsernameAvailability(context.Background(), "budisantoso")

	require.NoError(t, err)
	assert.True(t, available)
}

func TestUserUsecase_CheckUsernameAvailability_Taken(t *testing.T) {
	uc, users, _, _ := newUserUsecase(t)

	users.EXPECT().GetByUsername(mock.Anything, "budisantoso").
		Return(&entity.User{ID: uuid.New(), Username: strPtr("budisantoso")}, nil)

	available, err := uc.CheckUsernameAvailability(context.Background(), "budisantoso")

	require.NoError(t, err)
	assert.False(t, available)
}

func TestUserUsecase_CheckUsernameAvailability_RepositoryError(t *testing.T) {
	uc, users, _, _ := newUserUsecase(t)

	wantErr := errors.New("db exploded")
	users.EXPECT().GetByUsername(mock.Anything, "budisantoso").Return(nil, wantErr)

	_, err := uc.CheckUsernameAvailability(context.Background(), "budisantoso")

	assert.ErrorIs(t, err, wantErr)
}

func TestUserUsecase_SetUsername_Success(t *testing.T) {
	uc, users, _, _ := newUserUsecase(t)

	userID := uuid.New()
	updated := &entity.User{ID: userID, Username: strPtr("budisantoso")}
	users.EXPECT().SetUsername(mock.Anything, userID, "budisantoso").Return(updated, nil)

	result, err := uc.SetUsername(context.Background(), userID, "budisantoso")

	require.NoError(t, err)
	assert.Equal(t, updated, result)
}

func TestUserUsecase_SetUsername_AlreadyTaken(t *testing.T) {
	uc, users, _, _ := newUserUsecase(t)

	userID := uuid.New()
	users.EXPECT().SetUsername(mock.Anything, userID, "budisantoso").
		Return(nil, errs.ErrUsernameAlreadyExists)

	_, err := uc.SetUsername(context.Background(), userID, "budisantoso")

	assert.ErrorIs(t, err, errs.ErrUsernameAlreadyExists)
}

func TestUserUsecase_GetPublicProfileByUsername_Success(t *testing.T) {
	uc, users, _, _ := newUserUsecase(t)

	found := &entity.User{ID: uuid.New(), Name: "Gede", Username: strPtr("gede")}
	users.EXPECT().GetByUsername(mock.Anything, "gede").Return(found, nil)

	result, err := uc.GetPublicProfileByUsername(context.Background(), "gede")

	require.NoError(t, err)
	assert.Equal(t, found, result)
}

func TestUserUsecase_GetPublicProfileByUsername_NotFound(t *testing.T) {
	uc, users, _, _ := newUserUsecase(t)

	users.EXPECT().GetByUsername(mock.Anything, "ghost").Return(nil, errs.ErrNotFound)

	_, err := uc.GetPublicProfileByUsername(context.Background(), "ghost")

	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestUserUsecase_MarkUserKnown_Success(t *testing.T) {
	uc, users, knowns, _ := newUserUsecase(t)

	knowerID := uuid.New()
	target := &entity.User{ID: uuid.New(), Username: strPtr("gede")}
	users.EXPECT().GetByUsername(mock.Anything, "gede").Return(target, nil)
	knowns.EXPECT().MarkKnown(mock.Anything, knowerID, target.ID).Return(nil)

	err := uc.MarkUserKnown(context.Background(), knowerID, "gede")

	assert.NoError(t, err)
}

func TestUserUsecase_MarkUserKnown_UsernameNotFound(t *testing.T) {
	uc, users, _, _ := newUserUsecase(t)

	users.EXPECT().GetByUsername(mock.Anything, "ghost").Return(nil, errs.ErrNotFound)

	// knowns.MarkKnown must never be reached — no expectation set on it
	// at all, so mockery's t.Cleanup assertion fails the test if it is.
	err := uc.MarkUserKnown(context.Background(), uuid.New(), "ghost")

	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestUserUsecase_UploadProfileImage_NoExistingImage_Success(t *testing.T) {
	uc, users, _, storage := newUserUsecase(t)

	userID := uuid.New()
	body := strings.NewReader("fake image bytes")
	current := &entity.User{ID: userID, ImagePath: nil}
	updated := &entity.User{ID: userID, ImagePath: strPtr("https://cdn.example/users/x/y.jpg")}

	users.EXPECT().GetByID(mock.Anything, userID).Return(current, nil)
	storage.EXPECT().
		Put(mock.Anything, mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, "users/"+userID.String()+"/") && strings.HasSuffix(key, ".jpg")
		}), body, int64(17), "image/jpeg").
		Return(nil)
	storage.EXPECT().PublicURL(mock.Anything).Return("https://cdn.example/users/x/y.jpg")
	users.EXPECT().UpdateImagePath(mock.Anything, userID, updated.ImagePath).Return(updated, nil)
	// storage.Delete must never be reached — no prior image to clean up.

	result, err := uc.UploadProfileImage(context.Background(), usecase.UploadProfileImageInput{
		UserID:      userID,
		FileName:    "avatar.jpg",
		ContentType: "image/jpeg",
		Size:        17,
		Body:        body,
	})

	require.NoError(t, err)
	assert.Equal(t, updated, result)
}

func TestUserUsecase_UploadProfileImage_ReplacesExistingImage(t *testing.T) {
	uc, users, _, storage := newUserUsecase(t)

	userID := uuid.New()
	body := strings.NewReader("fake image bytes")
	oldURL := "https://cdn.example/users/x/old.jpg"
	current := &entity.User{ID: userID, ImagePath: &oldURL}
	newURL := "https://cdn.example/users/x/new.jpg"
	updated := &entity.User{ID: userID, ImagePath: &newURL}

	nonEmptyKey := mock.MatchedBy(func(k string) bool { return k != "" })

	users.EXPECT().GetByID(mock.Anything, userID).Return(current, nil)
	storage.EXPECT().Put(mock.Anything, mock.Anything, body, int64(17), "image/jpeg").Return(nil)
	storage.EXPECT().PublicURL(nonEmptyKey).Return(newURL)
	users.EXPECT().UpdateImagePath(mock.Anything, userID, &newURL).Return(updated, nil)
	storage.EXPECT().PublicURL("").Return("https://cdn.example/")
	storage.EXPECT().Delete(mock.Anything, "users/x/old.jpg").Return(nil)

	result, err := uc.UploadProfileImage(context.Background(), usecase.UploadProfileImageInput{
		UserID:      userID,
		FileName:    "avatar.jpg",
		ContentType: "image/jpeg",
		Size:        17,
		Body:        body,
	})

	require.NoError(t, err)
	assert.Equal(t, updated, result)
}

func TestUserUsecase_UploadProfileImage_DBSaveFails_CleansUpUpload(t *testing.T) {
	uc, users, _, storage := newUserUsecase(t)

	userID := uuid.New()
	body := strings.NewReader("fake image bytes")
	current := &entity.User{ID: userID, ImagePath: nil}
	dbErr := errors.New("db exploded")

	users.EXPECT().GetByID(mock.Anything, userID).Return(current, nil)
	storage.EXPECT().Put(mock.Anything, mock.Anything, body, int64(17), "image/jpeg").Return(nil)
	storage.EXPECT().PublicURL(mock.Anything).Return("https://cdn.example/users/x/y.jpg")
	users.EXPECT().UpdateImagePath(mock.Anything, userID, mock.Anything).Return(nil, dbErr)
	storage.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)

	_, err := uc.UploadProfileImage(context.Background(), usecase.UploadProfileImageInput{
		UserID:      userID,
		FileName:    "avatar.jpg",
		ContentType: "image/jpeg",
		Size:        17,
		Body:        body,
	})

	assert.ErrorIs(t, err, dbErr)
}
