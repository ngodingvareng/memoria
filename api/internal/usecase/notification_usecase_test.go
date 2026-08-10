//go:build unit

package usecase_test

import (
	"context"
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
	"github.com/stretchr/testify/require"
)

func newNotificationUsecaseDeps(t *testing.T) (*mocks.MockNotificationRepository, *mocks.MockNotificationPreferenceRepository, *mocks.MockUserPolicyReader) {
	t.Helper()
	return mocks.NewMockNotificationRepository(t),
		mocks.NewMockNotificationPreferenceRepository(t),
		mocks.NewMockUserPolicyReader(t)
}

func allEnabledPreference(userID uuid.UUID) *entity.NotificationPreference {
	return &entity.NotificationPreference{
		UserID:                           userID,
		CommitmentUpcomingEnabled:        true,
		MomentDueEnabled:                 true,
		MomentMissedEnabled:              true,
		EchoReadyEnabled:                 true,
		RecapGeneratedEnabled:            true,
		CircleInviteReceivedEnabled:      true,
		CircleJoinRequestReceivedEnabled: true,
		MentionedInMomentEnabled:         true,
		AddedToCircleEnabled:             true,
		CircleJoinRequestApprovedEnabled: true,
		ResponseDigestEnabled:            true,
		ResponseDigestHour:               20,
	}
}

func TestNotificationUsecase_CreateNotification_KindDisabled_Skips(t *testing.T) {
	repo, prefs, users := newNotificationUsecaseDeps(t)
	uc := usecase.NewNotificationUsecase(repo, prefs, users)

	userID := uuid.New()
	pref := allEnabledPreference(userID)
	pref.MentionedInMomentEnabled = false
	prefs.EXPECT().Get(mock.Anything, userID).Return(pref, nil)
	// repo.Create must never be called — no expectation set on it at all,
	// so mockery's t.Cleanup assertion fails the test if it is.

	result, err := uc.CreateNotification(context.Background(), usecase.CreateNotificationInput{
		UserID: userID, Kind: enum.NotificationKindMentionedInMoment,
	})

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestNotificationUsecase_CreateNotification_KindEnabled_CreatesImmediately(t *testing.T) {
	repo, prefs, users := newNotificationUsecaseDeps(t)
	uc := usecase.NewNotificationUsecase(repo, prefs, users)

	userID := uuid.New()
	pref := allEnabledPreference(userID)
	prefs.EXPECT().Get(mock.Anything, userID).Return(pref, nil)

	created := &entity.Notification{ID: uuid.New(), UserID: userID, Kind: enum.NotificationKindMentionedInMoment}
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(n *entity.Notification) bool {
			return n.UserID == userID &&
				n.Kind == enum.NotificationKindMentionedInMoment &&
				!n.DeliverAfter.After(time.Now())
		})).
		Return(created, nil)

	result, err := uc.CreateNotification(context.Background(), usecase.CreateNotificationInput{
		UserID: userID, Kind: enum.NotificationKindMentionedInMoment,
	})

	require.NoError(t, err)
	assert.Equal(t, created, result)
}

func TestNotificationUsecase_CreateNotification_InsideQuietHours_DelaysDelivery(t *testing.T) {
	repo, prefs, users := newNotificationUsecaseDeps(t)
	uc := usecase.NewNotificationUsecase(repo, prefs, users)

	userID := uuid.New()
	pref := allEnabledPreference(userID)
	// A wrapping window (start > end) covering virtually the entire day
	// except the single midnight instant — deterministic regardless of
	// when this test actually runs, without duplicating the usecase's
	// own wall-clock arithmetic here.
	quietStart := time.Date(1, 1, 1, 0, 0, 1, 0, time.UTC)
	quietEnd := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	pref.QuietHoursStart = &quietStart
	pref.QuietHoursEnd = &quietEnd
	prefs.EXPECT().Get(mock.Anything, userID).Return(pref, nil)
	users.EXPECT().GetByID(mock.Anything, userID).Return(&entity.User{ID: userID, Timezone: "UTC"}, nil)

	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(n *entity.Notification) bool {
			return n.DeliverAfter.After(time.Now())
		})).
		Return(&entity.Notification{ID: uuid.New(), UserID: userID}, nil)

	_, err := uc.CreateNotification(context.Background(), usecase.CreateNotificationInput{
		UserID: userID, Kind: enum.NotificationKindMentionedInMoment,
	})

	assert.NoError(t, err)
}

func TestNotificationUsecase_ListNotifications_DefaultsPagination(t *testing.T) {
	repo, _, users := newNotificationUsecaseDeps(t)
	uc := usecase.NewNotificationUsecase(repo, mocks.NewMockNotificationPreferenceRepository(t), users)

	userID := uuid.New()
	repo.EXPECT().ListByUserID(mock.Anything, userID, int32(20), int32(0)).Return([]*entity.Notification{}, nil)
	repo.EXPECT().Count(mock.Anything, userID).Return(int64(0), nil)
	repo.EXPECT().CountUnread(mock.Anything, userID).Return(int64(0), nil)

	result, err := uc.ListNotifications(context.Background(), usecase.ListNotificationsInput{UserID: userID})

	require.NoError(t, err)
	assert.EqualValues(t, 1, result.Page)
	assert.EqualValues(t, 20, result.PageSize)
}

func TestNotificationUsecase_UpdatePreferences_QuietHoursHalfSet_Invalid(t *testing.T) {
	repo, prefs, users := newNotificationUsecaseDeps(t)
	uc := usecase.NewNotificationUsecase(repo, prefs, users)

	start := time.Now()
	_, err := uc.UpdatePreferences(context.Background(), usecase.UpdateNotificationPreferencesInput{
		UserID:          uuid.New(),
		QuietHoursStart: &start,
		QuietHoursEnd:   nil,
	})

	assert.ErrorIs(t, err, errs.ErrInvalidInput)
}

func TestNotificationUsecase_UpdatePreferences_DigestHourOutOfRange_Invalid(t *testing.T) {
	repo, prefs, users := newNotificationUsecaseDeps(t)
	uc := usecase.NewNotificationUsecase(repo, prefs, users)

	_, err := uc.UpdatePreferences(context.Background(), usecase.UpdateNotificationPreferencesInput{
		UserID:             uuid.New(),
		ResponseDigestHour: 24,
	})

	assert.ErrorIs(t, err, errs.ErrInvalidInput)
}

func TestNotificationUsecase_UpdatePreferences_Success(t *testing.T) {
	repo, prefs, users := newNotificationUsecaseDeps(t)
	uc := usecase.NewNotificationUsecase(repo, prefs, users)

	userID := uuid.New()
	existing := allEnabledPreference(userID)
	updated := allEnabledPreference(userID)
	updated.MentionedInMomentEnabled = false

	prefs.EXPECT().Get(mock.Anything, userID).Return(existing, nil)
	prefs.EXPECT().
		Update(mock.Anything, mock.MatchedBy(func(p *entity.NotificationPreference) bool {
			return p.UserID == userID && !p.MentionedInMomentEnabled
		})).
		Return(updated, nil)

	result, err := uc.UpdatePreferences(context.Background(), usecase.UpdateNotificationPreferencesInput{
		UserID:                   userID,
		MentionedInMomentEnabled: false,
		ResponseDigestHour:       20,
	})

	require.NoError(t, err)
	assert.Equal(t, updated, result)
}
