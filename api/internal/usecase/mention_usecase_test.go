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

func newMentionUsecaseDeps(t *testing.T) (*mocks.MockMentionRepository, *mocks.MockMomentAccessChecker, *mocks.MockCircleAccessChecker, *mocks.MockCircleShareChecker, *mocks.MockUserPolicyReader, *mocks.MockUserKnownChecker, *mocks.MockUserBlockChecker, *mocks.MockUserSearcher, *mocks.MockNotificationCreator) {
	t.Helper()
	return mocks.NewMockMentionRepository(t),
		mocks.NewMockMomentAccessChecker(t),
		mocks.NewMockCircleAccessChecker(t),
		mocks.NewMockCircleShareChecker(t),
		mocks.NewMockUserPolicyReader(t),
		mocks.NewMockUserKnownChecker(t),
		mocks.NewMockUserBlockChecker(t),
		mocks.NewMockUserSearcher(t),
		mocks.NewMockNotificationCreator(t)
}

func TestMentionUsecase_CreateMention_AnyonePolicy_Success(t *testing.T) {
	repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications := newMentionUsecaseDeps(t)
	uc := usecase.NewMentionUsecase(repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications)

	momentID, ownerID, targetID := uuid.New(), uuid.New(), uuid.New()
	target := &entity.User{ID: targetID, Name: "Gede", Username: strPtr("gede"), MentionPolicy: enum.AudiencePolicyAnyone}

	moments.EXPECT().GetByID(mock.Anything, momentID, ownerID).Return(&entity.Moment{ID: momentID, UserID: ownerID}, nil)
	users.EXPECT().GetByUsername(mock.Anything, "gede").Return(target, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, ownerID, targetID).Return(false, nil)
	repo.EXPECT().Create(mock.Anything, momentID, targetID, "Gede").
		Return(&entity.MomentMention{MomentID: momentID, MentionedUserID: &targetID, DisplayName: "Gede"}, nil)
	notifications.EXPECT().
		CreateNotification(mock.Anything, mock.MatchedBy(func(input usecase.CreateNotificationInput) bool {
			return input.UserID == targetID && input.Kind == enum.NotificationKindMentionedInMoment
		})).
		Return(nil, nil)
	sharedCircleID := uuid.New()
	circleShares.EXPECT().ListSharedCircleIDs(mock.Anything, ownerID, targetID).
		Return([]uuid.UUID{sharedCircleID}, nil)

	result, err := uc.CreateMention(context.Background(), usecase.CreateMentionInput{
		MomentID: momentID, OwnerUserID: ownerID, Username: "gede",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Gede", result.Mention.DisplayName)
	assert.Equal(t, []uuid.UUID{sharedCircleID}, result.SharedCircleIDs)
}

func TestMentionUsecase_CreateMention_Blocked_Denied(t *testing.T) {
	repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications := newMentionUsecaseDeps(t)
	uc := usecase.NewMentionUsecase(repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications)

	momentID, ownerID, targetID := uuid.New(), uuid.New(), uuid.New()
	target := &entity.User{ID: targetID, MentionPolicy: enum.AudiencePolicyAnyone}

	moments.EXPECT().GetByID(mock.Anything, momentID, ownerID).Return(&entity.Moment{ID: momentID, UserID: ownerID}, nil)
	users.EXPECT().GetByUsername(mock.Anything, "gede").Return(target, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, ownerID, targetID).Return(true, nil)

	mention, err := uc.CreateMention(context.Background(), usecase.CreateMentionInput{
		MomentID: momentID, OwnerUserID: ownerID, Username: "gede",
	})

	assert.Nil(t, mention)
	assert.ErrorIs(t, err, errs.ErrAccessDenied)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestMentionUsecase_CreateMention_NobodyPolicy_Denied(t *testing.T) {
	repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications := newMentionUsecaseDeps(t)
	uc := usecase.NewMentionUsecase(repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications)

	momentID, ownerID, targetID := uuid.New(), uuid.New(), uuid.New()
	target := &entity.User{ID: targetID, MentionPolicy: enum.AudiencePolicyNobody}

	moments.EXPECT().GetByID(mock.Anything, momentID, ownerID).Return(&entity.Moment{ID: momentID, UserID: ownerID}, nil)
	users.EXPECT().GetByUsername(mock.Anything, "gede").Return(target, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, ownerID, targetID).Return(false, nil)

	mention, err := uc.CreateMention(context.Background(), usecase.CreateMentionInput{
		MomentID: momentID, OwnerUserID: ownerID, Username: "gede",
	})

	assert.Nil(t, mention)
	assert.ErrorIs(t, err, errs.ErrAccessDenied)
}

func TestMentionUsecase_CreateMention_KnownPolicy_RequiresKnown(t *testing.T) {
	repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications := newMentionUsecaseDeps(t)
	uc := usecase.NewMentionUsecase(repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications)

	momentID, ownerID, targetID := uuid.New(), uuid.New(), uuid.New()
	target := &entity.User{ID: targetID, MentionPolicy: enum.AudiencePolicyKnown}

	moments.EXPECT().GetByID(mock.Anything, momentID, ownerID).Return(&entity.Moment{ID: momentID, UserID: ownerID}, nil)
	users.EXPECT().GetByUsername(mock.Anything, "gede").Return(target, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, ownerID, targetID).Return(false, nil)
	knowns.EXPECT().IsKnownTo(mock.Anything, targetID, ownerID).Return(false, nil)

	mention, err := uc.CreateMention(context.Background(), usecase.CreateMentionInput{
		MomentID: momentID, OwnerUserID: ownerID, Username: "gede",
	})

	assert.Nil(t, mention)
	assert.ErrorIs(t, err, errs.ErrAccessDenied)
}

func TestMentionUsecase_LeaveMention_Success(t *testing.T) {
	repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications := newMentionUsecaseDeps(t)
	uc := usecase.NewMentionUsecase(repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications)

	momentID, userID := uuid.New(), uuid.New()
	repo.EXPECT().Remove(mock.Anything, momentID, userID).Return(nil)

	err := uc.LeaveMention(context.Background(), momentID, userID)

	assert.NoError(t, err)
}

func TestMentionUsecase_ShareToCircle_RequiresOwnershipAndMembership(t *testing.T) {
	repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications := newMentionUsecaseDeps(t)
	uc := usecase.NewMentionUsecase(repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications)

	momentID, circleID, userID := uuid.New(), uuid.New(), uuid.New()
	moments.EXPECT().GetByID(mock.Anything, momentID, userID).Return(&entity.Moment{ID: momentID, UserID: userID}, nil)
	circles.EXPECT().GetActiveMember(mock.Anything, circleID, userID).Return(nil, errs.ErrNotFound)

	share, err := uc.ShareToCircle(context.Background(), usecase.ShareMomentToCircleInput{
		MomentID: momentID, CircleID: circleID, UserID: userID,
	})

	assert.Nil(t, share)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestMentionUsecase_ShareToCircle_Idempotent(t *testing.T) {
	repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications := newMentionUsecaseDeps(t)
	uc := usecase.NewMentionUsecase(repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications)

	momentID, circleID, userID := uuid.New(), uuid.New(), uuid.New()
	moments.EXPECT().GetByID(mock.Anything, momentID, userID).Return(&entity.Moment{ID: momentID, UserID: userID}, nil)
	circles.EXPECT().GetActiveMember(mock.Anything, circleID, userID).Return(&entity.CircleMember{}, nil)
	repo.EXPECT().ShareToCircle(mock.Anything, momentID, circleID, userID).Return(nil, nil)

	share, err := uc.ShareToCircle(context.Background(), usecase.ShareMomentToCircleInput{
		MomentID: momentID, CircleID: circleID, UserID: userID,
	})

	assert.NoError(t, err)
	assert.Nil(t, share)
}

func TestMentionUsecase_ListSharedCircles_Success(t *testing.T) {
	repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications := newMentionUsecaseDeps(t)
	uc := usecase.NewMentionUsecase(repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications)

	momentID, ownerID, circleID := uuid.New(), uuid.New(), uuid.New()
	moments.EXPECT().GetByID(mock.Anything, momentID, ownerID).Return(&entity.Moment{ID: momentID, UserID: ownerID}, nil)
	repo.EXPECT().ListSharedCircleIDs(mock.Anything, momentID).Return([]uuid.UUID{circleID}, nil)

	ids, err := uc.ListSharedCircles(context.Background(), momentID, ownerID)

	assert.NoError(t, err)
	assert.Equal(t, []uuid.UUID{circleID}, ids)
}

func TestMentionUsecase_ListSharedCircles_NotOwner_NotFound(t *testing.T) {
	repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications := newMentionUsecaseDeps(t)
	uc := usecase.NewMentionUsecase(repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications)

	momentID, userID := uuid.New(), uuid.New()
	moments.EXPECT().GetByID(mock.Anything, momentID, userID).Return(nil, errs.ErrNotFound)
	// repo.ListSharedCircleIDs deliberately not stubbed.

	ids, err := uc.ListSharedCircles(context.Background(), momentID, userID)

	assert.Nil(t, ids)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestMentionUsecase_ListMentionedMoments_DefaultsPagination(t *testing.T) {
	repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications := newMentionUsecaseDeps(t)
	uc := usecase.NewMentionUsecase(repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications)

	userID := uuid.New()
	repo.EXPECT().ListMentionedMoments(mock.Anything, userID, int32(20), int32(0)).
		Return([]*entity.Moment{}, nil)

	result, err := uc.ListMentionedMoments(context.Background(), usecase.ListMentionedMomentsInput{UserID: userID})

	assert.NoError(t, err)
	assert.EqualValues(t, 1, result.Page)
	assert.EqualValues(t, 20, result.PageSize)
}

func TestMentionUsecase_SearchMentionableUsers_EmptyQuery_SkipsSearch(t *testing.T) {
	repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications := newMentionUsecaseDeps(t)
	uc := usecase.NewMentionUsecase(repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications)

	results, err := uc.SearchMentionableUsers(context.Background(), uuid.New(), "  ")

	assert.NoError(t, err)
	assert.Empty(t, results)
	searcher.AssertNotCalled(t, "SearchByUsernamePrefix", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestMentionUsecase_SearchMentionableUsers_FiltersByPolicyAndBlocking(t *testing.T) {
	repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications := newMentionUsecaseDeps(t)
	uc := usecase.NewMentionUsecase(repo, moments, circles, circleShares, users, knowns, blocks, searcher, notifications)

	requesterID := uuid.New()
	anyoneUser := &entity.User{ID: uuid.New(), Username: strPtr("gede"), MentionPolicy: enum.AudiencePolicyAnyone}
	knownUserAllowed := &entity.User{ID: uuid.New(), Username: strPtr("gerald"), MentionPolicy: enum.AudiencePolicyKnown}
	knownUserDenied := &entity.User{ID: uuid.New(), Username: strPtr("geri"), MentionPolicy: enum.AudiencePolicyKnown}
	nobodyUser := &entity.User{ID: uuid.New(), Username: strPtr("george"), MentionPolicy: enum.AudiencePolicyNobody}
	blockedUser := &entity.User{ID: uuid.New(), Username: strPtr("geo"), MentionPolicy: enum.AudiencePolicyAnyone}

	searcher.EXPECT().SearchByUsernamePrefix(mock.Anything, requesterID, "ge", int32(20)).
		Return([]*entity.User{anyoneUser, knownUserAllowed, knownUserDenied, nobodyUser, blockedUser}, nil)

	for _, u := range []*entity.User{anyoneUser, knownUserAllowed, knownUserDenied, nobodyUser} {
		blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, requesterID, u.ID).Return(false, nil)
	}
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, requesterID, blockedUser.ID).Return(true, nil)
	knowns.EXPECT().IsKnownTo(mock.Anything, knownUserAllowed.ID, requesterID).Return(true, nil)
	knowns.EXPECT().IsKnownTo(mock.Anything, knownUserDenied.ID, requesterID).Return(false, nil)

	results, err := uc.SearchMentionableUsers(context.Background(), requesterID, "ge")

	assert.NoError(t, err)
	assert.ElementsMatch(t, []*entity.User{anyoneUser, knownUserAllowed}, results)
}
