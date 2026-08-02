//go:build unit

package usecase_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/usecase/mocks"
)

func TestActivityImageUsecase_UploadActivityImage_Success(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	activities := mocks.NewMockActivityAccessChecker(t)
	uc := usecase.NewActivityImageUsecase(repo, store, activities)

	activityID := uuid.New()
	userID := uuid.New()
	body := bytes.NewReader([]byte("fake image bytes"))

	activities.EXPECT().GetActivityByID(mock.Anything, activityID, userID).
		Return(&entity.Activity{ID: activityID, UserID: userID}, nil)

	store.EXPECT().
		Put(mock.Anything, mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, "activities/"+activityID.String()+"/") && strings.HasSuffix(key, ".jpg")
		}), mock.Anything, int64(17), "image/jpeg").
		Return(nil)

	var capturedKey string
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(img *entity.ActivityImage) bool {
			return img.ActivityID == activityID
		})).
		Run(func(ctx context.Context, img *entity.ActivityImage) {
			capturedKey = img.ImagePath
		}).
		Return(&entity.ActivityImage{ID: uuid.New(), ActivityID: activityID, ImagePath: "placeholder"}, nil)

	store.EXPECT().
		PresignGet(mock.Anything, mock.Anything, mock.Anything).
		Return("https://storage.example.com/presigned", nil)

	result, err := uc.UploadActivityImage(context.Background(), usecase.UploadActivityImageInput{
		ActivityID:  activityID,
		UserID:      userID,
		FileName:    "cat.jpg",
		ContentType: "image/jpeg",
		Size:        17,
		Body:        body,
	})

	assert.NoError(t, err)
	assert.Equal(t, "https://storage.example.com/presigned", result.URL)
	assert.NotEmpty(t, capturedKey)
}

