//go:build unit

package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func expectPassthroughMomentTransaction(repo *mocks.MockMomentRepository) {
	repo.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(usecase.MomentRepository) error) error {
			return fn(repo)
		})
}

// --- CreateMoment ---

func TestMomentUsecase_CreateMoment_Success(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	userID := uuid.New()
	occurredAt := time.Date(2026, 8, 4, 5, 30, 0, 0, time.UTC)
	expected := &entity.Moment{ID: uuid.New(), UserID: userID, Origin: enum.MomentOriginManual}

	expectPassthroughMomentTransaction(repo)
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(m *entity.Moment) bool {
			return m.UserID == userID &&
				m.Origin == enum.MomentOriginManual &&
				m.OccurredAt.Equal(occurredAt) &&
				m.ThreadID == nil
		})).
		Return(expected, nil)

	result, err := uc.CreateMoment(context.Background(), usecase.CreateMomentInput{
		UserID:     userID,
		OccurredAt: occurredAt,
	})

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestMomentUsecase_CreateMoment_DerivesOccurredLocalFromOffset(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	// 05:30 UTC with a +540 (JST, +9h) offset should read as 14:30 local.
	occurredAt := time.Date(2026, 8, 4, 5, 30, 0, 0, time.UTC)
	var captured *entity.Moment

	expectPassthroughMomentTransaction(repo)
	repo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Run(func(ctx context.Context, m *entity.Moment) { captured = m }).
		Return(&entity.Moment{}, nil)

	_, err := uc.CreateMoment(context.Background(), usecase.CreateMomentInput{
		UserID:                   uuid.New(),
		OccurredAt:               occurredAt,
		OccurredUTCOffsetMinutes: 540,
	})

	assert.NoError(t, err)
	assert.Equal(t, 14, captured.OccurredLocal.Hour())
	assert.Equal(t, 30, captured.OccurredLocal.Minute())
	assert.Equal(t, int16(540), captured.OccurredUTCOffsetMinutes)
}

func TestMomentUsecase_CreateMoment_DefaultsOriginToManual(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	var captured *entity.Moment
	expectPassthroughMomentTransaction(repo)
	repo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Run(func(ctx context.Context, m *entity.Moment) { captured = m }).
		Return(&entity.Moment{}, nil)

	_, err := uc.CreateMoment(context.Background(), usecase.CreateMomentInput{
		UserID:     uuid.New(),
		OccurredAt: time.Now(),
		// Origin intentionally left empty.
	})

	assert.NoError(t, err)
	assert.Equal(t, enum.MomentOriginManual, captured.Origin)
}

func TestMomentUsecase_CreateMoment_WithThread_ChecksOwnership(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	userID := uuid.New()
	threadID := uuid.New()

	threads.EXPECT().GetByID(mock.Anything, threadID, userID).
		Return(&entity.Thread{ID: threadID, UserID: userID}, nil)

	expectPassthroughMomentTransaction(repo)
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(m *entity.Moment) bool {
			return m.ThreadID != nil && *m.ThreadID == threadID
		})).
		Return(&entity.Moment{}, nil)

	_, err := uc.CreateMoment(context.Background(), usecase.CreateMomentInput{
		UserID:     userID,
		ThreadID:   &threadID,
		OccurredAt: time.Now(),
	})

	assert.NoError(t, err)
}

