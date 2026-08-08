package enum

// CircleInviteKind is deliberately asymmetric: Username invites are
// addressed to one specific user who may join directly, Link invites are
// a shareable token that requires an invite-privileged member's approval
// (see CircleJoinRequestStatus).
type CircleInviteKind string

const (
	CircleInviteKindUsername CircleInviteKind = "username"
	CircleInviteKindLink     CircleInviteKind = "link"
)
