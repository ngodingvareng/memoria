package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// CircleMember is never deleted on leaving — LeftAt is stamped instead,
// so a past comment/reaction can still render with its Circle badge.
// Active membership is always (LeftAt == nil), never mere row existence.
type CircleMember struct {
	CircleID   uuid.UUID
	UserID     uuid.UUID
	Role       enum.CircleRole
	CanInvite  bool
	CanCapture bool
	JoinedAt   time.Time
	LeftAt     *time.Time
}