func TestMomentUsecase_CreateMoment_ThreadNotOwned_Rejected(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	threadID := uuid.New()
	threads.EXPECT().GetByID(mock.Anything, threadID, mock.Anything).
		Return(nil, errs.ErrNotFound)
	// repo.Create deliberately not stubbed — a call here would fail the
	// test with "unexpected call".

	result, err := uc.CreateMoment(context.Background(), usecase.CreateMomentInput{
		UserID:     uuid.New(),
		ThreadID:   &threadID,
		OccurredAt: time.Now(),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestMomentUsecase_CreateMoment_RepositoryError(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	wantErr := errors.New("db exploded")
	repo.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		Return(wantErr)

	result, err := uc.CreateMoment(context.Background(), usecase.CreateMomentInput{
		UserID:     uuid.New(),
		OccurredAt: time.Now(),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}

// --- UpdateMoment ---

func TestMomentUsecase_UpdateMoment_Success(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	momentID := uuid.New()
	userID := uuid.New()
	expected := &entity.Moment{ID: momentID, UserID: userID}

	expectPassthroughMomentTransaction(repo)
	repo.EXPECT().
		Update(mock.Anything, mock.MatchedBy(func(m *entity.Moment) bool {
			return m.ID == momentID && m.UserID == userID
		})).
		Return(expected, nil)

	result, err := uc.UpdateMoment(context.Background(), usecase.UpdateMomentInput{
		ID:         momentID,
		UserID:     userID,
		OccurredAt: time.Now(),
	})

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestMomentUsecase_UpdateMoment_NotFound(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	expectPassthroughMomentTransaction(repo)
	repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil, errs.ErrNotFound)

	result, err := uc.UpdateMoment(context.Background(), usecase.UpdateMomentInput{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		OccurredAt: time.Now(),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestMomentUsecase_UpdateMoment_ThreadNotOwned_Rejected(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	threadID := uuid.New()
	threads.EXPECT().GetByID(mock.Anything, threadID, mock.Anything).Return(nil, errs.ErrNotFound)

	result, err := uc.UpdateMoment(context.Background(), usecase.UpdateMomentInput{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		ThreadID:   &threadID,
		OccurredAt: time.Now(),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

// --- SoftDeleteMoment ---

func TestMomentUsecase_SoftDeleteMoment_Success(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	momentID, userID := uuid.New(), uuid.New()
	expectPassthroughMomentTransaction(repo)
	repo.EXPECT().SoftDelete(mock.Anything, momentID, userID).Return(nil)

	err := uc.SoftDeleteMoment(context.Background(), momentID, userID)

	assert.NoError(t, err)
}

func TestMomentUsecase_SoftDeleteMoment_RepositoryError(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	wantErr := errors.New("db exploded")
	expectPassthroughMomentTransaction(repo)
	repo.EXPECT().SoftDelete(mock.Anything, mock.Anything, mock.Anything).Return(wantErr)

	err := uc.SoftDeleteMoment(context.Background(), uuid.New(), uuid.New())

	assert.ErrorIs(t, err, wantErr)
}

// --- GetMoment ---

func TestMomentUsecase_GetMoment_Success(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	userID, momentID := uuid.New(), uuid.New()
	expected := &entity.Moment{ID: momentID, UserID: userID}
	repo.EXPECT().GetByID(mock.Anything, momentID, userID).Return(expected, nil)

	result, err := uc.GetMoment(context.Background(), userID, momentID)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestMomentUsecase_GetMoment_NotFound(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	repo.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).Return(nil, errs.ErrNotFound)

	result, err := uc.GetMoment(context.Background(), uuid.New(), uuid.New())

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

// --- ListMoments ---

func TestMomentUsecase_ListMoments_DefaultsPagination(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	userID := uuid.New()
	moments := []*entity.Moment{{ID: uuid.New(), UserID: userID}}

	repo.EXPECT().ListByUserID(mock.Anything, userID, int32(20), int32(0)).Return(moments, nil)

	result, err := uc.ListMoments(context.Background(), usecase.ListMomentsInput{UserID: userID})

	assert.NoError(t, err)
	assert.Equal(t, moments, result.Moments)
	assert.Equal(t, int32(1), result.Page)
	assert.Equal(t, int32(20), result.PageSize)
}

func TestMomentUsecase_ListMoments_CapsPageSize(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	userID := uuid.New()
	repo.EXPECT().ListByUserID(mock.Anything, userID, int32(100), int32(200)).Return(nil, nil)

	result, err := uc.ListMoments(context.Background(), usecase.ListMomentsInput{
		UserID:   userID,
		Page:     3,
		PageSize: 500,
	})

	assert.NoError(t, err)
	assert.Equal(t, int32(100), result.PageSize)
}

// --- ListThreadMoments ---

func TestMomentUsecase_ListThreadMoments_Success(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	userID, threadID := uuid.New(), uuid.New()
	moments := []*entity.Moment{{ID: uuid.New(), ThreadID: &threadID}}

	threads.EXPECT().GetByID(mock.Anything, threadID, userID).Return(&entity.Thread{}, nil)
	repo.EXPECT().ListByThreadID(mock.Anything, threadID, int32(20), int32(0)).Return(moments, nil)

	result, err := uc.ListThreadMoments(context.Background(), usecase.ListThreadMomentsInput{
		UserID:   userID,
		ThreadID: threadID,
	})

	assert.NoError(t, err)
	assert.Equal(t, moments, result.Moments)
}

func TestMomentUsecase_ListThreadMoments_ThreadNotOwned_Rejected(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	threads.EXPECT().GetByID(mock.Anything, mock.Anything, mock.Anything).Return(nil, errs.ErrNotFound)
	// repo.ListByThreadID deliberately not stubbed.

	result, err := uc.ListThreadMoments(context.Background(), usecase.ListThreadMomentsInput{
		UserID:   uuid.New(),
		ThreadID: uuid.New(),
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

// --- SearchMoments ---

func TestMomentUsecase_SearchMoments_Success(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	userID := uuid.New()
	moments := []*entity.Moment{{ID: uuid.New(), UserID: userID}}

	repo.EXPECT().Search(mock.Anything, userID, "wet shoes", int32(20), int32(0)).Return(moments, nil)

	result, err := uc.SearchMoments(context.Background(), usecase.SearchMomentsInput{
		UserID: userID,
		Query:  "wet shoes",
	})

	assert.NoError(t, err)
	assert.Equal(t, moments, result.Moments)
}

func TestMomentUsecase_SearchMoments_RepositoryError(t *testing.T) {
	repo := mocks.NewMockMomentRepository(t)
	threads := mocks.NewMockThreadAccessChecker(t)
	uc := usecase.NewMomentUsecase(repo, threads)

	wantErr := errors.New("search index unavailable")
	repo.EXPECT().Search(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, wantErr)

	result, err := uc.SearchMoments(context.Background(), usecase.SearchMomentsInput{
		UserID: uuid.New(),
		Query:  "test",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}
