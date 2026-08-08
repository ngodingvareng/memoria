package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// CircleInvite has two deliberately asymmetric shapes (see
// enum.CircleInviteKind): Username invites carry InviteeUserID and
// ExpiresAt, and no TokenHash; Link invites carry TokenHash and
// RequiresApproval, never expire, and have no usage limit.
type CircleInvite struct {
	ID              uuid.UUID
	CircleID        uuid.UUID
	InvitedByUserID *uuid.UUID
	Kind            enum.CircleInviteKind
	InviteeUserID   *uuid.UUID
	TokenHash       *string
	// RequiresApproval is set on Link invites only: whether following
	// the link joins the Circle outright or opens a CircleJoinRequest.
	RequiresApproval *bool
	Status           enum.CircleInviteStatus
	ExpiresAt        *time.Time
	RespondedAt      *time.Time
	CreatedAt        time.Time
}
