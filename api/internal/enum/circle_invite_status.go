package enum

// CircleInviteStatus is split by kind (the DB enforces the split via
// chk_circle_invites_status_by_kind): a Username invite runs Pending ->
// Accepted/Declined, or ends as Revoked/Expired, while a Link is only
// ever Active or Revoked. A Link is not pending anything, and it can
// neither be declined nor expire.
type CircleInviteStatus string

const (
	CircleInviteStatusPending  CircleInviteStatus = "pending"
	CircleInviteStatusActive   CircleInviteStatus = "active"
	CircleInviteStatusAccepted CircleInviteStatus = "accepted"
	CircleInviteStatusDeclined CircleInviteStatus = "declined"
	CircleInviteStatusRevoked  CircleInviteStatus = "revoked"
	CircleInviteStatusExpired  CircleInviteStatus = "expired"
)
