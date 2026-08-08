package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// UserDevice is a push notification target. Notifications are queued
// separately (see Notification) and delivered to every live device row
// for the recipient.
type UserDevice struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Platform   enum.DevicePlatform
	PushToken  string
	Timezone   *string // the device's own zone, which may differ from User.Timezone while travelling
	LastSeenAt time.Time
	CreatedAt  time.Time
	RevokedAt  *time.Time
}
