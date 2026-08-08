//go:build unit

package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func expectPassthroughCircleTransaction(repo *mocks.MockCircleRepository) {
	repo.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(usecase.CircleRepository) error) error {
			return fn(repo)
		})
}

func TestCircleUsecase_CreateCircle_SeatsCreatorAsAdmin(t *testing.T) {
	repo := mocks.NewMockCircleRepository(t)
	uc := usecase.NewCircleUsecase(repo)

	userID := uuid.New()
	created := &entity.Circle{ID: uuid.New(), Name: "NgodingVareng"}

	expectPassthroughCircleTransaction(repo)
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(c *entity.Circle) bool { return c.Name == "NgodingVareng" })).
		Return(created, nil)
	repo.EXPECT().
		AddCreatorAsAdmin(mock.Anything, created.ID, userID).
		Return(&entity.CircleMember{CircleID: created.ID, UserID: userID, Role: enum.CircleRoleAdmin}, nil)

	result, err := uc.CreateCircle(context.Background(), usecase.CreateCircleInput{UserID: userID, Name: "NgodingVareng"})

	assert.NoError(t, err)
	assert.Equal(t, created, result)
}

func TestCircleUsecase_CreateCircle_AdminSeatFailurePropagates(t *testing.T) {
	repo := mocks.NewMockCircleRepository(t)
	uc := usecase.NewCircleUsecase(repo)

	wantErr := errors.New("db exploded")
	created := &entity.Circle{ID: uuid.New()}

	expectPassthroughCircleTransaction(repo)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(created, nil)
	repo.EXPECT().AddCreatorAsAdmin(mock.Anything, created.ID, mock.Anything).Return(nil, wantErr)

	result, err := uc.CreateCircle(context.Background(), usecase.CreateCircleInput{UserID: uuid.New(), Name: "Test"})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}

func TestCircleUsecase_UpdateCircle_NotFound(t *testing.T) {
	repo := mocks.NewMockCircleRepository(t)
	uc := usecase.NewCircleUsecase(repo)

	repo.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil, errs.ErrNotFound)

	result, err := uc.UpdateCircle(context.Background(), usecase.UpdateCircleInput{
		ID: uuid.New(), UserID: uuid.New(), Name: "New name",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestCircleUsecase_DissolveCircle_Success(t *testing.T) {
	repo := mocks.NewMockCircleRepository(t)
	uc := usecase.NewCircleUsecase(repo)

	circleID, userID := uuid.New(), uuid.New()
	repo.EXPECT().Dissolve(mock.Anything, circleID, userID).Return(nil)

	err := uc.DissolveCircle(context.Background(), circleID, userID)

	assert.NoError(t, err)
}

func TestCircleUsecase_GetCircle_NotFound(t *testing.T) {
	repo := mocks.NewMockCircleRepository(t)
	uc := usecase.NewCircleUsecase(repo)

	circleID, userID := uuid.New(), uuid.New()
	repo.EXPECT().GetByID(mock.Anything, circleID, userID).Return(nil, errs.ErrNotFound)

	result, err := uc.GetCircle(context.Background(), circleID, userID)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestCircleUsecase_ListMembers_ChecksAccessFirst(t *testing.T) {
	repo := mocks.NewMockCircleRepository(t)
	uc := usecase.NewCircleUsecase(repo)

	circleID, userID := uuid.New(), uuid.New()
	repo.EXPECT().GetByID(mock.Anything, circleID, userID).Return(nil, errs.ErrNotFound)

	result, err := uc.ListMembers(context.Background(), circleID, userID)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
	// ListMembers must never be called once access resolution already failed.
	repo.AssertNotCalled(t, "ListMembers", mock.Anything, mock.Anything)
}

func TestCircleUsecase_ListMembers_Success(t *testing.T) {
	repo := mocks.NewMockCircleRepository(t)
	uc := usecase.NewCircleUsecase(repo)

	circleID, userID := uuid.New(), uuid.New()
	repo.EXPECT().GetByID(mock.Anything, circleID, userID).Return(&entity.Circle{ID: circleID}, nil)
	members := []*entity.CircleMember{{CircleID: circleID, UserID: userID, Role: enum.CircleRoleAdmin}}
	repo.EXPECT().ListMembers(mock.Anything, circleID).Return(members, nil)

	result, err := uc.ListMembers(context.Background(), circleID, userID)

	assert.NoError(t, err)
	assert.Equal(t, members, result)
}

func TestCircleUsecase_LeaveCircle_Success(t *testing.T) {
	repo := mocks.NewMockCircleRepository(t)
	uc := usecase.NewCircleUsecase(repo)

	circleID, userID := uuid.New(), uuid.New()
	repo.EXPECT().Leave(mock.Anything, circleID, userID).Return(nil)

	err := uc.LeaveCircle(context.Background(), circleID, userID)

	assert.NoError(t, err)
}

func TestCircleUsecase_RemoveMember_Success(t *testing.T) {
	repo := mocks.NewMockCircleRepository(t)
	uc := usecase.NewCircleUsecase(repo)

	circleID, targetUserID, removedByUserID := uuid.New(), uuid.New(), uuid.New()
	repo.EXPECT().RemoveMember(mock.Anything, circleID, targetUserID, removedByUserID).Return(nil)

	err := uc.RemoveMember(context.Background(), circleID, targetUserID, removedByUserID)

	assert.NoError(t, err)
}

func TestCircleUsecase_UpdateMemberRole_NotFound(t *testing.T) {
	repo := mocks.NewMockCircleRepository(t)
	uc := usecase.NewCircleUsecase(repo)

	repo.EXPECT().
		UpdateMemberRole(mock.Anything, mock.Anything, mock.Anything, mock.Anything, enum.CircleRoleAdmin).
		Return(nil, errs.ErrNotFound)

	result, err := uc.UpdateMemberRole(context.Background(), usecase.UpdateCircleMemberRoleInput{
		CircleID: uuid.New(), UserID: uuid.New(), ChangedByUserID: uuid.New(), Role: enum.CircleRoleAdmin,
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestCircleUsecase_UpdateMemberPermissions_Success(t *testing.T) {
	repo := mocks.NewMockCircleRepository(t)
	uc := usecase.NewCircleUsecase(repo)

	circleID, userID, changedBy := uuid.New(), uuid.New(), uuid.New()
	expected := &entity.CircleMember{CircleID: circleID, UserID: userID, CanInvite: true, CanCapture: false}
	repo.EXPECT().
		UpdateMemberPermissions(mock.Anything, circleID, userID, changedBy, true, false).
		Return(expected, nil)

	result, err := uc.UpdateMemberPermissions(context.Background(), usecase.UpdateCircleMemberPermissionsInput{
		CircleID: circleID, UserID: userID, ChangedByUserID: changedBy, CanInvite: true, CanCapture: false,
	})

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
