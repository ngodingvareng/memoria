package enum

// ResurfacingReason is why a Moment was surfaced as an Echo. Date
// matching (SameDay) is the crudest signal and deliberately not the
// only one (FEATURES.md, Echoes).
type ResurfacingReason string

const (
	ResurfacingReasonSameDay ResurfacingReason = "same_day"
	ResurfacingReasonDormant ResurfacingReason = "dormant"
	ResurfacingReasonNearby  ResurfacingReason = "nearby"
	ResurfacingReasonPerson  ResurfacingReason = "person"
)
