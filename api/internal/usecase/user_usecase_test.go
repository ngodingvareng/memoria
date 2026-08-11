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
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/usecase/mocks"
)

func newUserUsecase(t *testing.T) (usecase.UserUsecase, *mocks.MockUserRepository, *mocks.MockUserKnownRepository, *mocks.MockProfileImageStorage) {
	uc, users, knowns, _, _, storage := newUserUsecaseWithPrivacy(t)
	return uc, users, knowns, storage
}

func newUserUsecaseWithPrivacy(t *testing.T) (usecase.UserUsecase, *mocks.MockUserRepository, *mocks.MockUserKnownRepository, *mocks.MockUserBlockRepository, *mocks.MockUserMuteRepository, *mocks.MockProfileImageStorage) {
	users := mocks.NewMockUserRepository(t)
	knowns := mocks.NewMockUserKnownRepository(t)
	blocks := mocks.NewMockUserBlockRepository(t)
	mutes := mocks.NewMockUserMuteRepository(t)
	storage := mocks.NewMockProfileImageStorage(t)
	return usecase.NewUserUsecase(users, knowns, blocks, mutes, storage), users, knowns, blocks, mutes, storage
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

func TestUserUsecase_UnmarkUserKnown_ResolvesUsernameThenUnmarks(t *testing.T) {
	uc, users, knowns, _ := newUserUsecase(t)

	knowerID := uuid.New()
	target := &entity.User{ID: uuid.New(), Username: strPtr("gede")}
	users.EXPECT().GetByUsername(mock.Anything, "gede").Return(target, nil)
	knowns.EXPECT().Unmark(mock.Anything, knowerID, target.ID).Return(nil)

	err := uc.UnmarkUserKnown(context.Background(), knowerID, "gede")

	assert.NoError(t, err)
}

func TestUserUsecase_UnmarkUserKnown_UsernameNotFound(t *testing.T) {
	uc, users, _, _ := newUserUsecase(t)

	users.EXPECT().GetByUsername(mock.Anything, "ghost").Return(nil, errs.ErrNotFound)

	// knowns.Unmark must never be reached — no expectation set on it at
	// all, so mockery's t.Cleanup assertion fails the test if it is.
	err := uc.UnmarkUserKnown(context.Background(), uuid.New(), "ghost")

	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestUserUsecase_ListKnownUsers_SkipsUnresolvableIDs(t *testing.T) {
	uc, users, knowns, _ := newUserUsecase(t)

	knowerID := uuid.New()
	staleID, activeID := uuid.New(), uuid.New()
	active := &entity.User{ID: activeID, Name: "Gede"}

	knowns.EXPECT().ListKnownUserIDs(mock.Anything, knowerID).Return([]uuid.UUID{staleID, activeID}, nil)
	users.EXPECT().GetByID(mock.Anything, staleID).Return(nil, errs.ErrNotFound)
	users.EXPECT().GetByID(mock.Anything, activeID).Return(active, nil)

	result, err := uc.ListKnownUsers(context.Background(), knowerID)

	require.NoError(t, err)
	assert.Equal(t, []*entity.User{active}, result)
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

func TestUserUsecase_GetOwnProfile_Success(t *testing.T) {
	uc, users, _, _ := newUserUsecase(t)

	userID := uuid.New()
	found := &entity.User{ID: userID, Name: "Gede"}
	users.EXPECT().GetByID(mock.Anything, userID).Return(found, nil)

	result, err := uc.GetOwnProfile(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, found, result)
}

func TestUserUsecase_UpdatePrivacySettings_Success(t *testing.T) {
	uc, users, _, _, _, _ := newUserUsecaseWithPrivacy(t)

	userID := uuid.New()
	updated := &entity.User{ID: userID, MentionPolicy: enum.AudiencePolicyKnown}
	users.EXPECT().
		UpdatePrivacySettings(mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.ID == userID && u.MentionPolicy == enum.AudiencePolicyKnown
		})).
		Return(updated, nil)

	result, err := uc.UpdatePrivacySettings(context.Background(), usecase.UpdatePrivacySettingsInput{
		UserID:                 userID,
		MentionPolicy:          enum.AudiencePolicyKnown,
		CircleInvitePolicy:     enum.AudiencePolicyAnyone,
		DiscoverableByUsername: true,
	})

	require.NoError(t, err)
	assert.Equal(t, updated, result)
}

func TestUserUsecase_BlockUser_ResolvesUsernameThenBlocks(t *testing.T) {
	uc, users, _, blocks, _, _ := newUserUsecaseWithPrivacy(t)

	blockerID := uuid.New()
	target := &entity.User{ID: uuid.New(), Username: strPtr("gede")}
	users.EXPECT().GetByUsername(mock.Anything, "gede").Return(target, nil)
	blocks.EXPECT().Block(mock.Anything, blockerID, target.ID).Return(nil)

	err := uc.BlockUser(context.Background(), blockerID, "gede")

	assert.NoError(t, err)
}

func TestUserUsecase_BlockUser_UsernameNotFound(t *testing.T) {
	uc, users, _, _, _, _ := newUserUsecaseWithPrivacy(t)

	users.EXPECT().GetByUsername(mock.Anything, "ghost").Return(nil, errs.ErrNotFound)

	// blocks.Block must never be reached — no expectation set on it at
	// all, so mockery's t.Cleanup assertion fails the test if it is.
	err := uc.BlockUser(context.Background(), uuid.New(), "ghost")

	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestUserUsecase_ListBlockedUsers_SkipsUnresolvableIDs(t *testing.T) {
	uc, users, _, blocks, _, _ := newUserUsecaseWithPrivacy(t)

	blockerID := uuid.New()
	staleID, activeID := uuid.New(), uuid.New()
	active := &entity.User{ID: activeID, Name: "Gede"}

	blocks.EXPECT().ListBlockedUserIDs(mock.Anything, blockerID).Return([]uuid.UUID{staleID, activeID}, nil)
	users.EXPECT().GetByID(mock.Anything, staleID).Return(nil, errs.ErrNotFound)
	users.EXPECT().GetByID(mock.Anything, activeID).Return(active, nil)

	result, err := uc.ListBlockedUsers(context.Background(), blockerID)

	require.NoError(t, err)
	assert.Equal(t, []*entity.User{active}, result)
}

func TestUserUsecase_MuteUser_ResolvesUsernameThenMutes(t *testing.T) {
	uc, users, _, _, mutes, _ := newUserUsecaseWithPrivacy(t)

	muterID := uuid.New()
	target := &entity.User{ID: uuid.New(), Username: strPtr("gede")}
	users.EXPECT().GetByUsername(mock.Anything, "gede").Return(target, nil)
	mutes.EXPECT().Mute(mock.Anything, muterID, target.ID).Return(nil)

	err := uc.MuteUser(context.Background(), muterID, "gede")

	assert.NoError(t, err)
}

func TestUserUsecase_ListMutedUsers_Success(t *testing.T) {
	uc, users, _, _, mutes, _ := newUserUsecaseWithPrivacy(t)

	muterID := uuid.New()
	mutedID := uuid.New()
	muted := &entity.User{ID: mutedID, Name: "Gede"}

	mutes.EXPECT().ListMutedUserIDs(mock.Anything, muterID).Return([]uuid.UUID{mutedID}, nil)
	users.EXPECT().GetByID(mock.Anything, mutedID).Return(muted, nil)

	result, err := uc.ListMutedUsers(context.Background(), muterID)

	require.NoError(t, err)
	assert.Equal(t, []*entity.User{muted}, result)
}
