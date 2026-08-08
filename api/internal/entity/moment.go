package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// Moment has no state of its own (FEATURES.md: "A manually captured
// Moment has no state") — Commitment outcomes (due/kept/missed) live on
// CommitmentOccurrence instead, one migration later, and never touch
// this struct.
//
// moments.search_document has no field here: it's a tsvector search
// index, not a domain value, and stays a repository/mapper concern.
type Moment struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	ThreadID *uuid.UUID
	Origin   enum.MomentOrigin

	// ----- The Time table (FEATURES.md) -----

	// When it actually happened. The only one of the four timestamps
	// the user may edit.
	OccurredAt time.Time

	// The same instant as wall-clock time where it happened, plus the
	// UTC offset in effect there — never the viewer's current zone.
	OccurredLocal            time.Time
	OccurredUTCOffsetMinutes int16

	// The stored local calendar date. What Echoes match on and what the
	// Rhythms heatmap buckets by — never OccurredAt directly.
	OccurredOn time.Time

	// When the user actually wrote it down. Never editable.
	RecordedAt time.Time

	// Time to Tell, materialized: RecordedAt - OccurredAt. Meaningful
	// only for Origin == MomentOriginManual.
	SettlingTime time.Duration

	// ----- Content, all optional -----
	Note      *string
	ColorHex  *string
	PlaceName *string
	Latitude  *float64
	Longitude *float64

	// Client-generated id for idempotent offline sync.
	ClientID *uuid.UUID

	// Owner-only. Never exposed to anyone but the owner; not a read
	// receipt and not a view count.
	LastViewedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
