package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// --- Repository interfaces, defined here per the Dependency Rule ---

type NotificationRepository interface {
	Create(ctx context.Context, n *entity.Notification) (*entity.Notification, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.Notification, error)
	Count(ctx context.Context, userID uuid.UUID) (int64, error)
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	// MarkRead and MarkAllRead are sqlc :exec queries — a WHERE clause
	// matching zero rows (already read, wrong owner, nothing pending) is
	// a silent no-op.
	MarkRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
}

// NotificationPreferenceRepository.Get always returns a row — it lazily
// creates the all-defaults one the first time a user has none, so
// "missing row" never has to mean "off" above this layer.
type NotificationPreferenceRepository interface {
	Get(ctx context.Context, userID uuid.UUID) (*entity.NotificationPreference, error)
	Update(ctx context.Context, pref *entity.NotificationPreference) (*entity.NotificationPreference, error)
}

// --- Inputs / outputs ---

// CreateNotificationInput carries every possible subject id; callers only
// populate the one(s) relevant to Kind (see entity.Notification's field
// comments for which subject goes with which kind).
type CreateNotificationInput struct {
	UserID      uuid.UUID
	Kind        enum.NotificationKind
	ActorUserID *uuid.UUID

	MomentID               *uuid.UUID
	ThreadID               *uuid.UUID
	CommitmentID           *uuid.UUID
	CommitmentOccurrenceID *uuid.UUID
	CircleID               *uuid.UUID
	CircleInviteID         *uuid.UUID
	CircleJoinRequestID    *uuid.UUID
	RecapID                *uuid.UUID
	ResurfacingID          *uuid.UUID
}

type ListNotificationsInput struct {
	UserID   uuid.UUID
	Page     int32
	PageSize int32
}

type NotificationListResult struct {
	Notifications []*entity.Notification
	Total         int64
	UnreadCount   int64
	Page          int32
	PageSize      int32
}

type UpdateNotificationPreferencesInput struct {
	UserID uuid.UUID

	CommitmentUpcomingEnabled bool
	MomentDueEnabled          bool
	MomentMissedEnabled       bool

	EchoReadyEnabled      bool
	RecapGeneratedEnabled bool

	CircleInviteReceivedEnabled      bool
	CircleJoinRequestReceivedEnabled bool
	MentionedInMomentEnabled         bool

	AddedToCircleEnabled             bool
	CircleJoinRequestApprovedEnabled bool

	ResponseDigestEnabled bool
	ResponseDigestHour    int16

	// Both set or both nil — validated in UpdatePreferences before it
	// ever reaches the DB's matching CHECK constraint.
	QuietHoursStart *time.Time
	QuietHoursEnd   *time.Time
}

// --- Usecase ---

// NotificationCreator is the capability every other domain usecase needs
// to raise its own real-time trigger — declared once here since
// mention_usecase.go, circle_invite_usecase.go, and
// circle_join_request_usecase.go all need the identical signature
// (CODING_STANDARDS.md §5).
type NotificationCreator interface {
	// CreateNotification returns (nil, nil) when the recipient has this
	// Kind disabled — callers never need to branch on "was it created."
	CreateNotification(ctx context.Context, input CreateNotificationInput) (*entity.Notification, error)
}

