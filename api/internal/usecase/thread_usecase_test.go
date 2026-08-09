//go:build unit

package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func expectPassthroughTransaction(repo *mocks.MockThreadRepository) {
	repo.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(usecase.ThreadRepository) error) error {
			return fn(repo)
		})
}

func TestThreadUsecase_CreateThread_Success(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	userID := uuid.New()
	expected := &entity.Thread{
		ID:     uuid.New(),
		UserID: userID,
		Name:   "Morning cook",
	}

	expectPassthroughTransaction(repo)
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(e *entity.Thread) bool {
			return e.UserID == userID && e.Name == expected.Name
		})).
		Return(expected, nil)

	result, err := uc.CreateThread(context.Background(), usecase.CreateThreadInput{
		UserID: userID,
		Name:   expected.Name,
	})

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestThreadUsecase_CreateThread_RepositoryError(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	wantErr := errors.New("db exploded")

	// WithTransaction itself fails (e.g. BeginTx failed) — the callback
	// never runs, so a plain Return is enough here.
	repo.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		Return(wantErr)

	result, err := uc.CreateThread(context.Background(), usecase.CreateThreadInput{
		UserID: uuid.New(),
		Name:   "Test",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}

func TestThreadUsecase_CreateThread_ErrorInsideTransactionPropagates(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	wantErr := errors.New("unique constraint violated")

	// This time WithTransaction itself "succeeds" in the sense of
	// running the callback, but the callback's own Create call fails —
	// RunAndReturn correctly propagates that, unlike a hardcoded
	// .Return(nil) would.
	expectPassthroughTransaction(repo)
	repo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(nil, wantErr)

	result, err := uc.CreateThread(context.Background(), usecase.CreateThreadInput{
		UserID: uuid.New(),
		Name:   "Test",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}

func TestThreadUsecase_CreateThread_DefaultColorHex(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	var captured *entity.Thread

	expectPassthroughTransaction(repo)
	repo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Run(func(ctx context.Context, thread *entity.Thread) {
			captured = thread
		}).
		Return(&entity.Thread{}, nil)

	_, err := uc.CreateThread(context.Background(), usecase.CreateThreadInput{
		UserID: uuid.New(),
		Name:   "Test",
		// ColorHex intentionally omitted (nil), to verify the usecase
		// applies the default gray itself.
	})

	assert.NoError(t, err)
	if assert.NotNil(t, captured.ColorHex) {
		assert.Equal(t, "#374151", *captured.ColorHex)
	}
}

func TestThreadUsecase_CreateThread_ExplicitColorHex(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	customColor := "#FF5733"
	var captured *entity.Thread

	expectPassthroughTransaction(repo)
	repo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Run(func(ctx context.Context, thread *entity.Thread) {
			captured = thread
		}).
		Return(&entity.Thread{}, nil)

	_, err := uc.CreateThread(context.Background(), usecase.CreateThreadInput{
		UserID:   uuid.New(),
		Name:     "Test",
		ColorHex: &customColor,
	})

	assert.NoError(t, err)
	if assert.NotNil(t, captured.ColorHex) {
		assert.Equal(t, "#FF5733", *captured.ColorHex)
	}
}

func TestThreadUsecase_CreateThread_CircleID_ActiveMember_Success(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	circles := mocks.NewMockCircleAccessChecker(t)
	uc := usecase.NewThreadUsecase(repo, circles)

	userID := uuid.New()
	circleID := uuid.New()
	expected := &entity.Thread{ID: uuid.New(), UserID: userID, CircleID: &circleID, Name: "Team standup"}

	circles.EXPECT().
		GetActiveMember(mock.Anything, circleID, userID).
		Return(&entity.CircleMember{CircleID: circleID, UserID: userID}, nil)

	expectPassthroughTransaction(repo)
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(e *entity.Thread) bool {
			return e.CircleID != nil && *e.CircleID == circleID
		})).
		Return(expected, nil)

	result, err := uc.CreateThread(context.Background(), usecase.CreateThreadInput{
		UserID:   userID,
		CircleID: &circleID,
		Name:     expected.Name,
	})

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestThreadUsecase_CreateThread_CircleID_NotAMember_Rejected(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	circles := mocks.NewMockCircleAccessChecker(t)
	uc := usecase.NewThreadUsecase(repo, circles)

	userID := uuid.New()
	circleID := uuid.New()

	circles.EXPECT().
		GetActiveMember(mock.Anything, circleID, userID).
		Return(nil, errs.ErrNotFound)

	// repo.Create must never be reached — no expectation set on repo at
	// all, so mockery's t.Cleanup assertion fails the test if it is.
	result, err := uc.CreateThread(context.Background(), usecase.CreateThreadInput{
		UserID:   userID,
		CircleID: &circleID,
		Name:     "Team standup",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

// --- ListCircleThreads ---

func TestThreadUsecase_ListCircleThreads_ActiveMember_Success(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	circles := mocks.NewMockCircleAccessChecker(t)
	uc := usecase.NewThreadUsecase(repo, circles)

	userID := uuid.New()
	circleID := uuid.New()
	expected := []*entity.Thread{{ID: uuid.New(), CircleID: &circleID, Name: "A"}}

	circles.EXPECT().
		GetActiveMember(mock.Anything, circleID, userID).
		Return(&entity.CircleMember{CircleID: circleID, UserID: userID}, nil)
	repo.EXPECT().ListByCircle(mock.Anything, circleID).Return(expected, nil)

	result, err := uc.ListCircleThreads(context.Background(), circleID, userID)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestThreadUsecase_ListCircleThreads_NotAMember_Rejected(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	circles := mocks.NewMockCircleAccessChecker(t)
	uc := usecase.NewThreadUsecase(repo, circles)

	userID := uuid.New()
	circleID := uuid.New()

	circles.EXPECT().
		GetActiveMember(mock.Anything, circleID, userID).
		Return(nil, errs.ErrNotFound)

	result, err := uc.ListCircleThreads(context.Background(), circleID, userID)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

// --- UpdateThread ---

func TestThreadUsecase_UpdateThread_Success(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	threadID := uuid.New()
	userID := uuid.New()
	expected := &entity.Thread{ID: threadID, UserID: userID, Name: "Olahraga sore"}

	expectPassthroughTransaction(repo)
	repo.EXPECT().
		Update(mock.Anything, mock.MatchedBy(func(a *entity.Thread) bool {
			return a.ID == threadID && a.UserID == userID && a.Name == "Olahraga sore"
		})).
		Return(expected, nil)

	result, err := uc.UpdateThread(context.Background(), usecase.UpdateThreadInput{
		ID:     threadID,
		UserID: userID,
		Name:   "Olahraga sore",
	})

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestThreadUsecase_UpdateThread_NotFound(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	expectPassthroughTransaction(repo)
	repo.EXPECT().
		Update(mock.Anything, mock.Anything).
		Return(nil, errs.ErrNotFound)

	result, err := uc.UpdateThread(context.Background(), usecase.UpdateThreadInput{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Name:   "Test",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestThreadUsecase_UpdateThread_DefaultsAppliedLikeCreate(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	var captured *entity.Thread

	expectPassthroughTransaction(repo)
	repo.EXPECT().
		Update(mock.Anything, mock.Anything).
		Run(func(ctx context.Context, thread *entity.Thread) {
			captured = thread
		}).
		Return(&entity.Thread{}, nil)

	_, err := uc.UpdateThread(context.Background(), usecase.UpdateThreadInput{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Name:   "Test",
		// ColorHex intentionally omitted (nil) — same defaulting
		// behavior as CreateThread.
	})

	assert.NoError(t, err)
	if assert.NotNil(t, captured.ColorHex) {
		assert.Equal(t, "#374151", *captured.ColorHex)
	}
}

// --- DeleteThread ---

func TestThreadUsecase_SoftDeleteThread_Success(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	threadID := uuid.New()
	userID := uuid.New()

	expectPassthroughTransaction(repo)
	repo.EXPECT().SoftDelete(mock.Anything, threadID, userID).Return(nil)

	err := uc.SoftDeleteThread(context.Background(), threadID, userID)

	assert.NoError(t, err)
}

func TestThreadUsecase_SoftDeleteThread_RepositoryError(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	wantErr := errors.New("db exploded")

	expectPassthroughTransaction(repo)
	repo.EXPECT().SoftDelete(mock.Anything, mock.Anything, mock.Anything).Return(wantErr)

	err := uc.SoftDeleteThread(context.Background(), uuid.New(), uuid.New())

	assert.ErrorIs(t, err, wantErr)
}

// --- GetThread ---

func TestThreadUsecase_GetThread_Success(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	userID := uuid.New()
	threadID := uuid.New()
	expected := &entity.Thread{ID: threadID, UserID: userID, Name: "Morning run"}

	repo.EXPECT().GetByID(mock.Anything, threadID, userID).Return(expected, nil)

	result, err := uc.GetThread(context.Background(), userID, threadID)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestThreadUsecase_GetThread_NotFound(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	userID, threadID := uuid.New(), uuid.New()
	repo.EXPECT().GetByID(mock.Anything, threadID, userID).Return(nil, errs.ErrNotFound)

	result, err := uc.GetThread(context.Background(), userID, threadID)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

// --- SearchThreads ---

func TestThreadUsecase_SearchThreads_DefaultsPageAndPageSize(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	var captured usecase.SearchThreadsParams
	repo.EXPECT().
		Search(mock.Anything, mock.Anything).
		Run(func(ctx context.Context, params usecase.SearchThreadsParams) {
			captured = params
		}).
		Return([]*entity.Thread{}, int64(0), nil)

	result, err := uc.SearchThreads(context.Background(), usecase.SearchThreadsInput{
		UserID: uuid.New(),
		// Page and PageSize intentionally omitted (zero value) to verify
		// the usecase applies its own defaults (page=1, page_size=20).
	})

	assert.NoError(t, err)
	assert.EqualValues(t, 1, result.Page)
	assert.EqualValues(t, 20, result.PageSize)
	assert.EqualValues(t, 0, captured.Offset)
	assert.EqualValues(t, 20, captured.Limit)
}

func TestThreadUsecase_SearchThreads_PageSizeCappedAtMax(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	var captured usecase.SearchThreadsParams
	repo.EXPECT().
		Search(mock.Anything, mock.Anything).
		Run(func(ctx context.Context, params usecase.SearchThreadsParams) {
			captured = params
		}).
		Return([]*entity.Thread{}, int64(0), nil)

	result, err := uc.SearchThreads(context.Background(), usecase.SearchThreadsInput{
		UserID:   uuid.New(),
		PageSize: 500, // way above the 100 cap
	})

	assert.NoError(t, err)
	assert.EqualValues(t, 100, result.PageSize)
	assert.EqualValues(t, 100, captured.Limit)
}

func TestThreadUsecase_SearchThreads_ComputesOffsetFromPage(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	var captured usecase.SearchThreadsParams
	repo.EXPECT().
		Search(mock.Anything, mock.Anything).
		Run(func(ctx context.Context, params usecase.SearchThreadsParams) {
			captured = params
		}).
		Return([]*entity.Thread{}, int64(0), nil)

	_, err := uc.SearchThreads(context.Background(), usecase.SearchThreadsInput{
		UserID:   uuid.New(),
		Page:     3,
		PageSize: 10,
	})

	assert.NoError(t, err)
	assert.EqualValues(t, 20, captured.Offset) // (page-1) * pageSize = 2*10
	assert.EqualValues(t, 10, captured.Limit)
}

func TestThreadUsecase_SearchThreads_PassesFiltersThrough(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	userID := uuid.New()
	name := "run"
	archived := true

	repo.EXPECT().
		Search(mock.Anything, mock.MatchedBy(func(p usecase.SearchThreadsParams) bool {
			return p.UserID == userID &&
				p.Name != nil && *p.Name == name &&
				p.Archived != nil && *p.Archived == archived
		})).
		Return([]*entity.Thread{}, int64(0), nil)

	_, err := uc.SearchThreads(context.Background(), usecase.SearchThreadsInput{
		UserID:   userID,
		Name:     &name,
		Archived: &archived,
	})

	assert.NoError(t, err)
}

func TestThreadUsecase_SearchThreads_RepositoryError(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	wantErr := errors.New("db exploded")
	repo.EXPECT().Search(mock.Anything, mock.Anything).Return(nil, int64(0), wantErr)

	result, err := uc.SearchThreads(context.Background(), usecase.SearchThreadsInput{UserID: uuid.New()})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}

func TestThreadUsecase_SearchThreads_ReturnsTotalAndResults(t *testing.T) {
	repo := mocks.NewMockThreadRepository(t)
	uc := usecase.NewThreadUsecase(repo, mocks.NewMockCircleAccessChecker(t))

	threads := []*entity.Thread{{ID: uuid.New(), Name: "A"}, {ID: uuid.New(), Name: "B"}}
	repo.EXPECT().Search(mock.Anything, mock.Anything).Return(threads, int64(42), nil)

	result, err := uc.SearchThreads(context.Background(), usecase.SearchThreadsInput{UserID: uuid.New()})

	assert.NoError(t, err)
	assert.Len(t, result.Threads, 2)
	assert.EqualValues(t, 42, result.Total)
}
