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

// newTestResponseAccessChecker builds a real ResponseAccessChecker over
// mocked leaves — it's a concrete shared helper (CODING_STANDARDS.md
// §5), not a swappable interface, so tests exercise it for real rather
// than mocking it directly.
func newTestResponseAccessChecker(t *testing.T) (*usecase.ResponseAccessChecker, *mocks.MockMomentReader, *mocks.MockUserBlockChecker) {
	t.Helper()
	moments := mocks.NewMockMomentReader(t)
	blocks := mocks.NewMockUserBlockChecker(t)
	return usecase.NewResponseAccessChecker(moments, blocks), moments, blocks
}

func TestCommentUsecase_CreateComment_MentionContext_OwnerNoResponseEvent(t *testing.T) {
	repo := mocks.NewMockCommentRepository(t)
	events := mocks.NewMockResponseEventRepository(t)
	access, moments, blocks := newTestResponseAccessChecker(t)
	uc := usecase.NewCommentUsecase(repo, access, events)

	momentID, ownerID := uuid.New(), uuid.New()
	moments.EXPECT().GetWithAccess(mock.Anything, momentID, ownerID).
		Return(&entity.Moment{ID: momentID, UserID: ownerID}, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, ownerID, ownerID).Return(false, nil)
	moments.EXPECT().ListVisibleCircleIDs(mock.Anything, momentID, ownerID).Return(nil, nil)
	repo.EXPECT().Create(mock.Anything, momentID, ownerID, (*uuid.UUID)(nil), "nice run").
		Return(&entity.Comment{ID: uuid.New(), MomentID: momentID, UserID: &ownerID, Body: "nice run"}, nil)

	comment, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		MomentID: momentID, UserID: ownerID, Body: "nice run",
	})

	assert.NoError(t, err)
	assert.Equal(t, "nice run", comment.Body)
	// The owner commenting on their own Moment must not create a
	// response_events row — verified implicitly by never stubbing
	// events.Create; testify/mock fails the test if it's called
	// unexpectedly.
}

func TestCommentUsecase_CreateComment_MentionContext_NonOwner_RecordsResponseEvent(t *testing.T) {
	repo := mocks.NewMockCommentRepository(t)
	events := mocks.NewMockResponseEventRepository(t)
	access, moments, blocks := newTestResponseAccessChecker(t)
	uc := usecase.NewCommentUsecase(repo, access, events)

	momentID, ownerID, actorID := uuid.New(), uuid.New(), uuid.New()
	moments.EXPECT().GetWithAccess(mock.Anything, momentID, actorID).
		Return(&entity.Moment{ID: momentID, UserID: ownerID}, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, actorID, ownerID).Return(false, nil)
	moments.EXPECT().ListVisibleCircleIDs(mock.Anything, momentID, actorID).Return(nil, nil)
	moments.EXPECT().IsActivelyMentioned(mock.Anything, momentID, actorID).Return(true, nil)

	created := &entity.Comment{ID: uuid.New(), MomentID: momentID, UserID: &actorID, Body: "congrats!"}
	repo.EXPECT().Create(mock.Anything, momentID, actorID, (*uuid.UUID)(nil), "congrats!").Return(created, nil)
	events.EXPECT().Create(mock.Anything, mock.MatchedBy(func(e *entity.ResponseEvent) bool {
		return e.RecipientUserID == ownerID && *e.ActorUserID == actorID && *e.CommentID == created.ID
	})).Return(nil)

	comment, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		MomentID: momentID, UserID: actorID, Body: "congrats!",
	})

	assert.NoError(t, err)
	assert.Equal(t, created, comment)
}

func TestCommentUsecase_CreateComment_MentionContext_NotMentioned_Forbidden(t *testing.T) {
	repo := mocks.NewMockCommentRepository(t)
	events := mocks.NewMockResponseEventRepository(t)
	access, moments, blocks := newTestResponseAccessChecker(t)
	uc := usecase.NewCommentUsecase(repo, access, events)

	momentID, ownerID, circleMemberID, circleID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	// The viewer reaches the Moment only via a Circle share, so
	// GetWithAccess succeeds, but they have no active mention — posting
	// into the mention context (circle_id nil) must be forbidden.
	moments.EXPECT().GetWithAccess(mock.Anything, momentID, circleMemberID).
		Return(&entity.Moment{ID: momentID, UserID: ownerID}, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, circleMemberID, ownerID).Return(false, nil)
	moments.EXPECT().ListVisibleCircleIDs(mock.Anything, momentID, circleMemberID).Return([]uuid.UUID{circleID}, nil)
	moments.EXPECT().IsActivelyMentioned(mock.Anything, momentID, circleMemberID).Return(false, nil)

	comment, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		MomentID: momentID, UserID: circleMemberID, Body: "hi",
	})

	assert.Nil(t, comment)
	assert.ErrorIs(t, err, errs.ErrForbidden)
}