type NotificationUsecase interface {
	NotificationCreator
	ListNotifications(ctx context.Context, input ListNotificationsInput) (*NotificationListResult, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	GetPreferences(ctx context.Context, userID uuid.UUID) (*entity.NotificationPreference, error)
	// UpdatePreferences returns errs.ErrInvalidInput if quiet hours are
	// only half-set or the digest hour is out of range.
	UpdatePreferences(
		ctx context.Context,
		input UpdateNotificationPreferencesInput,
	) (*entity.NotificationPreference, error)
}

type notificationUsecase struct {
	repo  NotificationRepository
	prefs NotificationPreferenceRepository
	users UserPolicyReader
}

func NewNotificationUsecase(
	repo NotificationRepository,
	prefs NotificationPreferenceRepository,
	users UserPolicyReader,
) NotificationUsecase {
	return &notificationUsecase{repo: repo, prefs: prefs, users: users}
}

// CreateNotification implements [NotificationCreator]. Skips silently if
// the recipient has this Kind disabled, otherwise pushes deliver_after
// forward when the recipient is currently inside their own quiet hours
// (FEATURES.md, Notification Preferences) rather than delaying the
// insert — see the notifications table's own deliver_after comment.
func (u *notificationUsecase) CreateNotification(
	ctx context.Context,
	input CreateNotificationInput,
) (*entity.Notification, error) {
	pref, err := u.prefs.Get(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("loading notification preferences: %w", err)
	}
	if !kindEnabled(pref, input.Kind) {
		return nil, nil
	}

	deliverAfter := time.Now()
	if pref.QuietHoursStart != nil && pref.QuietHoursEnd != nil {
		recipient, err := u.users.GetByID(ctx, input.UserID)
		if err != nil {
			return nil, fmt.Errorf("loading recipient for quiet hours: %w", err)
		}
		loc, err := time.LoadLocation(recipient.Timezone)
		if err != nil {
			// A malformed/unknown tz on the account shouldn't block
			// delivery outright — fall back to always-deliverable UTC.
			loc = time.UTC
		}
		deliverAfter = deliverAfterRespectingQuietHours(deliverAfter, *pref.QuietHoursStart, *pref.QuietHoursEnd, loc)
	}

	created, err := u.repo.Create(ctx, &entity.Notification{
		UserID:                 input.UserID,
		Kind:                   input.Kind,
		ActorUserID:            input.ActorUserID,
		MomentID:               input.MomentID,
		ThreadID:               input.ThreadID,
		CommitmentID:           input.CommitmentID,
		CommitmentOccurrenceID: input.CommitmentOccurrenceID,
		CircleID:               input.CircleID,
		CircleInviteID:         input.CircleInviteID,
		CircleJoinRequestID:    input.CircleJoinRequestID,
		RecapID:                input.RecapID,
		ResurfacingID:          input.ResurfacingID,
		DeliverAfter:           deliverAfter,
	})
	if err != nil {
		return nil, fmt.Errorf("creating notification: %w", err)
	}
	return created, nil
}

// ListNotifications implements [NotificationUsecase]: the in-app feed,
// plus the unread count the same screen needs for its header badge.
func (u *notificationUsecase) ListNotifications(
	ctx context.Context,
	input ListNotificationsInput,
) (*NotificationListResult, error) {
	page, pageSize := normalizedPage(input.Page, input.PageSize)

	notifications, err := u.repo.ListByUserID(ctx, input.UserID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("listing notifications: %w", err)
	}
	total, err := u.repo.Count(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("counting notifications: %w", err)
	}
	unread, err := u.repo.CountUnread(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("counting unread notifications: %w", err)
	}

	return &NotificationListResult{
		Notifications: notifications,
		Total:         total,
		UnreadCount:   unread,
		Page:          page,
		PageSize:      pageSize,
	}, nil
}

// MarkRead implements [NotificationUsecase].
func (u *notificationUsecase) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	if err := u.repo.MarkRead(ctx, id, userID); err != nil {
		return fmt.Errorf("marking notification read: %w", err)
	}
	return nil
}

// MarkAllRead implements [NotificationUsecase].
func (u *notificationUsecase) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	if err := u.repo.MarkAllRead(ctx, userID); err != nil {
		return fmt.Errorf("marking all notifications read: %w", err)
	}
	return nil
}

// GetPreferences implements [NotificationUsecase].
func (u *notificationUsecase) GetPreferences(
	ctx context.Context,
	userID uuid.UUID,
) (*entity.NotificationPreference, error) {
	pref, err := u.prefs.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting notification preferences: %w", err)
	}
	return pref, nil
}

