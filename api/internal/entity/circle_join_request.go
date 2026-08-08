package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// CircleJoinRequest is the approval step for a Link CircleInvite whose
// RequiresApproval is true. A Link without it admits people directly,
// and a direct add never produces one of these.
type CircleJoinRequest struct {
	ID              uuid.UUID
	CircleID        uuid.UUID
	UserID          uuid.UUID
	InviteID        *uuid.UUID
	Status          enum.CircleJoinRequestStatus
	DecidedByUserID *uuid.UUID
	DecidedAt       *time.Time
	CreatedAt       time.Time
}
