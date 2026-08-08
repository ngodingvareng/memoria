package enum

// MomentOrigin is provenance only — it never gates access and is never
// used to implement the Commitment Firewall (see moments.origin).
type MomentOrigin string

const (
	MomentOriginManual     MomentOrigin = "manual"
	MomentOriginCommitment MomentOrigin = "commitment"
	MomentOriginImport     MomentOrigin = "import"
)
