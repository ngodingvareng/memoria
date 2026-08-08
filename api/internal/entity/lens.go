package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// Lens is a saved, named, living view over the archive — a way of
// looking, not a container you fill (FEATURES.md, Lens). There is no
// join table: its contents are always the result of evaluating the
// predicate fields below against Moment at read time. Which fields are
// meaningful depends on Kind.
type Lens struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	Name     string
	Kind     enum.LensKind
	ColorHex *string

	// Kind == LensKindPerson.
	PersonUserID *uuid.UUID

	// Kind == LensKindPlace.
	PlaceName    *string
	Latitude     *float64
	Longitude    *float64
	RadiusMeters *int32

	// Kind == LensKindPeriod. Local calendar dates, matched against
	// Moment.OccurredOn so the Lens doesn't shift when the user travels.
	FromDate *time.Time
	ToDate   *time.Time

	// Kind == LensKindQuery.
	QueryText *string

	PinnedAt  *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
