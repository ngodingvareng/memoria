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

func TestReactionUsecase_SetReaction_OwnerNoResponseEvent(t *testing.T) {
	repo := mocks.NewMockReactionRepository(t)
	events := mocks.NewMockResponseEventRepository(t)
	access, moments, blocks := newTestResponseAccessChecker(t)
	uc := usecase.NewReactionUsecase(repo, access, events)

	momentID, ownerID := uuid.New(), uuid.New()
	moments.EXPECT().GetWithAccess(mock.Anything, momentID, ownerID).
		Return(&entity.Moment{ID: momentID, UserID: ownerID}, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, ownerID, ownerID).Return(false, nil)
	moments.EXPECT().ListVisibleCircleIDs(mock.Anything, momentID, ownerID).Return(nil, nil)
	repo.EXPECT().Upsert(mock.Anything, momentID, ownerID, (*uuid.UUID)(nil), enum.ReactionKindHeart).
		Return(&entity.Reaction{ID: uuid.New(), MomentID: momentID, UserID: &ownerID, Kind: enum.ReactionKindHeart}, nil)

	reaction, err := uc.SetReaction(context.Background(), usecase.SetReactionInput{
		MomentID: momentID, UserID: ownerID, Kind: enum.ReactionKindHeart,
	})

	assert.NoError(t, err)
	assert.Equal(t, enum.ReactionKindHeart, reaction.Kind)
}

func TestReactionUsecase_SetReaction_NonOwner_RecordsResponseEvent(t *testing.T) {
	repo := mocks.NewMockReactionRepository(t)
	events := mocks.NewMockResponseEventRepository(t)
	access, moments, blocks := newTestResponseAccessChecker(t)
	uc := usecase.NewReactionUsecase(repo, access, events)

	momentID, ownerID, actorID := uuid.New(), uuid.New(), uuid.New()
	moments.EXPECT().GetWithAccess(mock.Anything, momentID, actorID).
		Return(&entity.Moment{ID: momentID, UserID: ownerID}, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, actorID, ownerID).Return(false, nil)
	moments.EXPECT().ListVisibleCircleIDs(mock.Anything, momentID, actorID).Return(nil, nil)
	moments.EXPECT().IsActivelyMentioned(mock.Anything, momentID, actorID).Return(true, nil)

	created := &entity.Reaction{ID: uuid.New(), MomentID: momentID, UserID: &actorID, Kind: enum.ReactionKindJoy}
	repo.EXPECT().Upsert(mock.Anything, momentID, actorID, (*uuid.UUID)(nil), enum.ReactionKindJoy).Return(created, nil)
	events.EXPECT().Create(mock.Anything, mock.MatchedBy(func(e *entity.ResponseEvent) bool {
		return e.RecipientUserID == ownerID && *e.ActorUserID == actorID && *e.ReactionID == created.ID
	})).Return(nil)

	reaction, err := uc.SetReaction(context.Background(), usecase.SetReactionInput{
		MomentID: momentID, UserID: actorID, Kind: enum.ReactionKindJoy,
	})

	assert.NoError(t, err)
	assert.Equal(t, created, reaction)
}

func TestReactionUsecase_SetReaction_Blocked_Denied(t *testing.T) {
	repo := mocks.NewMockReactionRepository(t)
	events := mocks.NewMockResponseEventRepository(t)
	access, moments, blocks := newTestResponseAccessChecker(t)
	uc := usecase.NewReactionUsecase(repo, access, events)

	momentID, ownerID, viewerID := uuid.New(), uuid.New(), uuid.New()
	moments.EXPECT().GetWithAccess(mock.Anything, momentID, viewerID).
		Return(&entity.Moment{ID: momentID, UserID: ownerID}, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, viewerID, ownerID).Return(true, nil)

	reaction, err := uc.SetReaction(context.Background(), usecase.SetReactionInput{
		MomentID: momentID, UserID: viewerID, Kind: enum.ReactionKindTender,
	})

	assert.Nil(t, reaction)
	assert.ErrorIs(t, err, errs.ErrAccessDenied)
}

func TestReactionUsecase_RemoveReaction_Success(t *testing.T) {
	repo := mocks.NewMockReactionRepository(t)
	events := mocks.NewMockResponseEventRepository(t)
	access, _, _ := newTestResponseAccessChecker(t)
	uc := usecase.NewReactionUsecase(repo, access, events)

	momentID, userID := uuid.New(), uuid.New()
	repo.EXPECT().Delete(mock.Anything, momentID, userID, (*uuid.UUID)(nil)).Return(nil)

	err := uc.RemoveReaction(context.Background(), usecase.RemoveReactionInput{MomentID: momentID, UserID: userID})

	assert.NoError(t, err)
}

func TestReactionUsecase_ListReactions_PassesResolvedAudienceThrough(t *testing.T) {
	repo := mocks.NewMockReactionRepository(t)
	events := mocks.NewMockResponseEventRepository(t)
	access, moments, blocks := newTestResponseAccessChecker(t)
	uc := usecase.NewReactionUsecase(repo, access, events)

	momentID, ownerID, viewerID, circleID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	moments.EXPECT().GetWithAccess(mock.Anything, momentID, viewerID).
		Return(&entity.Moment{ID: momentID, UserID: ownerID}, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, viewerID, ownerID).Return(false, nil)
	moments.EXPECT().ListVisibleCircleIDs(mock.Anything, momentID, viewerID).Return([]uuid.UUID{circleID}, nil)
	moments.EXPECT().IsActivelyMentioned(mock.Anything, momentID, viewerID).Return(false, nil)

	expected := []*entity.Reaction{{ID: uuid.New(), MomentID: momentID, Kind: enum.ReactionKindHeart}}
	repo.EXPECT().ListForMoment(mock.Anything, momentID, viewerID, false, []uuid.UUID{circleID}).Return(expected, nil)

	reactions, err := uc.ListReactions(context.Background(), momentID, viewerID)

	assert.NoError(t, err)
	assert.Equal(t, expected, reactions)
}
