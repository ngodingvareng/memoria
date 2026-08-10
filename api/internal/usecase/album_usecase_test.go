//go:build unit

package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/usecase/mocks"
)

func newAlbumUsecaseDeps(t *testing.T) (*mocks.MockAlbumRepository, *mocks.MockStorage, *mocks.MockCircleAccessChecker) {
	return mocks.NewMockAlbumRepository(t), mocks.NewMockStorage(t), mocks.NewMockCircleAccessChecker(t)
}

func TestAlbumUsecase_ListAlbum_Success_PresignsEveryImage(t *testing.T) {
	repo, store, circles := newAlbumUsecaseDeps(t)
	uc := usecase.NewAlbumUsecase(repo, store, circles)

	userID := uuid.New()
	img1 := &entity.AlbumImage{ID: uuid.New(), ImagePath: "moments/a/1.jpg"}
	img2 := &entity.AlbumImage{ID: uuid.New(), ImagePath: "moments/a/2.jpg"}

	repo.EXPECT().ListPersonalImages(mock.Anything, userID, int32(20), int32(0)).
		Return([]*entity.AlbumImage{img1, img2}, nil)
	store.EXPECT().PresignGet(mock.Anything, "moments/a/1.jpg", mock.Anything).
		Return("https://storage.example.com/1", nil)
	store.EXPECT().PresignGet(mock.Anything, "moments/a/2.jpg", mock.Anything).
		Return("https://storage.example.com/2", nil)

	result, err := uc.ListAlbum(context.Background(), usecase.ListAlbumInput{UserID: userID})

	assert.NoError(t, err)
	assert.Len(t, result.Images, 2)
	assert.Equal(t, "https://storage.example.com/1", result.Images[0].URL)
	assert.Equal(t, "https://storage.example.com/2", result.Images[1].URL)
}

func TestAlbumUsecase_ListAlbum_DefaultsPageAndPageSize(t *testing.T) {
	repo, store, circles := newAlbumUsecaseDeps(t)
	uc := usecase.NewAlbumUsecase(repo, store, circles)

	userID := uuid.New()
	repo.EXPECT().ListPersonalImages(mock.Anything, userID, int32(20), int32(0)).
		Return(nil, nil)

	result, err := uc.ListAlbum(context.Background(), usecase.ListAlbumInput{UserID: userID, Page: 0, PageSize: 0})

	assert.NoError(t, err)
	assert.Equal(t, int32(1), result.Page)
	assert.Equal(t, int32(20), result.PageSize)
}

func TestAlbumUsecase_ListAlbum_StoragePresignFails_ErrorPropagated(t *testing.T) {
	repo, store, circles := newAlbumUsecaseDeps(t)
	uc := usecase.NewAlbumUsecase(repo, store, circles)

	userID := uuid.New()
	img := &entity.AlbumImage{ID: uuid.New(), ImagePath: "moments/a/1.jpg"}
	repo.EXPECT().ListPersonalImages(mock.Anything, userID, int32(20), int32(0)).
		Return([]*entity.AlbumImage{img}, nil)
	store.EXPECT().PresignGet(mock.Anything, "moments/a/1.jpg", mock.Anything).
		Return("", errors.New("boom"))

	result, err := uc.ListAlbum(context.Background(), usecase.ListAlbumInput{UserID: userID})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAlbumUsecase_ListCircleAlbum_NotMember_ErrorPropagated_RepoNeverCalled(t *testing.T) {
	repo, store, circles := newAlbumUsecaseDeps(t)
	uc := usecase.NewAlbumUsecase(repo, store, circles)

	userID := uuid.New()
	circleID := uuid.New()
	circles.EXPECT().GetActiveMember(mock.Anything, circleID, userID).Return(nil, errs.ErrNotFound)
	// repo.ListCircleImages deliberately never stubbed — asserting it's
	// never called.

	result, err := uc.ListCircleAlbum(context.Background(), usecase.ListCircleAlbumInput{
		UserID:   userID,
		CircleID: circleID,
	})

	assert.ErrorIs(t, err, errs.ErrNotFound)
	assert.Nil(t, result)
}

func TestAlbumUsecase_ListCircleAlbum_Success_AttributionPassesThrough(t *testing.T) {
	repo, store, circles := newAlbumUsecaseDeps(t)
	uc := usecase.NewAlbumUsecase(repo, store, circles)

	userID := uuid.New()
	circleID := uuid.New()
	sharerID := uuid.New()
	sharerUsername := "gede"

	nativeImg := &entity.AlbumImage{ID: uuid.New(), ImagePath: "moments/native/1.jpg", IsShared: false}
	sharedImg := &entity.AlbumImage{
		ID:               uuid.New(),
		ImagePath:        "moments/shared/1.jpg",
		IsShared:         true,
		SharedByUserID:   &sharerID,
		SharedByUsername: &sharerUsername,
	}

	circles.EXPECT().GetActiveMember(mock.Anything, circleID, userID).Return(&entity.CircleMember{}, nil)
	repo.EXPECT().ListCircleImages(mock.Anything, circleID, int32(20), int32(0)).
		Return([]*entity.AlbumImage{nativeImg, sharedImg}, nil)
	store.EXPECT().PresignGet(mock.Anything, mock.Anything, mock.Anything).
		Return("https://storage.example.com/presigned", nil)

	result, err := uc.ListCircleAlbum(context.Background(), usecase.ListCircleAlbumInput{
		UserID:   userID,
		CircleID: circleID,
	})

	assert.NoError(t, err)
	assert.Len(t, result.Images, 2)
	assert.False(t, result.Images[0].IsShared)
	assert.Nil(t, result.Images[0].SharedByUserID)
	assert.True(t, result.Images[1].IsShared)
	assert.Equal(t, sharerID, *result.Images[1].SharedByUserID)
	assert.Equal(t, sharerUsername, *result.Images[1].SharedByUsername)
}
