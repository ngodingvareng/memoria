package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// Recap is a monthly or yearly period summary. Payload is a cache of an
// algorithmic rendering, not a source of truth — it is fully regenerable
// from the archive, and generation must honour ResurfacingExclusion,
// including for the aggregate adherence figures it's allowed to include.
type Recap struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Period enum.RecapPeriod

	// Local calendar bounds, inclusive.
	PeriodStart time.Time
	PeriodEnd   time.Time

	Payload     json.RawMessage
	GeneratedAt time.Time
	OpenedAt    *time.Time
}
