package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserBlock is symmetric in effect even though only the blocker's row is
// stored: an access check must test the pair both ways (see the
// user_blocks migration comment).
type UserBlock struct {
	BlockerUserID uuid.UUID
	BlockedUserID uuid.UUID
	CreatedAt     time.Time
}
