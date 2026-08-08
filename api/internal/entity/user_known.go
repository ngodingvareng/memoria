package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserKnown records that KnowerUserID marked KnownUserID as someone they
// know. One-directional and silent: the known user is not notified, does
// not confirm, and cannot see the row.
//
// The direction is a safety property, not a detail. A row grants the
// known user access to the knower, so marking someone gains the marker
// nothing and only the protected party's decision is ever recorded —
// which is why no confirmation step exists (see enum.AudiencePolicy).
type UserKnown struct {
	KnowerUserID uuid.UUID
	KnownUserID  uuid.UUID
	CreatedAt    time.Time
}
