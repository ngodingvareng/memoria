//go:build unit

package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func expectPassthroughCircleJoinRequestTransaction(repo *mocks.MockCircleJoinRequestRepository) {
	repo.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(usecase.CircleJoinRequestRepository) error) error {
			return fn(repo)
		})
}

func TestCircleJoinRequestUsecase_FollowInviteLink_NoApprovalRequired_JoinsImmediately(t *testing.T) {
	repo := mocks.NewMockCircleJoinRequestRepository(t)
	circles := mocks.NewMockCircleAccessChecker(t)
	links := mocks.NewMockCircleInviteLinkResolver(t)
	tokens := mocks.NewMockRefreshTokenGenerator(t)
	uc := usecase.NewCircleJoinRequestUsecase(repo, circles, links, tokens)

	circleID, userID := uuid.New(), uuid.New()
	requiresApproval := false
	tokens.EXPECT().Hash("raw-token").Return("hashed-token")
	links.EXPECT().GetActiveLinkByTokenHash(mock.Anything, "hashed-token").
		Return(&entity.CircleInvite{CircleID: circleID, RequiresApproval: &requiresApproval}, nil)
	repo.EXPECT().AddMembers(mock.Anything, circleID, []uuid.UUID{userID}).
		Return([]*entity.CircleMember{{CircleID: circleID, UserID: userID}}, nil)

	result, err := uc.FollowInviteLink(context.Background(), "raw-token", userID)

	assert.NoError(t, err)
	assert.NotNil(t, result.Joined)
	assert.Nil(t, result.JoinRequest)
}

func TestCircleJoinRequestUsecase_FollowInviteLink_ApprovalRequired_CreatesRequest(t *testing.T) {
	repo := mocks.NewMockCircleJoinRequestRepository(t)
	circles := mocks.NewMockCircleAccessChecker(t)
	links := mocks.NewMockCircleInviteLinkResolver(t)
	tokens := mocks.NewMockRefreshTokenGenerator(t)
	uc := usecase.NewCircleJoinRequestUsecase(repo, circles, links, tokens)

	circleID, userID, inviteID := uuid.New(), uuid.New(), uuid.New()
	requiresApproval := true
	tokens.EXPECT().Hash("raw-token").Return("hashed-token")
	links.EXPECT().GetActiveLinkByTokenHash(mock.Anything, "hashed-token").
		Return(&entity.CircleInvite{ID: inviteID, CircleID: circleID, RequiresApproval: &requiresApproval}, nil)
	repo.EXPECT().GetPendingByUser(mock.Anything, circleID, userID).Return(nil, errs.ErrNotFound)
	repo.EXPECT().Create(mock.Anything, circleID, userID, &inviteID).
		Return(&entity.CircleJoinRequest{CircleID: circleID, UserID: userID, InviteID: &inviteID}, nil)

	result, err := uc.FollowInviteLink(context.Background(), "raw-token", userID)

	assert.NoError(t, err)
	assert.Nil(t, result.Joined)
	assert.NotNil(t, result.JoinRequest)
}

func TestCircleJoinRequestUsecase_FollowInviteLink_ApprovalRequired_AlreadyPending_Idempotent(t *testing.T) {
	repo := mocks.NewMockCircleJoinRequestRepository(t)
	circles := mocks.NewMockCircleAccessChecker(t)
	links := mocks.NewMockCircleInviteLinkResolver(t)
	tokens := mocks.NewMockRefreshTokenGenerator(t)
	uc := usecase.NewCircleJoinRequestUsecase(repo, circles, links, tokens)

	circleID, userID := uuid.New(), uuid.New()
	requiresApproval := true
	existing := &entity.CircleJoinRequest{CircleID: circleID, UserID: userID}
	tokens.EXPECT().Hash("raw-token").Return("hashed-token")
	links.EXPECT().GetActiveLinkByTokenHash(mock.Anything, "hashed-token").
		Return(&entity.CircleInvite{CircleID: circleID, RequiresApproval: &requiresApproval}, nil)
	repo.EXPECT().GetPendingByUser(mock.Anything, circleID, userID).Return(existing, nil)

	result, err := uc.FollowInviteLink(context.Background(), "raw-token", userID)

	assert.NoError(t, err)
	assert.Equal(t, existing, result.JoinRequest)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestCircleJoinRequestUsecase_ListPending_RequiresCanInvite(t *testing.T) {
	repo := mocks.NewMockCircleJoinRequestRepository(t)
	circles := mocks.NewMockCircleAccessChecker(t)
	links := mocks.NewMockCircleInviteLinkResolver(t)
	tokens := mocks.NewMockRefreshTokenGenerator(t)
	uc := usecase.NewCircleJoinRequestUsecase(repo, circles, links, tokens)

	circleID, userID := uuid.New(), uuid.New()
	circles.EXPECT().GetActiveMember(mock.Anything, circleID, userID).Return(&entity.CircleMember{CanInvite: false}, nil)

	result, err := uc.ListPending(context.Background(), circleID, userID, 1, 20)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrInsufficientPermission)
}

func TestCircleJoinRequestUsecase_Approve_SeatsMemberInSameTransaction(t *testing.T) {
	repo := mocks.NewMockCircleJoinRequestRepository(t)
	circles := mocks.NewMockCircleAccessChecker(t)
	links := mocks.NewMockCircleInviteLinkResolver(t)
	tokens := mocks.NewMockRefreshTokenGenerator(t)
	uc := usecase.NewCircleJoinRequestUsecase(repo, circles, links, tokens)

	circleID, decidedBy, requestID, requesterID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	circles.EXPECT().GetActiveMember(mock.Anything, circleID, decidedBy).Return(&entity.CircleMember{CanInvite: true}, nil)
	expectPassthroughCircleJoinRequestTransaction(repo)
	repo.EXPECT().Approve(mock.Anything, requestID, circleID, decidedBy).
		Return(&entity.CircleJoinRequest{ID: requestID, CircleID: circleID, UserID: requesterID}, nil)
	repo.EXPECT().AddMembers(mock.Anything, circleID, []uuid.UUID{requesterID}).
		Return([]*entity.CircleMember{{CircleID: circleID, UserID: requesterID}}, nil)

	member, err := uc.Approve(context.Background(), requestID, circleID, decidedBy)

	assert.NoError(t, err)
	assert.Equal(t, requesterID, member.UserID)
}

func TestCircleJoinRequestUsecase_Cancel_Success(t *testing.T) {
	repo := mocks.NewMockCircleJoinRequestRepository(t)
	circles := mocks.NewMockCircleAccessChecker(t)
	links := mocks.NewMockCircleInviteLinkResolver(t)
	tokens := mocks.NewMockRefreshTokenGenerator(t)
	uc := usecase.NewCircleJoinRequestUsecase(repo, circles, links, tokens)

	requestID, userID := uuid.New(), uuid.New()
	repo.EXPECT().Cancel(mock.Anything, requestID, userID).
		Return(&entity.CircleJoinRequest{ID: requestID, UserID: userID}, nil)

	request, err := uc.Cancel(context.Background(), requestID, userID)

	assert.NoError(t, err)
	assert.Equal(t, requestID, request.ID)
}
