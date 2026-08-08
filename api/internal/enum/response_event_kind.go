package enum

type ResponseEventKind string

const (
	ResponseEventKindComment  ResponseEventKind = "comment"
	ResponseEventKindReaction ResponseEventKind = "reaction"
)
