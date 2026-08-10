package entity

import (
	"time"

	"github.com/google/uuid"
)

// AlbumImage is a single photo attachment flattened out of its parent
// Moment for the Album feed (FEATURES.md, Looking Back) — one row per
// image, not one row per Moment, so the Album can render "every photo
// attachment from every Moment" without an N+1 per-Moment image fetch.
// It denormalizes the parent Moment's OccurredLocal/
// OccurredUTCOffsetMinutes (the Album's sort/group key) and, for a
// Circle Album row, whether/by whom it was shared in — fields plain
// MomentImage has no reason to carry.
type AlbumImage struct {
	ID                       uuid.UUID
	MomentID                 uuid.UUID
	ImagePath                string
	ImageAlt                 *string
	Width                    *int32
	Height                   *int32
	OccurredLocal            time.Time
	OccurredUTCOffsetMinutes int16

	// IsShared and the two SharedBy* fields are always false/nil for a
	// Personal Album row. For a Circle Album row, IsShared is true iff
	// this image's Moment reached the Circle via the mention/share flow
	// rather than being captured natively into one of the Circle's own
	// collaborative Threads — the "from @username" attribution label
	// (FEATURES.md, Looking Back) renders only when this is true.
	IsShared         bool
	SharedByUserID   *uuid.UUID
	SharedByUsername *string
}
