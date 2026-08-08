package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// Resurfacing is Echoes, one row per Moment actually surfaced, so the
// same flashback isn't served twice and declining is always reachable
// directly from the resurfaced item (FEATURES.md, Echoes).
type Resurfacing struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	MomentID uuid.UUID
	Reason   enum.ResurfacingReason

	// The user's local date this echo was served for.
	SurfacedOn time.Time

	OpenedAt    *time.Time
	DismissedAt *time.Time
	CreatedAt   time.Time
}
