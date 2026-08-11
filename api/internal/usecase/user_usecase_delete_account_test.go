//go:build unit

package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/usecase/mocks"
)

// deleteAccountMocks bundles every mock DeleteAccount's transaction
// touches, plus the top-level ProfileImageStorage mock userUsecase
// itself holds outside the transaction.
type deleteAccountMocks struct {
	uow           *mocks.MockAccountDeletionUnitOfWork
	users         *mocks.MockUserRepository
	refreshTokens *mocks.MockRefreshTokenRepository
	moments       *mocks.MockMomentRepository
	momentImages  *mocks.MockMomentImageRepository
	threads       *mocks.MockThreadRepository
	threadImages  *mocks.MockThreadImageRepository
	comments      *mocks.MockCommentRepository
	reactions     *mocks.MockReactionRepository
	circles       *mocks.MockCircleRepository
	storage       *mocks.MockProfileImageStorage
}

func newUserUsecaseForDeleteAccount(t *testing.T) (usecase.UserUsecase, *deleteAccountMocks) {
	m := &deleteAccountMocks{
		uow:           mocks.NewMockAccountDeletionUnitOfWork(t),
		users:         mocks.NewMockUserRepository(t),
		refreshTokens: mocks.NewMockRefreshTokenRepository(t),
		moments:       mocks.NewMockMomentRepository(t),
		momentImages:  mocks.NewMockMomentImageRepository(t),
		threads:       mocks.NewMockThreadRepository(t),
		threadImages:  mocks.NewMockThreadImageRepository(t),
		comments:      mocks.NewMockCommentRepository(t),
		reactions:     mocks.NewMockReactionRepository(t),
		circles:       mocks.NewMockCircleRepository(t),
		storage:       mocks.NewMockProfileImageStorage(t),
	}
	knowns := mocks.NewMockUserKnownRepository(t)
	blocks := mocks.NewMockUserBlockRepository(t)
	mutes := mocks.NewMockUserMuteRepository(t)

	m.uow.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(usecase.AccountDeletionRepositories) error) error {
			return fn(usecase.AccountDeletionRepositories{
				User:         m.users,
				RefreshToken: m.refreshTokens,
				Moment:       m.moments,
				MomentImage:  m.momentImages,
				Thread:       m.threads,
				ThreadImage:  m.threadImages,
				Comment:      m.comments,
				Reaction:     m.reactions,
				Circle:       m.circles,
			})
		})

	uc := usecase.NewUserUsecase(m.users, knowns, blocks, mutes, m.storage, m.uow)
	return uc, m
}

// expectNoCircleMemberships is the shared tail every DeleteAccount test
// needs once circle handling is done: no memberships, then the
// account/session teardown steps.
func (m *deleteAccountMocks) expectCoreDeletionSteps(userID uuid.UUID) {
	m.moments.EXPECT().SoftDeleteAllByUserID(mock.Anything, userID).Return(nil)
	m.threads.EXPECT().SoftDeleteAllByUserID(mock.Anything, userID).Return(nil)
	m.comments.EXPECT().AnonymizeByUserID(mock.Anything, userID).Return(nil)
	m.reactions.EXPECT().AnonymizeByUserID(mock.Anything, userID).Return(nil)
	m.users.EXPECT().SoftDelete(mock.Anything, userID).Return(nil)
	m.users.EXPECT().UpdateImagePath(mock.Anything, userID, (*string)(nil)).Return(&entity.User{ID: userID}, nil)
	m.refreshTokens.EXPECT().RevokeAllByUserID(mock.Anything, userID).Return(nil)
}

func TestUserUsecase_DeleteAccount_Success_NoCircles(t *testing.T) {
	uc, m := newUserUsecaseForDeleteAccount(t)

	userID := uuid.New()
	m.users.EXPECT().GetByID(mock.Anything, userID).Return(&entity.User{ID: userID, ImagePath: nil}, nil)
	m.momentImages.EXPECT().ListImagePathsByOwnerID(mock.Anything, userID).Return(nil, nil)
	m.threadImages.EXPECT().ListImagePathsByOwnerID(mock.Anything, userID).Return(nil, nil)
	m.circles.EXPECT().ListByUserID(mock.Anything, userID).Return(nil, nil)
	m.expectCoreDeletionSteps(userID)

	err := uc.DeleteAccount(context.Background(), userID)

	require.NoError(t, err)
}

func TestUserUsecase_DeleteAccount_CleansUpImagesAndProfilePhoto(t *testing.T) {
	uc, m := newUserUsecaseForDeleteAccount(t)

	userID := uuid.New()
	profileImageURL := "https://cdn.example/users/" + userID.String() + "/avatar.jpg"
	m.users.EXPECT().GetByID(mock.Anything, userID).Return(&entity.User{ID: userID, ImagePath: &profileImageURL}, nil)
	m.storage.EXPECT().PublicURL("").Return("https://cdn.example/")
	m.momentImages.EXPECT().ListImagePathsByOwnerID(mock.Anything, userID).Return([]string{"moments/a/1.jpg"}, nil)
	m.threadImages.EXPECT().ListImagePathsByOwnerID(mock.Anything, userID).Return([]string{"threads/b/2.jpg"}, nil)
	m.circles.EXPECT().ListByUserID(mock.Anything, userID).Return(nil, nil)
	m.expectCoreDeletionSteps(userID)

	m.storage.EXPECT().Delete(mock.Anything, "moments/a/1.jpg").Return(nil)
	m.storage.EXPECT().Delete(mock.Anything, "threads/b/2.jpg").Return(nil)
	m.storage.EXPECT().Delete(mock.Anything, "users/"+userID.String()+"/avatar.jpg").Return(nil)

	err := uc.DeleteAccount(context.Background(), userID)

	require.NoError(t, err)
}

