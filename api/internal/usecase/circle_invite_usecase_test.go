//go:build unit

package usecase_test

import (
	"context"
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

func expectPassthroughCircleInviteTransaction(repo *mocks.MockCircleInviteRepository) {
	repo.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(usecase.CircleInviteRepository) error) error {
			return fn(repo)
		})
}

func newCircleInviteUsecaseDeps(t *testing.T) (*mocks.MockCircleInviteRepository, *mocks.MockCircleAccessChecker, *mocks.MockUserPolicyReader, *mocks.MockUserBlockChecker, *mocks.MockRefreshTokenGenerator) {
	t.Helper()
	return mocks.NewMockCircleInviteRepository(t),
		mocks.NewMockCircleAccessChecker(t),
		mocks.NewMockUserPolicyReader(t),
		mocks.NewMockUserBlockChecker(t),
		mocks.NewMockRefreshTokenGenerator(t)
}

func TestCircleInviteUsecase_InviteDirect_NotAllowedToInvite(t *testing.T) {
	repo, circles, users, blocks, tokens := newCircleInviteUsecaseDeps(t)
	uc := usecase.NewCircleInviteUsecase(repo, circles, users, blocks, tokens)

	circleID, userID := uuid.New(), uuid.New()
	circles.EXPECT().GetActiveMember(mock.Anything, circleID, userID).
		Return(&entity.CircleMember{CircleID: circleID, UserID: userID, CanInvite: false}, nil)

	result, err := uc.InviteDirect(context.Background(), usecase.InviteDirectInput{
		CircleID: circleID, InvitedByUserID: userID, Usernames: []string{"gede"},
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, errs.ErrInsufficientPermission)
}

func TestCircleInviteUsecase_InviteDirect_SplitsByPolicy(t *testing.T) {
	repo, circles, users, blocks, tokens := newCircleInviteUsecaseDeps(t)
	uc := usecase.NewCircleInviteUsecase(repo, circles, users, blocks, tokens)

	circleID, userID := uuid.New(), uuid.New()
	openUser := &entity.User{ID: uuid.New(), Username: strPtr("open"), DiscoverableByUsername: true, CircleInvitePolicy: enum.AudiencePolicyAnyone}
	restrictedUser := &entity.User{ID: uuid.New(), Username: strPtr("restricted"), DiscoverableByUsername: true, CircleInvitePolicy: enum.AudiencePolicyKnown}

	circles.EXPECT().GetActiveMember(mock.Anything, circleID, userID).
		Return(&entity.CircleMember{CircleID: circleID, UserID: userID, CanInvite: true}, nil)
	users.EXPECT().GetByUsername(mock.Anything, "open").Return(openUser, nil)
	users.EXPECT().GetByUsername(mock.Anything, "restricted").Return(restrictedUser, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, userID, openUser.ID).Return(false, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, userID, restrictedUser.ID).Return(false, nil)

	expectPassthroughCircleInviteTransaction(repo)
	repo.EXPECT().
		AddMembers(mock.Anything, circleID, []uuid.UUID{openUser.ID}).
		Return([]*entity.CircleMember{{CircleID: circleID, UserID: openUser.ID}}, nil)
	repo.EXPECT().
		CreateUsernameInvite(mock.Anything, circleID, userID, restrictedUser.ID, mock.Anything).
		Return(&entity.CircleInvite{CircleID: circleID, InviteeUserID: &restrictedUser.ID}, nil)

	result, err := uc.InviteDirect(context.Background(), usecase.InviteDirectInput{
		CircleID: circleID, InvitedByUserID: userID, Usernames: []string{"open", "restricted"},
	})

	assert.NoError(t, err)
	assert.Len(t, result.AddedMembers, 1)
	assert.Len(t, result.PendingInvites, 1)
	assert.Empty(t, result.SkippedUsernames)
}

func TestCircleInviteUsecase_InviteDirect_SkipsBlockedAndUnknownUsernames(t *testing.T) {
	repo, circles, users, blocks, tokens := newCircleInviteUsecaseDeps(t)
	uc := usecase.NewCircleInviteUsecase(repo, circles, users, blocks, tokens)

	circleID, userID := uuid.New(), uuid.New()
	blockedUser := &entity.User{ID: uuid.New(), Username: strPtr("blocked"), DiscoverableByUsername: true, CircleInvitePolicy: enum.AudiencePolicyAnyone}

	circles.EXPECT().GetActiveMember(mock.Anything, circleID, userID).
		Return(&entity.CircleMember{CircleID: circleID, UserID: userID, CanInvite: true}, nil)
	users.EXPECT().GetByUsername(mock.Anything, "ghost").Return(nil, errs.ErrNotFound)
	users.EXPECT().GetByUsername(mock.Anything, "blocked").Return(blockedUser, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, userID, blockedUser.ID).Return(true, nil)

	expectPassthroughCircleInviteTransaction(repo)

	result, err := uc.InviteDirect(context.Background(), usecase.InviteDirectInput{
		CircleID: circleID, InvitedByUserID: userID, Usernames: []string{"ghost", "blocked"},
	})

	assert.NoError(t, err)
	assert.Empty(t, result.AddedMembers)
	assert.Empty(t, result.PendingInvites)
	assert.ElementsMatch(t, []string{"ghost", "blocked"}, result.SkippedUsernames)
}

func TestCircleInviteUsecase_AcceptInvite_Success(t *testing.T) {
	repo, circles, users, blocks, tokens := newCircleInviteUsecaseDeps(t)
	uc := usecase.NewCircleInviteUsecase(repo, circles, users, blocks, tokens)

	inviteID, userID, circleID := uuid.New(), uuid.New(), uuid.New()
	expectPassthroughCircleInviteTransaction(repo)
	repo.EXPECT().AcceptUsernameInvite(mock.Anything, inviteID, userID).
		Return(&entity.CircleInvite{ID: inviteID, CircleID: circleID}, nil)
	repo.EXPECT().AddMembers(mock.Anything, circleID, []uuid.UUID{userID}).
		Return([]*entity.CircleMember{{CircleID: circleID, UserID: userID}}, nil)

	member, err := uc.AcceptInvite(context.Background(), inviteID, userID)

	assert.NoError(t, err)
	assert.Equal(t, userID, member.UserID)
}

func TestCircleInviteUsecase_AcceptInvite_AlreadyAnswered_NotFound(t *testing.T) {
	repo, circles, users, blocks, tokens := newCircleInviteUsecaseDeps(t)
	uc := usecase.NewCircleInviteUsecase(repo, circles, users, blocks, tokens)

	expectPassthroughCircleInviteTransaction(repo)
	repo.EXPECT().AcceptUsernameInvite(mock.Anything, mock.Anything, mock.Anything).Return(nil, errs.ErrNotFound)

	member, err := uc.AcceptInvite(context.Background(), uuid.New(), uuid.New())

	assert.Nil(t, member)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestCircleInviteUsecase_RevokeInvite_RequiresCanInvite(t *testing.T) {
	repo, circles, users, blocks, tokens := newCircleInviteUsecaseDeps(t)
	uc := usecase.NewCircleInviteUsecase(repo, circles, users, blocks, tokens)

	circleID, userID := uuid.New(), uuid.New()
	circles.EXPECT().GetActiveMember(mock.Anything, circleID, userID).
		Return(&entity.CircleMember{CanInvite: false}, nil)

	invite, err := uc.RevokeInvite(context.Background(), uuid.New(), circleID, userID)

	assert.Nil(t, invite)
	assert.ErrorIs(t, err, errs.ErrInsufficientPermission)
}

func TestCircleInviteUsecase_CreateOrRotateInviteLink_Success(t *testing.T) {
	repo, circles, users, blocks, tokens := newCircleInviteUsecaseDeps(t)
	uc := usecase.NewCircleInviteUsecase(repo, circles, users, blocks, tokens)

	circleID, userID := uuid.New(), uuid.New()
	circles.EXPECT().GetActiveMember(mock.Anything, circleID, userID).
		Return(&entity.CircleMember{CanInvite: true}, nil)
	tokens.EXPECT().Generate().Return("raw-token", nil)
	tokens.EXPECT().Hash("raw-token").Return("hashed-token")

	expectPassthroughCircleInviteTransaction(repo)
	repo.EXPECT().RevokeActiveLink(mock.Anything, circleID).Return(nil, nil)
	repo.EXPECT().
		CreateLink(mock.Anything, circleID, userID, "hashed-token", true).
		Return(&entity.CircleInvite{CircleID: circleID}, nil)

	result, err := uc.CreateOrRotateInviteLink(context.Background(), usecase.CreateOrRotateInviteLinkInput{
		CircleID: circleID, UserID: userID, RequiresApproval: true,
	})

	assert.NoError(t, err)
	assert.Equal(t, "raw-token", result.RawToken)
}
