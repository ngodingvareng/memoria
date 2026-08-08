package enum

// LensKind selects which of lenses' mutually-exclusive predicate shapes
// is in effect: person, place, period, or query.
type LensKind string

const (
	LensKindPerson LensKind = "person"
	LensKindPlace  LensKind = "place"
	LensKindPeriod LensKind = "period"
	LensKindQuery  LensKind = "query"
)
