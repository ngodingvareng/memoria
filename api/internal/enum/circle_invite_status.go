package enum

type CircleInviteStatus string

const (
	CircleInviteStatusPending  CircleInviteStatus = "pending"
	CircleInviteStatusAccepted CircleInviteStatus = "accepted"
	CircleInviteStatusDeclined CircleInviteStatus = "declined"
	CircleInviteStatusRevoked  CircleInviteStatus = "revoked"
	CircleInviteStatusExpired  CircleInviteStatus = "expired"
)
