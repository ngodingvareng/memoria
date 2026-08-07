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

func TestThreadImageUsecase_UploadThreadImage_Success(t *testing.T) {
	repo := mocks.NewMockThreadImageRepository(t)
	store := mocks.NewMockStorage(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewThreadImageUsecase(repo, store, threads)

	threadID := uuid.New()
	userID := uuid.New()
	body := bytes.NewReader([]byte("fake image bytes"))

	threads.EXPECT().GetByID(mock.Anything, threadID, userID).
		Return(&entity.Thread{ID: threadID, UserID: userID}, nil)

	store.EXPECT().
		Put(mock.Anything, mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, "threads/"+threadID.String()+"/") && strings.HasSuffix(key, ".jpg")
		}), mock.Anything, int64(17), "image/jpeg").
		Return(nil)

	var capturedKey string
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(img *entity.ThreadImage) bool {
			return img.ThreadID == threadID
		})).
		Run(func(ctx context.Context, img *entity.ThreadImage) {
			capturedKey = img.ImagePath
		}).
		Return(&entity.ThreadImage{ID: uuid.New(), ThreadID: threadID, ImagePath: "placeholder"}, nil)

	store.EXPECT().
		PresignGet(mock.Anything, mock.Anything, mock.Anything).
		Return("https://storage.example.com/presigned", nil)

	result, err := uc.UploadThreadImage(context.Background(), usecase.UploadThreadImageInput{
		ThreadID:    threadID,
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

func TestThreadImageUsecase_UploadThreadImage_NotOwner_Rejected(t *testing.T) {
	repo := mocks.NewMockThreadImageRepository(t)
	store := mocks.NewMockStorage(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewThreadImageUsecase(repo, store, threads)

	threads.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errs.ErrNotFound)
	// store.Put and repo.Create deliberately never stubbed — if the
	// usecase somehow reached storage/DB after a failed ownership check,
	// this test fails with "unexpected call" instead of silently passing.

	result, err := uc.UploadThreadImage(context.Background(), usecase.UploadThreadImageInput{
		ThreadID: uuid.New(),
		UserID:   uuid.New(), // different from the thread's real owner
		FileName: "cat.jpg",
		Body:     bytes.NewReader(nil),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestThreadImageUsecase_UploadThreadImage_StorageUploadFails(t *testing.T) {
	repo := mocks.NewMockThreadImageRepository(t)
	store := mocks.NewMockStorage(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewThreadImageUsecase(repo, store, threads)

	threads.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).
		Return(&entity.Thread{}, nil)

	wantErr := errors.New("storage unreachable")
	store.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(wantErr)

	result, err := uc.UploadThreadImage(context.Background(), usecase.UploadThreadImageInput{
		ThreadID: uuid.New(),
		UserID:   uuid.New(),
		FileName: "cat.jpg",
		Body:     bytes.NewReader(nil),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}

func TestThreadImageUsecase_UploadThreadImage_DBInsertFails_CleansUpStorage(t *testing.T) {
	repo := mocks.NewMockThreadImageRepository(t)
	store := mocks.NewMockStorage(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewThreadImageUsecase(repo, store, threads)

	threads.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).
		Return(&entity.Thread{}, nil)

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

	result, err := uc.UploadThreadImage(context.Background(), usecase.UploadThreadImageInput{
		ThreadID: uuid.New(),
		UserID:   uuid.New(),
		FileName: "cat.jpg",
		Body:     bytes.NewReader(nil),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, dbErr)
}

func TestThreadImageUsecase_UploadThreadImage_DBInsertFails_CleanupAlsoFails(t *testing.T) {
	repo := mocks.NewMockThreadImageRepository(t)
	store := mocks.NewMockStorage(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewThreadImageUsecase(repo, store, threads)

	threads.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).
		Return(&entity.Thread{}, nil)

	dbErr := errors.New("unique constraint violated")
	cleanupErr := errors.New("storage unreachable during cleanup")

	store.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, dbErr)
	store.EXPECT().Delete(mock.Anything, mock.Anything).Return(cleanupErr)

	_, err := uc.UploadThreadImage(context.Background(), usecase.UploadThreadImageInput{
		ThreadID: uuid.New(),
		UserID:   uuid.New(),
		FileName: "cat.jpg",
		Body:     bytes.NewReader(nil),
	})

	assert.ErrorIs(t, err, dbErr)
	assert.ErrorContains(t, err, cleanupErr.Error())
}

func TestThreadImageUsecase_ListThreadImages_Success(t *testing.T) {
	repo := mocks.NewMockThreadImageRepository(t)
	store := mocks.NewMockStorage(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewThreadImageUsecase(repo, store, threads)

	threadID := uuid.New()
	userID := uuid.New()
	img1 := &entity.ThreadImage{ID: uuid.New(), ThreadID: threadID, ImagePath: "threads/x/a.jpg"}
	img2 := &entity.ThreadImage{ID: uuid.New(), ThreadID: threadID, ImagePath: "threads/x/b.jpg"}

	threads.EXPECT().GetByID(mock.Anything, threadID, userID).
		Return(&entity.Thread{ID: threadID, UserID: userID}, nil)
	repo.EXPECT().ListByThreadID(mock.Anything, threadID).Return([]*entity.ThreadImage{img1, img2}, nil)
	store.EXPECT().PresignGet(mock.Anything, "threads/x/a.jpg", mock.Anything).Return("url-a", nil)
	store.EXPECT().PresignGet(mock.Anything, "threads/x/b.jpg", mock.Anything).Return("url-b", nil)

	result, err := uc.ListThreadImages(context.Background(), threadID, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "url-a", result[0].URL)
	assert.Equal(t, "url-b", result[1].URL)
}

func TestThreadImageUsecase_ListThreadImages_NotOwner_Rejected(t *testing.T) {
	repo := mocks.NewMockThreadImageRepository(t)
	store := mocks.NewMockStorage(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewThreadImageUsecase(repo, store, threads)

	threads.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errs.ErrNotFound)
	// repo.ListByThreadID deliberately not stubbed.

	result, err := uc.ListThreadImages(context.Background(), uuid.New(), uuid.New())

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestThreadImageUsecase_ListThreadImages_PresignFails(t *testing.T) {
	repo := mocks.NewMockThreadImageRepository(t)
	store := mocks.NewMockStorage(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewThreadImageUsecase(repo, store, threads)

	threadID := uuid.New()
	userID := uuid.New()
	img := &entity.ThreadImage{ID: uuid.New(), ThreadID: threadID, ImagePath: "threads/x/a.jpg"}
	presignErr := errors.New("presign failed")

	threads.EXPECT().GetByID(mock.Anything, threadID, userID).Return(&entity.Thread{}, nil)
	repo.EXPECT().ListByThreadID(mock.Anything, threadID).Return([]*entity.ThreadImage{img}, nil)
	store.EXPECT().PresignGet(mock.Anything, mock.Anything, mock.Anything).Return("", presignErr)

	result, err := uc.ListThreadImages(context.Background(), threadID, userID)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, presignErr)
}

func TestThreadImageUsecase_DeleteThreadImage_Success(t *testing.T) {
	repo := mocks.NewMockThreadImageRepository(t)
	store := mocks.NewMockStorage(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewThreadImageUsecase(repo, store, threads)

	threadID, imageID, userID := uuid.New(), uuid.New(), uuid.New()
	threads.EXPECT().GetByID(mock.Anything, threadID, userID).Return(&entity.Thread{}, nil)
	repo.EXPECT().Delete(mock.Anything, threadID, imageID).Return(nil)

	err := uc.DeleteThreadImage(context.Background(), threadID, imageID, userID)

	assert.NoError(t, err)
}

func TestThreadImageUsecase_DeleteThreadImage_NotOwner_Rejected(t *testing.T) {
	repo := mocks.NewMockThreadImageRepository(t)
	store := mocks.NewMockStorage(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewThreadImageUsecase(repo, store, threads)

	threads.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errs.ErrNotFound)
	// repo.Delete deliberately not stubbed.

	err := uc.DeleteThreadImage(context.Background(), uuid.New(), uuid.New(), uuid.New())

	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestThreadImageUsecase_DeleteThreadImage_RepoError(t *testing.T) {
	repo := mocks.NewMockThreadImageRepository(t)
	store := mocks.NewMockStorage(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewThreadImageUsecase(repo, store, threads)

	wantErr := errors.New("not found")
	threads.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).Return(&entity.Thread{}, nil)
	repo.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).Return(wantErr)

	err := uc.DeleteThreadImage(context.Background(), uuid.New(), uuid.New(), uuid.New())

	assert.ErrorIs(t, err, wantErr)
}
