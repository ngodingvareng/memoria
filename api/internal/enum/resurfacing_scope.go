package enum

// ResurfacingScope is the granularity at which a user can permanently
// exclude content from all automatic resurfacing (FEATURES.md,
// Resurfacing Controls).
type ResurfacingScope string

const (
	ResurfacingScopeMoment    ResurfacingScope = "moment"
	ResurfacingScopePerson    ResurfacingScope = "person"
	ResurfacingScopeDateRange ResurfacingScope = "date_range"
)
