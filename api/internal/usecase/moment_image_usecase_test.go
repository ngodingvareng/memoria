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

func TestMomentImageUsecase_UploadMomentImage_Success(t *testing.T) {
	repo := mocks.NewMockMomentImageRepository(t)
	store := mocks.NewMockStorage(t)
	moments := mocks.NewMockMomentAccessChecker(t)
	uc := usecase.NewMomentImageUsecase(repo, store, moments)

	momentID := uuid.New()
	userID := uuid.New()
	body := bytes.NewReader([]byte("fake image bytes"))

	moments.EXPECT().GetByID(mock.Anything, momentID, userID).
		Return(&entity.Moment{ID: momentID, UserID: userID}, nil)

	store.EXPECT().
		Put(mock.Anything, mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, "moments/"+momentID.String()+"/") && strings.HasSuffix(key, ".jpg")
		}), mock.Anything, int64(17), "image/jpeg").
		Return(nil)

	var capturedKey string
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(img *entity.MomentImage) bool {
			return img.MomentID == momentID && img.ContentType != nil && *img.ContentType == "image/jpeg"
		})).
		Run(func(ctx context.Context, img *entity.MomentImage) {
			capturedKey = img.ImagePath
		}).
		Return(&entity.MomentImage{ID: uuid.New(), MomentID: momentID, ImagePath: "placeholder"}, nil)

	store.EXPECT().
		PresignGet(mock.Anything, mock.Anything, mock.Anything).
		Return("https://storage.example.com/presigned", nil)

	result, err := uc.UploadMomentImage(context.Background(), usecase.UploadMomentImageInput{
		MomentID:    momentID,
		UserID:      userID,
		FileName:    "wet-shoes.jpg",
		ContentType: "image/jpeg",
		Size:        17,
		Body:        body,
	})

	assert.NoError(t, err)
	assert.Equal(t, "https://storage.example.com/presigned", result.URL)
	assert.NotEmpty(t, capturedKey)
}

func TestMomentImageUsecase_UploadMomentImage_NotOwner_Rejected(t *testing.T) {
	repo := mocks.NewMockMomentImageRepository(t)
	store := mocks.NewMockStorage(t)
	moments := mocks.NewMockMomentAccessChecker(t)
	uc := usecase.NewMomentImageUsecase(repo, store, moments)

	moments.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errs.ErrNotFound)
	// store.Put and repo.Create deliberately never stubbed.

	result, err := uc.UploadMomentImage(context.Background(), usecase.UploadMomentImageInput{
		MomentID: uuid.New(),
		UserID:   uuid.New(),
		FileName: "wet-shoes.jpg",
		Body:     bytes.NewReader(nil),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestMomentImageUsecase_UploadMomentImage_StorageUploadFails(t *testing.T) {
	repo := mocks.NewMockMomentImageRepository(t)
	store := mocks.NewMockStorage(t)
	moments := mocks.NewMockMomentAccessChecker(t)
	uc := usecase.NewMomentImageUsecase(repo, store, moments)

	moments.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).
		Return(&entity.Moment{}, nil)

	wantErr := errors.New("storage unreachable")
	store.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(wantErr)

	result, err := uc.UploadMomentImage(context.Background(), usecase.UploadMomentImageInput{
		MomentID: uuid.New(),
		UserID:   uuid.New(),
		FileName: "wet-shoes.jpg",
		Body:     bytes.NewReader(nil),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}

func TestMomentImageUsecase_UploadMomentImage_DBInsertFails_CleansUpStorage(t *testing.T) {
	repo := mocks.NewMockMomentImageRepository(t)
	store := mocks.NewMockStorage(t)
	moments := mocks.NewMockMomentAccessChecker(t)
	uc := usecase.NewMomentImageUsecase(repo, store, moments)

	moments.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).
		Return(&entity.Moment{}, nil)

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

	result, err := uc.UploadMomentImage(context.Background(), usecase.UploadMomentImageInput{
		MomentID: uuid.New(),
		UserID:   uuid.New(),
		FileName: "wet-shoes.jpg",
		Body:     bytes.NewReader(nil),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, dbErr)
}

func TestMomentImageUsecase_UploadMomentImage_DBInsertFails_CleanupAlsoFails(t *testing.T) {
	repo := mocks.NewMockMomentImageRepository(t)
	store := mocks.NewMockStorage(t)
	moments := mocks.NewMockMomentAccessChecker(t)
	uc := usecase.NewMomentImageUsecase(repo, store, moments)

	moments.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).
		Return(&entity.Moment{}, nil)

	dbErr := errors.New("unique constraint violated")
	cleanupErr := errors.New("storage unreachable during cleanup")

	store.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, dbErr)
	store.EXPECT().Delete(mock.Anything, mock.Anything).Return(cleanupErr)

	_, err := uc.UploadMomentImage(context.Background(), usecase.UploadMomentImageInput{
		MomentID: uuid.New(),
		UserID:   uuid.New(),
		FileName: "wet-shoes.jpg",
		Body:     bytes.NewReader(nil),
	})

	assert.ErrorIs(t, err, dbErr)
	assert.ErrorContains(t, err, cleanupErr.Error())
}

func TestMomentImageUsecase_ListMomentImages_Success(t *testing.T) {
	repo := mocks.NewMockMomentImageRepository(t)
	store := mocks.NewMockStorage(t)
	moments := mocks.NewMockMomentAccessChecker(t)
	uc := usecase.NewMomentImageUsecase(repo, store, moments)

	momentID := uuid.New()
	userID := uuid.New()
	img1 := &entity.MomentImage{ID: uuid.New(), MomentID: momentID, ImagePath: "moments/x/a.jpg"}
	img2 := &entity.MomentImage{ID: uuid.New(), MomentID: momentID, ImagePath: "moments/x/b.jpg"}

	moments.EXPECT().GetByID(mock.Anything, momentID, userID).
		Return(&entity.Moment{ID: momentID, UserID: userID}, nil)
	repo.EXPECT().ListByMomentID(mock.Anything, momentID).Return([]*entity.MomentImage{img1, img2}, nil)
	store.EXPECT().PresignGet(mock.Anything, "moments/x/a.jpg", mock.Anything).Return("url-a", nil)
	store.EXPECT().PresignGet(mock.Anything, "moments/x/b.jpg", mock.Anything).Return("url-b", nil)

	result, err := uc.ListMomentImages(context.Background(), momentID, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "url-a", result[0].URL)
	assert.Equal(t, "url-b", result[1].URL)
}

func TestMomentImageUsecase_ListMomentImages_NotOwner_Rejected(t *testing.T) {
	repo := mocks.NewMockMomentImageRepository(t)
	store := mocks.NewMockStorage(t)
	moments := mocks.NewMockMomentAccessChecker(t)
	uc := usecase.NewMomentImageUsecase(repo, store, moments)

	moments.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errs.ErrNotFound)
	// repo.ListByMomentID deliberately not stubbed.

	result, err := uc.ListMomentImages(context.Background(), uuid.New(), uuid.New())

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestMomentImageUsecase_ListMomentImages_PresignFails(t *testing.T) {
	repo := mocks.NewMockMomentImageRepository(t)
	store := mocks.NewMockStorage(t)
	moments := mocks.NewMockMomentAccessChecker(t)
	uc := usecase.NewMomentImageUsecase(repo, store, moments)

	momentID := uuid.New()
	userID := uuid.New()
	img := &entity.MomentImage{ID: uuid.New(), MomentID: momentID, ImagePath: "moments/x/a.jpg"}
	presignErr := errors.New("presign failed")

	moments.EXPECT().GetByID(mock.Anything, momentID, userID).Return(&entity.Moment{}, nil)
	repo.EXPECT().ListByMomentID(mock.Anything, momentID).Return([]*entity.MomentImage{img}, nil)
	store.EXPECT().PresignGet(mock.Anything, mock.Anything, mock.Anything).Return("", presignErr)

	result, err := uc.ListMomentImages(context.Background(), momentID, userID)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, presignErr)
}