func TestCommentUsecase_CreateComment_CircleContext_NotAVisibleCircle_Forbidden(t *testing.T) {
	repo := mocks.NewMockCommentRepository(t)
	events := mocks.NewMockResponseEventRepository(t)
	access, moments, blocks := newTestResponseAccessChecker(t)
	uc := usecase.NewCommentUsecase(repo, access, events)

	momentID, ownerID, viewerID, otherCircleID, wantedCircleID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	moments.EXPECT().GetWithAccess(mock.Anything, momentID, viewerID).
		Return(&entity.Moment{ID: momentID, UserID: ownerID}, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, viewerID, ownerID).Return(false, nil)
	moments.EXPECT().ListVisibleCircleIDs(mock.Anything, momentID, viewerID).Return([]uuid.UUID{otherCircleID}, nil)
	moments.EXPECT().IsActivelyMentioned(mock.Anything, momentID, viewerID).Return(false, nil)

	comment, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		MomentID: momentID, UserID: viewerID, CircleID: &wantedCircleID, Body: "hi",
	})

	assert.Nil(t, comment)
	assert.ErrorIs(t, err, errs.ErrForbidden)
}

func TestCommentUsecase_CreateComment_Blocked_Denied(t *testing.T) {
	repo := mocks.NewMockCommentRepository(t)
	events := mocks.NewMockResponseEventRepository(t)
	access, moments, blocks := newTestResponseAccessChecker(t)
	uc := usecase.NewCommentUsecase(repo, access, events)

	momentID, ownerID, viewerID := uuid.New(), uuid.New(), uuid.New()
	moments.EXPECT().GetWithAccess(mock.Anything, momentID, viewerID).
		Return(&entity.Moment{ID: momentID, UserID: ownerID}, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, viewerID, ownerID).Return(true, nil)

	comment, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		MomentID: momentID, UserID: viewerID, Body: "hi",
	})

	assert.Nil(t, comment)
	assert.ErrorIs(t, err, errs.ErrAccessDenied)
}

func TestCommentUsecase_ListComments_PassesResolvedAudienceThrough(t *testing.T) {
	repo := mocks.NewMockCommentRepository(t)
	events := mocks.NewMockResponseEventRepository(t)
	access, moments, blocks := newTestResponseAccessChecker(t)
	uc := usecase.NewCommentUsecase(repo, access, events)

	momentID, ownerID, viewerID, circleID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	moments.EXPECT().GetWithAccess(mock.Anything, momentID, viewerID).
		Return(&entity.Moment{ID: momentID, UserID: ownerID}, nil)
	blocks.EXPECT().IsBlockedEitherDirection(mock.Anything, viewerID, ownerID).Return(false, nil)
	moments.EXPECT().ListVisibleCircleIDs(mock.Anything, momentID, viewerID).Return([]uuid.UUID{circleID}, nil)
	moments.EXPECT().IsActivelyMentioned(mock.Anything, momentID, viewerID).Return(false, nil)

	expected := []*entity.Comment{{ID: uuid.New(), MomentID: momentID}}
	repo.EXPECT().ListForMoment(mock.Anything, momentID, viewerID, false, []uuid.UUID{circleID}).Return(expected, nil)

	comments, err := uc.ListComments(context.Background(), momentID, viewerID)

	assert.NoError(t, err)
	assert.Equal(t, expected, comments)
}

func TestCommentUsecase_UpdateComment_NotFound(t *testing.T) {
	repo := mocks.NewMockCommentRepository(t)
	events := mocks.NewMockResponseEventRepository(t)
	access, _, _ := newTestResponseAccessChecker(t)
	uc := usecase.NewCommentUsecase(repo, access, events)

	repo.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errs.ErrNotFound)

	comment, err := uc.UpdateComment(context.Background(), usecase.UpdateCommentInput{
		ID: uuid.New(), UserID: uuid.New(), Body: "edited",
	})

	assert.Nil(t, comment)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestCommentUsecase_SoftDeleteComment_Success(t *testing.T) {
	repo := mocks.NewMockCommentRepository(t)
	events := mocks.NewMockResponseEventRepository(t)
	access, _, _ := newTestResponseAccessChecker(t)
	uc := usecase.NewCommentUsecase(repo, access, events)

	id, userID := uuid.New(), uuid.New()
	repo.EXPECT().SoftDelete(mock.Anything, id, userID).Return(nil)

	err := uc.SoftDeleteComment(context.Background(), id, userID)

	assert.NoError(t, err)
}