func TestActivityImageUsecase_UploadActivityImage_NotOwner_Rejected(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	activities := mocks.NewMockActivityAccessChecker(t)
	uc := usecase.NewActivityImageUsecase(repo, store, activities)

	activities.EXPECT().GetActivityByID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errs.ErrNotFound)
	// store.Put and repo.Create deliberately never stubbed — if the
	// usecase somehow reached storage/DB after a failed ownership check,
	// this test fails with "unexpected call" instead of silently passing.

	result, err := uc.UploadActivityImage(context.Background(), usecase.UploadActivityImageInput{
		ActivityID: uuid.New(),
		UserID:     uuid.New(), // different from the activity's real owner
		FileName:   "cat.jpg",
		Body:       bytes.NewReader(nil),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestActivityImageUsecase_UploadActivityImage_StorageUploadFails(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	activities := mocks.NewMockActivityAccessChecker(t)
	uc := usecase.NewActivityImageUsecase(repo, store, activities)

	activities.EXPECT().GetActivityByID(mock.Anything, mock.Anything, mock.Anything).
		Return(&entity.Activity{}, nil)

	wantErr := errors.New("storage unreachable")
	store.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(wantErr)

	result, err := uc.UploadActivityImage(context.Background(), usecase.UploadActivityImageInput{
		ActivityID: uuid.New(),
		UserID:     uuid.New(),
		FileName:   "cat.jpg",
		Body:       bytes.NewReader(nil),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}

func TestActivityImageUsecase_UploadActivityImage_DBInsertFails_CleansUpStorage(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	activities := mocks.NewMockActivityAccessChecker(t)
	uc := usecase.NewActivityImageUsecase(repo, store, activities)

	activities.EXPECT().GetActivityByID(mock.Anything, mock.Anything, mock.Anything).
		Return(&entity.Activity{}, nil)

	dbErr := errors.New("unique constraint violated")
	var uploadedKey string

	store.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(ctx context.Context, key string, body io.Reader, size int64, contentType string) {
			uploadedKey = key
		}).
		Return(nil)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, dbErr)

	store.EXPECT().
		Delete(mock.Anything, mock.MatchedBy(func(key string) bool { return key == uploadedKey })).
		Return(nil)

	result, err := uc.UploadActivityImage(context.Background(), usecase.UploadActivityImageInput{
		ActivityID: uuid.New(),
		UserID:     uuid.New(),
		FileName:   "cat.jpg",
		Body:       bytes.NewReader(nil),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, dbErr)
}

func TestActivityImageUsecase_UploadActivityImage_DBInsertFails_CleanupAlsoFails(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	activities := mocks.NewMockActivityAccessChecker(t)
	uc := usecase.NewActivityImageUsecase(repo, store, activities)

	activities.EXPECT().GetActivityByID(mock.Anything, mock.Anything, mock.Anything).
		Return(&entity.Activity{}, nil)

	dbErr := errors.New("unique constraint violated")
	cleanupErr := errors.New("storage unreachable during cleanup")

	store.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, dbErr)
	store.EXPECT().Delete(mock.Anything, mock.Anything).Return(cleanupErr)

	_, err := uc.UploadActivityImage(context.Background(), usecase.UploadActivityImageInput{
		ActivityID: uuid.New(),
		UserID:     uuid.New(),
		FileName:   "cat.jpg",
		Body:       bytes.NewReader(nil),
	})

	assert.ErrorIs(t, err, dbErr)
	assert.ErrorContains(t, err, cleanupErr.Error())
}

func TestActivityImageUsecase_ListActivityImages_Success(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	activities := mocks.NewMockActivityAccessChecker(t)
	uc := usecase.NewActivityImageUsecase(repo, store, activities)

	activityID := uuid.New()
	userID := uuid.New()
	img1 := &entity.ActivityImage{ID: uuid.New(), ActivityID: activityID, ImagePath: "activities/x/a.jpg"}
	img2 := &entity.ActivityImage{ID: uuid.New(), ActivityID: activityID, ImagePath: "activities/x/b.jpg"}

	activities.EXPECT().GetActivityByID(mock.Anything, activityID, userID).
		Return(&entity.Activity{ID: activityID, UserID: userID}, nil)
	repo.EXPECT().ListByActivityID(mock.Anything, activityID).Return([]*entity.ActivityImage{img1, img2}, nil)
	store.EXPECT().PresignGet(mock.Anything, "activities/x/a.jpg", mock.Anything).Return("url-a", nil)
	store.EXPECT().PresignGet(mock.Anything, "activities/x/b.jpg", mock.Anything).Return("url-b", nil)

	result, err := uc.ListActivityImages(context.Background(), activityID, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "url-a", result[0].URL)
	assert.Equal(t, "url-b", result[1].URL)
}

func TestActivityImageUsecase_ListActivityImages_NotOwner_Rejected(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	activities := mocks.NewMockActivityAccessChecker(t)
	uc := usecase.NewActivityImageUsecase(repo, store, activities)

	activities.EXPECT().GetActivityByID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errs.ErrNotFound)
	// repo.ListByActivityID deliberately not stubbed.

	result, err := uc.ListActivityImages(context.Background(), uuid.New(), uuid.New())

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestActivityImageUsecase_ListActivityImages_PresignFails(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	activities := mocks.NewMockActivityAccessChecker(t)
	uc := usecase.NewActivityImageUsecase(repo, store, activities)

	activityID := uuid.New()
	userID := uuid.New()
	img := &entity.ActivityImage{ID: uuid.New(), ActivityID: activityID, ImagePath: "activities/x/a.jpg"}
	presignErr := errors.New("presign failed")

	activities.EXPECT().GetActivityByID(mock.Anything, activityID, userID).Return(&entity.Activity{}, nil)
	repo.EXPECT().ListByActivityID(mock.Anything, activityID).Return([]*entity.ActivityImage{img}, nil)
	store.EXPECT().PresignGet(mock.Anything, mock.Anything, mock.Anything).Return("", presignErr)

	result, err := uc.ListActivityImages(context.Background(), activityID, userID)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, presignErr)
}

func TestActivityImageUsecase_DeleteActivityImage_Success(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	activities := mocks.NewMockActivityAccessChecker(t)
	uc := usecase.NewActivityImageUsecase(repo, store, activities)

	activityID, imageID, userID := uuid.New(), uuid.New(), uuid.New()
	activities.EXPECT().GetActivityByID(mock.Anything, activityID, userID).Return(&entity.Activity{}, nil)
	repo.EXPECT().Delete(mock.Anything, activityID, imageID).Return(nil)

	err := uc.DeleteActivityImage(context.Background(), activityID, imageID, userID)

	assert.NoError(t, err)
}

func TestActivityImageUsecase_DeleteActivityImage_NotOwner_Rejected(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	activities := mocks.NewMockActivityAccessChecker(t)
	uc := usecase.NewActivityImageUsecase(repo, store, activities)

	activities.EXPECT().GetActivityByID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errs.ErrNotFound)
	// repo.Delete deliberately not stubbed.

	err := uc.DeleteActivityImage(context.Background(), uuid.New(), uuid.New(), uuid.New())

	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestActivityImageUsecase_DeleteActivityImage_RepoError(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	activities := mocks.NewMockActivityAccessChecker(t)
	uc := usecase.NewActivityImageUsecase(repo, store, activities)

	wantErr := errors.New("not found")
	activities.EXPECT().GetActivityByID(mock.Anything, mock.Anything, mock.Anything).Return(&entity.Activity{}, nil)
	repo.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).Return(wantErr)

	err := uc.DeleteActivityImage(context.Background(), uuid.New(), uuid.New(), uuid.New())

	assert.ErrorIs(t, err, wantErr)
}
