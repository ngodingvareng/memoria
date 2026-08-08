package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserMute is one-directional and purely a view filter: the muted user
// is not restricted and is not notified.
type UserMute struct {
	MuterUserID uuid.UUID
	MutedUserID uuid.UUID
	CreatedAt   time.Time
}
