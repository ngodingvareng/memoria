package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// CommitmentOccurrence is the Commitment Firewall in structural form:
// one row per scheduled occurrence, entirely separate from Moment. A
// missed occurrence has no MomentID and no moments row, so Album,
// Echoes, Search, and a Thread's timeline — which all query Moment —
// cannot surface it (FEATURES.md, Commitment Firewall).
type CommitmentOccurrence struct {
	ID           uuid.UUID
	CommitmentID uuid.UUID
	ThreadID     uuid.UUID
	UserID       uuid.UUID

	Outcome enum.CommitmentOutcome

	// When it came due, with the same instant/local/offset triple the
	// archive uses.
	DueAt               time.Time
	DueLocal            time.Time
	DueUTCOffsetMinutes int16
	DueOn               time.Time

	// Frozen at generation from the Commitment's window and strictness
	// at that time — editing a Commitment affects future occurrences
	// only and never rewrites this one.
	ExpiresAt       time.Time
	StrictnessAtDue enum.CommitmentStrictness

	// When the user actually answered. May be after ExpiresAt: a missed
	// occurrence can still be recorded later, at any time.
	CapturedAt *time.Time

	// Whether the answer arrived after ExpiresAt. Under 'standard' a
	// late recording still counts as kept but is noted as late; under
	// 'strict' only on-time recordings count.
	WasLate bool

	// The Moment recorded in answer to this occurrence, if any.
	MomentID *uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
}
