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
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/usecase/mocks"
)

func TestActivityImageUsecase_UploadActivityImage_Success(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	uc := usecase.NewActivityImageUsecase(repo, store)

	activityID := uuid.New()
	body := bytes.NewReader([]byte("fake image bytes"))

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
		FileName:    "cat.jpg",
		ContentType: "image/jpeg",
		Size:        17,
		Body:        body,
	})

	assert.NoError(t, err)
	assert.Equal(t, "https://storage.example.com/presigned", result.URL)
	assert.NotEmpty(t, capturedKey)
}

func TestActivityImageUsecase_UploadActivityImage_StorageUploadFails(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	uc := usecase.NewActivityImageUsecase(repo, store)

	wantErr := errors.New("storage unreachable")
	store.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(wantErr)
	// repo.Create is deliberately never stubbed — if the usecase called
	// it after a failed upload, this test would fail with "unexpected
	// call" rather than silently passing.

	result, err := uc.UploadActivityImage(context.Background(), usecase.UploadActivityImageInput{
		ActivityID: uuid.New(),
		FileName:   "cat.jpg",
		Body:       bytes.NewReader(nil),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}

func TestActivityImageUsecase_UploadActivityImage_DBInsertFails_CleansUpStorage(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	uc := usecase.NewActivityImageUsecase(repo, store)

	dbErr := errors.New("unique constraint violated")
	var uploadedKey string

	store.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(ctx context.Context, key string, body io.Reader, size int64, contentType string) {
			uploadedKey = key
		}).
		Return(nil)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, dbErr)

	// The cleanup delete must target the SAME key that was uploaded.
	store.EXPECT().
		Delete(mock.Anything, mock.MatchedBy(func(key string) bool { return key == uploadedKey })).
		Return(nil)

	result, err := uc.UploadActivityImage(context.Background(), usecase.UploadActivityImageInput{
		ActivityID: uuid.New(),
		FileName:   "cat.jpg",
		Body:       bytes.NewReader(nil),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, dbErr)
}

func TestActivityImageUsecase_UploadActivityImage_DBInsertFails_CleanupAlsoFails(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	uc := usecase.NewActivityImageUsecase(repo, store)

	dbErr := errors.New("unique constraint violated")
	cleanupErr := errors.New("storage unreachable during cleanup")

	store.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, dbErr)
	store.EXPECT().Delete(mock.Anything, mock.Anything).Return(cleanupErr)

	_, err := uc.UploadActivityImage(context.Background(), usecase.UploadActivityImageInput{
		ActivityID: uuid.New(),
		FileName:   "cat.jpg",
		Body:       bytes.NewReader(nil),
	})

	// Both failures should be visible in the error, not just one
	// silently swallowing the other — this is the scenario the code
	// comment calls out as "still ends up orphaned" (tracked in
	// SCHEMA_REVIEW.md / TODO.md), so at minimum the error needs to say so.
	assert.ErrorIs(t, err, dbErr)
	assert.ErrorContains(t, err, cleanupErr.Error())
}

func TestActivityImageUsecase_ListActivityImages_Success(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	uc := usecase.NewActivityImageUsecase(repo, store)

	activityID := uuid.New()
	img1 := &entity.ActivityImage{ID: uuid.New(), ActivityID: activityID, ImagePath: "activities/x/a.jpg"}
	img2 := &entity.ActivityImage{ID: uuid.New(), ActivityID: activityID, ImagePath: "activities/x/b.jpg"}

	repo.EXPECT().ListByActivityID(mock.Anything, activityID).Return([]*entity.ActivityImage{img1, img2}, nil)
	store.EXPECT().PresignGet(mock.Anything, "activities/x/a.jpg", mock.Anything).Return("url-a", nil)
	store.EXPECT().PresignGet(mock.Anything, "activities/x/b.jpg", mock.Anything).Return("url-b", nil)

	result, err := uc.ListActivityImages(context.Background(), activityID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "url-a", result[0].URL)
	assert.Equal(t, "url-b", result[1].URL)
}

func TestActivityImageUsecase_ListActivityImages_PresignFails(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	uc := usecase.NewActivityImageUsecase(repo, store)

	activityID := uuid.New()
	img := &entity.ActivityImage{ID: uuid.New(), ActivityID: activityID, ImagePath: "activities/x/a.jpg"}
	presignErr := errors.New("presign failed")

	repo.EXPECT().ListByActivityID(mock.Anything, activityID).Return([]*entity.ActivityImage{img}, nil)
	store.EXPECT().PresignGet(mock.Anything, mock.Anything, mock.Anything).Return("", presignErr)

	result, err := uc.ListActivityImages(context.Background(), activityID)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, presignErr)
}

func TestActivityImageUsecase_SoftDeleteActivityImage_Success(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	uc := usecase.NewActivityImageUsecase(repo, store)

	activityID, imageID := uuid.New(), uuid.New()
	repo.EXPECT().Delete(mock.Anything, activityID, imageID).Return(nil)

	err := uc.DeleteActivityImage(context.Background(), activityID, imageID)

	assert.NoError(t, err)
}

func TestActivityImageUsecase_SoftDeleteActivityImage_RepoError(t *testing.T) {
	repo := mocks.NewMockActivityImageRepository(t)
	store := mocks.NewMockStorage(t)
	uc := usecase.NewActivityImageUsecase(repo, store)

	wantErr := errors.New("not found")
	repo.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).Return(wantErr)

	err := uc.DeleteActivityImage(context.Background(), uuid.New(), uuid.New())

	assert.ErrorIs(t, err, wantErr)
}
