package entity

import (
	"time"

	"github.com/google/uuid"
)

// MomentMention is how a personal Moment, private by default, reaches
// someone outside its owner. DisplayName is a snapshot taken at mention
// time; it's what renders when the mention goes inert (mentioned user
// deleted their account, or MentionedUserID is blocked).
type MomentMention struct {
	ID              uuid.UUID
	MomentID        uuid.UUID
	MentionedUserID *uuid.UUID
	DisplayName     string
	RemovedAt       *time.Time
	CreatedAt       time.Time
}
