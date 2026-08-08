package enum

// CommitmentStrictness affects only the confirmation window and how
// adherence is counted — never how a Moment is stored (FEATURES.md,
// Strictness).
type CommitmentStrictness string

const (
	CommitmentStrictnessGentle   CommitmentStrictness = "gentle"
	CommitmentStrictnessStandard CommitmentStrictness = "standard"
	CommitmentStrictnessStrict   CommitmentStrictness = "strict"
)