// UpdatePreferences implements [NotificationUsecase].
func (u *notificationUsecase) UpdatePreferences(
	ctx context.Context,
	input UpdateNotificationPreferencesInput,
) (*entity.NotificationPreference, error) {
	if (input.QuietHoursStart == nil) != (input.QuietHoursEnd == nil) {
		return nil, errs.ErrInvalidInput
	}
	if input.ResponseDigestHour < 0 || input.ResponseDigestHour > 23 {
		return nil, errs.ErrInvalidInput
	}

	// Ensures a row exists before UPDATE — a user changing preferences
	// for the first time may not have one yet (Get lazily creates it).
	if _, err := u.prefs.Get(ctx, input.UserID); err != nil {
		return nil, fmt.Errorf("ensuring notification preferences exist: %w", err)
	}

	updated, err := u.prefs.Update(ctx, &entity.NotificationPreference{
		UserID:                           input.UserID,
		CommitmentUpcomingEnabled:        input.CommitmentUpcomingEnabled,
		MomentDueEnabled:                 input.MomentDueEnabled,
		MomentMissedEnabled:              input.MomentMissedEnabled,
		EchoReadyEnabled:                 input.EchoReadyEnabled,
		RecapGeneratedEnabled:            input.RecapGeneratedEnabled,
		CircleInviteReceivedEnabled:      input.CircleInviteReceivedEnabled,
		CircleJoinRequestReceivedEnabled: input.CircleJoinRequestReceivedEnabled,
		MentionedInMomentEnabled:         input.MentionedInMomentEnabled,
		AddedToCircleEnabled:             input.AddedToCircleEnabled,
		CircleJoinRequestApprovedEnabled: input.CircleJoinRequestApprovedEnabled,
		ResponseDigestEnabled:            input.ResponseDigestEnabled,
		ResponseDigestHour:               input.ResponseDigestHour,
		QuietHoursStart:                  input.QuietHoursStart,
		QuietHoursEnd:                    input.QuietHoursEnd,
	})
	if err != nil {
		return nil, fmt.Errorf("updating notification preferences: %w", err)
	}
	return updated, nil
}

func kindEnabled(pref *entity.NotificationPreference, kind enum.NotificationKind) bool {
	switch kind {
	case enum.NotificationKindCommitmentUpcoming:
		return pref.CommitmentUpcomingEnabled
	case enum.NotificationKindMomentDue:
		return pref.MomentDueEnabled
	case enum.NotificationKindMomentMissed:
		return pref.MomentMissedEnabled
	case enum.NotificationKindEchoReady:
		return pref.EchoReadyEnabled
	case enum.NotificationKindRecapGenerated:
		return pref.RecapGeneratedEnabled
	case enum.NotificationKindCircleInviteReceived:
		return pref.CircleInviteReceivedEnabled
	case enum.NotificationKindCircleJoinRequestReceived:
		return pref.CircleJoinRequestReceivedEnabled
	case enum.NotificationKindMentionedInMoment:
		return pref.MentionedInMomentEnabled
	case enum.NotificationKindAddedToCircle:
		return pref.AddedToCircleEnabled
	case enum.NotificationKindCircleJoinRequestApproved:
		return pref.CircleJoinRequestApprovedEnabled
	case enum.NotificationKindResponseDigest:
		return pref.ResponseDigestEnabled
	default:
		return false
	}
}

// deliverAfterRespectingQuietHours returns now unless the recipient's
// local time currently falls inside quietStart/quietEnd (only their
// wall-clock hour/minute/second matter), in which case it returns the
// local instant the quiet window ends. The window may wrap past midnight
// (e.g. 22:00 -> 07:00), per notification_preferences.quiet_hours_*'s own
// migration comment.
func deliverAfterRespectingQuietHours(now, quietStart, quietEnd time.Time, loc *time.Location) time.Time {
	local := now.In(loc)
	startOfDay := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	start := startOfDay.Add(wallClockOffset(quietStart))
	end := startOfDay.Add(wallClockOffset(quietEnd))

	if !start.After(end) {
		// Same-day window, e.g. 01:00 -> 06:00.
		if !local.Before(start) && local.Before(end) {
			return end
		}
		return now
	}

	// Wraps past midnight, e.g. 22:00 -> 07:00.
	if !local.Before(start) {
		return end.AddDate(0, 0, 1)
	}
	if local.Before(end) {
		return end
	}
	return now
}

func wallClockOffset(t time.Time) time.Duration {
	return time.Duration(t.Hour())*time.Hour +
		time.Duration(t.Minute())*time.Minute +
		time.Duration(t.Second())*time.Second
}
