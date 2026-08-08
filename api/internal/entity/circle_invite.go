package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// CircleInvite has two deliberately asymmetric shapes (see
// enum.CircleInviteKind): Username invites carry InviteeUserID and no
// TokenHash; Link invites carry TokenHash (and optionally MaxUses) and
// no InviteeUserID.
type CircleInvite struct {
	ID              uuid.UUID
	CircleID        uuid.UUID
	InvitedByUserID *uuid.UUID
	Kind            enum.CircleInviteKind
	InviteeUserID   *uuid.UUID
	TokenHash       *string
	Status          enum.CircleInviteStatus
	MaxUses         *int32
	UseCount        int32
	ExpiresAt       *time.Time
	RespondedAt     *time.Time
	CreatedAt       time.Time
}
