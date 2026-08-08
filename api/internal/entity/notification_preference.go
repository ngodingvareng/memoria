package entity

import (
	"time"

	"github.com/google/uuid"
)

// NotificationPreference is one row per user, every NotificationKind
// togglable independently (column-per-type, so a missing row can never
// silently mean "off"). Defaults encode FEATURES.md's Notification
// Preferences exactly: gifts on, "someone needs you" on, the response
// digest on at a fixed daily hour, Commitment reminders on except
// MomentMissed, off until asked for.
type NotificationPreference struct {
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

	// Quiet hours in the user's local time; may wrap past midnight. Both
	// set or both nil — never just one.
	QuietHoursStart *time.Time
	QuietHoursEnd   *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
