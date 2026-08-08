package enum

type CommitmentOutcome string

const (
	CommitmentOutcomeDue    CommitmentOutcome = "due"
	CommitmentOutcomeKept   CommitmentOutcome = "kept"
	CommitmentOutcomeMissed CommitmentOutcome = "missed"
)
