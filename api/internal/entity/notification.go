package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// Notification is one row per deliverable alert. Subject fields are
// nullable and typed rather than a single polymorphic id, so every
// reference keeps its own meaning and a deleted subject can't leave a
// dangling notification pointing at nothing. Payload is rendering data
// only (e.g. digest counts, a Thread's name at send time) — never
// authoritative.
type Notification struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Kind   enum.NotificationKind

	// The person who caused it. Set only for MentionedInMoment and
	// CircleInviteReceived; nil for gifts and reminders, which have no
	// actor by design.
	ActorUserID *uuid.UUID

	MomentID               *uuid.UUID
	ThreadID               *uuid.UUID
	CommitmentID           *uuid.UUID
	CommitmentOccurrenceID *uuid.UUID
	CircleID               *uuid.UUID
	CircleInviteID         *uuid.UUID
	RecapID                *uuid.UUID
	ResurfacingID          *uuid.UUID

	Payload json.RawMessage

	// Earliest permitted delivery instant — quiet hours and the digest's
	// fixed daily hour are both expressed by pushing this forward.
	DeliverAfter time.Time

	DeliveredAt *time.Time
	ReadAt      *time.Time
	CreatedAt   time.Time
}