func TestMomentImageUsecase_DeleteMomentImage_Success(t *testing.T) {
	repo := mocks.NewMockMomentImageRepository(t)
	store := mocks.NewMockStorage(t)
	moments := mocks.NewMockMomentAccessChecker(t)
	uc := usecase.NewMomentImageUsecase(repo, store, moments)

	momentID, imageID, userID := uuid.New(), uuid.New(), uuid.New()
	deleted := &entity.MomentImage{ID: imageID, MomentID: momentID, ImagePath: "moments/x/a.jpg"}
	moments.EXPECT().GetByID(mock.Anything, momentID, userID).Return(&entity.Moment{}, nil)
	repo.EXPECT().Delete(mock.Anything, momentID, imageID).Return(deleted, nil)
	store.EXPECT().Delete(mock.Anything, deleted.ImagePath).Return(nil)

	err := uc.DeleteMomentImage(context.Background(), momentID, imageID, userID)

	assert.NoError(t, err)
}

func TestMomentImageUsecase_DeleteMomentImage_NotOwner_Rejected(t *testing.T) {
	repo := mocks.NewMockMomentImageRepository(t)
	store := mocks.NewMockStorage(t)
	moments := mocks.NewMockMomentAccessChecker(t)
	uc := usecase.NewMomentImageUsecase(repo, store, moments)

	moments.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errs.ErrNotFound)
	// repo.Delete deliberately not stubbed.

	err := uc.DeleteMomentImage(context.Background(), uuid.New(), uuid.New(), uuid.New())

	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestMomentImageUsecase_DeleteMomentImage_RepoError(t *testing.T) {
	repo := mocks.NewMockMomentImageRepository(t)
	store := mocks.NewMockStorage(t)
	moments := mocks.NewMockMomentAccessChecker(t)
	uc := usecase.NewMomentImageUsecase(repo, store, moments)

	wantErr := errors.New("not found")
	moments.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).Return(&entity.Moment{}, nil)
	repo.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).Return(nil, wantErr)

	err := uc.DeleteMomentImage(context.Background(), uuid.New(), uuid.New(), uuid.New())

	assert.ErrorIs(t, err, wantErr)
}