func TestUserUsecase_DeleteAccount_NotSoleAdmin_JustLeaves(t *testing.T) {
	uc, m := newUserUsecaseForDeleteAccount(t)

	userID := uuid.New()
	circleID := uuid.New()
	m.users.EXPECT().GetByID(mock.Anything, userID).Return(&entity.User{ID: userID}, nil)
	m.momentImages.EXPECT().ListImagePathsByOwnerID(mock.Anything, userID).Return(nil, nil)
	m.threadImages.EXPECT().ListImagePathsByOwnerID(mock.Anything, userID).Return(nil, nil)
	m.circles.EXPECT().ListByUserID(mock.Anything, userID).Return([]*entity.Circle{{ID: circleID}}, nil)
	m.circles.EXPECT().IsSoleActiveAdmin(mock.Anything, circleID, userID).Return(false, nil)
	m.circles.EXPECT().Leave(mock.Anything, circleID, userID).Return(nil)
	m.expectCoreDeletionSteps(userID)

	err := uc.DeleteAccount(context.Background(), userID)

	require.NoError(t, err)
}

func TestUserUsecase_DeleteAccount_SoleAdmin_PromotesLongestTenuredMember(t *testing.T) {
	uc, m := newUserUsecaseForDeleteAccount(t)

	userID := uuid.New()
	circleID := uuid.New()
	successorID := uuid.New()
	laterMemberID := uuid.New()

	m.users.EXPECT().GetByID(mock.Anything, userID).Return(&entity.User{ID: userID}, nil)
	m.momentImages.EXPECT().ListImagePathsByOwnerID(mock.Anything, userID).Return(nil, nil)
	m.threadImages.EXPECT().ListImagePathsByOwnerID(mock.Anything, userID).Return(nil, nil)
	m.circles.EXPECT().ListByUserID(mock.Anything, userID).Return([]*entity.Circle{{ID: circleID}}, nil)
	m.circles.EXPECT().IsSoleActiveAdmin(mock.Anything, circleID, userID).Return(true, nil)
	// ListMembers is ordered by joined_at ascending — the deleting user
	// first, then the longest-tenured other member, then someone newer.
	// The newer member must never be considered.
	m.circles.EXPECT().ListMembers(mock.Anything, circleID).Return([]*entity.CircleMember{
		{CircleID: circleID, UserID: userID, Role: enum.CircleRoleAdmin},
		{CircleID: circleID, UserID: successorID, Role: enum.CircleRoleMember},
		{CircleID: circleID, UserID: laterMemberID, Role: enum.CircleRoleMember},
	}, nil)
	m.circles.EXPECT().
		UpdateMemberRole(mock.Anything, circleID, successorID, userID, enum.CircleRoleAdmin).
		Return(&entity.CircleMember{CircleID: circleID, UserID: successorID, Role: enum.CircleRoleAdmin}, nil)
	m.circles.EXPECT().Leave(mock.Anything, circleID, userID).Return(nil)
	m.expectCoreDeletionSteps(userID)

	err := uc.DeleteAccount(context.Background(), userID)

	require.NoError(t, err)
}

func TestUserUsecase_DeleteAccount_SoleAdminAlone_DissolvesCircle(t *testing.T) {
	uc, m := newUserUsecaseForDeleteAccount(t)

	userID := uuid.New()
	circleID := uuid.New()

	m.users.EXPECT().GetByID(mock.Anything, userID).Return(&entity.User{ID: userID}, nil)
	m.momentImages.EXPECT().ListImagePathsByOwnerID(mock.Anything, userID).Return(nil, nil)
	m.threadImages.EXPECT().ListImagePathsByOwnerID(mock.Anything, userID).Return(nil, nil)
	m.circles.EXPECT().ListByUserID(mock.Anything, userID).Return([]*entity.Circle{{ID: circleID}}, nil)
	m.circles.EXPECT().IsSoleActiveAdmin(mock.Anything, circleID, userID).Return(true, nil)
	m.circles.EXPECT().ListMembers(mock.Anything, circleID).Return([]*entity.CircleMember{
		{CircleID: circleID, UserID: userID, Role: enum.CircleRoleAdmin},
	}, nil)
	// Dissolve, not Leave — no expectation set on Leave/UpdateMemberRole
	// at all, so mockery's t.Cleanup assertion fails the test if either
	// is reached.
	m.circles.EXPECT().Dissolve(mock.Anything, circleID, userID).Return(nil)
	m.expectCoreDeletionSteps(userID)

	err := uc.DeleteAccount(context.Background(), userID)

	require.NoError(t, err)
}
