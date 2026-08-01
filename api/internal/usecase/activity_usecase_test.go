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

func expectPassthroughTransaction(repo *mocks.MockActivityRepository) {
	repo.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(usecase.ActivityRepository) error) error {
			return fn(repo)
		})
}

func TestActivityUsecase_CreateActivity_Success(t *testing.T) {
	repo := mocks.NewMockActivityRepository(t)
	uc := usecase.NewActivityUsecase(repo)

	userID := uuid.New()
	expected := &entity.Activity{
		ID:     uuid.New(),
		UserID: userID,
		Name:   "Morning cook",
	}

	expectPassthroughTransaction(repo)
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(e *entity.Activity) bool {
			return e.UserID == userID && e.Name == expected.Name
		})).
		Return(expected, nil)

	result, err := uc.CreateActivity(context.Background(), usecase.CreateActivityInput{
		UserID: userID,
		Name:   expected.Name,
	})

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestActivityUsecase_CreateActivity_RepositoryError(t *testing.T) {
	repo := mocks.NewMockActivityRepository(t)
	uc := usecase.NewActivityUsecase(repo)

	wantErr := errors.New("db exploded")

	// WithTransaction itself fails (e.g. BeginTx failed) — the callback
	// never runs, so a plain Return is enough here.
	repo.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		Return(wantErr)

	result, err := uc.CreateActivity(context.Background(), usecase.CreateActivityInput{
		UserID: uuid.New(),
		Name:   "Test",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}

func TestActivityUsecase_CreateActivity_ErrorInsideTransactionPropagates(t *testing.T) {
	repo := mocks.NewMockActivityRepository(t)
	uc := usecase.NewActivityUsecase(repo)

	wantErr := errors.New("unique constraint violated")

	// This time WithTransaction itself "succeeds" in the sense of
	// running the callback, but the callback's own Create call fails —
	// RunAndReturn correctly propagates that, unlike a hardcoded
	// .Return(nil) would.
	expectPassthroughTransaction(repo)
	repo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(nil, wantErr)

	result, err := uc.CreateActivity(context.Background(), usecase.CreateActivityInput{
		UserID: uuid.New(),
		Name:   "Test",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}

func TestActivityUsecase_CreateActivity_DefaultConfirmationTimeout(t *testing.T) {
	repo := mocks.NewMockActivityRepository(t)
	uc := usecase.NewActivityUsecase(repo)

	var captured *entity.Activity

	expectPassthroughTransaction(repo)
	repo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Run(func(ctx context.Context, activity *entity.Activity) {
			captured = activity
		}).
		Return(&entity.Activity{}, nil)

	_, err := uc.CreateActivity(context.Background(), usecase.CreateActivityInput{
		UserID: uuid.New(),
		Name:   "Test",
		// ConfirmationTimeoutMinutes intentionally omitted (nil), to
		// verify the usecase applies the 1440-minute default itself —
		// see the comment on defaultConfirmationTimeoutMinutes for why
		// this can't just be left to the DB's own column DEFAULT.
	})

	assert.NoError(t, err)
	if assert.NotNil(t, captured.ConfirmationTimeoutMinutes) {
		assert.EqualValues(t, 1440, *captured.ConfirmationTimeoutMinutes)
	}
}

func TestActivityUsecase_CreateActivity_ExplicitConfirmationTimeout(t *testing.T) {
	repo := mocks.NewMockActivityRepository(t)
	uc := usecase.NewActivityUsecase(repo)

	customTimeout := int32(60)
	var captured *entity.Activity

	expectPassthroughTransaction(repo)
	repo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Run(func(ctx context.Context, activity *entity.Activity) {
			captured = activity
		}).
		Return(&entity.Activity{}, nil)

	_, err := uc.CreateActivity(context.Background(), usecase.CreateActivityInput{
		UserID:                     uuid.New(),
		Name:                       "Test",
		ConfirmationTimeoutMinutes: &customTimeout,
	})

	assert.NoError(t, err)
	if assert.NotNil(t, captured.ConfirmationTimeoutMinutes) {
		assert.EqualValues(t, 60, *captured.ConfirmationTimeoutMinutes)
	}
}

// --- UpdateActivity ---

func TestActivityUsecase_UpdateActivity_Success(t *testing.T) {
	repo := mocks.NewMockActivityRepository(t)
	uc := usecase.NewActivityUsecase(repo)

	activityID := uuid.New()
	userID := uuid.New()
	expected := &entity.Activity{ID: activityID, UserID: userID, Name: "Olahraga sore"}

	expectPassthroughTransaction(repo)
	repo.EXPECT().
		Update(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			return a.ID == activityID && a.UserID == userID && a.Name == "Olahraga sore"
		})).
		Return(expected, nil)

	result, err := uc.UpdateActivity(context.Background(), usecase.UpdateActivityInput{
		ID:     activityID,
		UserID: userID,
		Name:   "Olahraga sore",
	})

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestActivityUsecase_UpdateActivity_NotFound(t *testing.T) {
	repo := mocks.NewMockActivityRepository(t)
	uc := usecase.NewActivityUsecase(repo)

	expectPassthroughTransaction(repo)
	repo.EXPECT().
		Update(mock.Anything, mock.Anything).
		Return(nil, errs.ErrNotFound)

	result, err := uc.UpdateActivity(context.Background(), usecase.UpdateActivityInput{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Name:   "Test",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestActivityUsecase_UpdateActivity_DefaultsAppliedLikeCreate(t *testing.T) {
	repo := mocks.NewMockActivityRepository(t)
	uc := usecase.NewActivityUsecase(repo)

	var captured *entity.Activity

	expectPassthroughTransaction(repo)
	repo.EXPECT().
		Update(mock.Anything, mock.Anything).
		Run(func(ctx context.Context, activity *entity.Activity) {
			captured = activity
		}).
		Return(&entity.Activity{}, nil)

	_, err := uc.UpdateActivity(context.Background(), usecase.UpdateActivityInput{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Name:   "Test",
		// ColorHex and ConfirmationTimeoutMinutes intentionally omitted
		// (nil) — same defaulting behavior as CreateActivity.
	})

	assert.NoError(t, err)
	if assert.NotNil(t, captured.ConfirmationTimeoutMinutes) {
		assert.EqualValues(t, 1440, *captured.ConfirmationTimeoutMinutes)
	}
	if assert.NotNil(t, captured.ColorHex) {
		assert.Equal(t, "#374151", *captured.ColorHex)
	}
}

// --- DeleteActivity ---

func TestActivityUsecase_SoftDeleteActivity_Success(t *testing.T) {
	repo := mocks.NewMockActivityRepository(t)
	uc := usecase.NewActivityUsecase(repo)

	activityID := uuid.New()
	userID := uuid.New()

	expectPassthroughTransaction(repo)
	repo.EXPECT().SoftDelete(mock.Anything, activityID, userID).Return(nil)

	err := uc.SoftDeleteActivity(context.Background(), activityID, userID)

	assert.NoError(t, err)
}

func TestActivityUsecase_SoftDeleteActivity_RepositoryError(t *testing.T) {
	repo := mocks.NewMockActivityRepository(t)
	uc := usecase.NewActivityUsecase(repo)

	wantErr := errors.New("db exploded")

	expectPassthroughTransaction(repo)
	repo.EXPECT().SoftDelete(mock.Anything, mock.Anything, mock.Anything).Return(wantErr)

	err := uc.SoftDeleteActivity(context.Background(), uuid.New(), uuid.New())

	assert.ErrorIs(t, err, wantErr)
}
