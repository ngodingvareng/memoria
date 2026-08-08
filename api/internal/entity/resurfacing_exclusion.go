package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// ResurfacingExclusion permanently excludes content from all automatic
// resurfacing — Echoes, Recap, notifications, algorithmic Album
// groupings, and adherence summaries. Excluded content is never deleted
// or hidden, only kept from being brought up unprompted (FEATURES.md,
// Resurfacing Controls). Which fields are meaningful depends on Scope.
type ResurfacingExclusion struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Scope  enum.ResurfacingScope

	// Scope == ResurfacingScopeMoment.
	MomentID *uuid.UUID

	// Scope == ResurfacingScopePerson.
	PersonUserID         *uuid.UUID
	IncludeSharedCircles bool

	// Scope == ResurfacingScopeDateRange. Compared against stored local
	// dates (Moment.OccurredOn / CommitmentOccurrence.DueOn).
	FromDate *time.Time
	ToDate   *time.Time

	CreatedAt time.Time
}
