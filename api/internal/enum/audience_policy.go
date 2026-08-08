package enum

// AudiencePolicy controls who may reach a user through a given social
// affordance (see users.mention_policy / users.circle_invite_policy).
//
// Known resolves to a union, not just the user_knows table: someone the
// user marked (see entity.UserKnown), OR someone they already share an
// active Circle with. The second half is load-bearing — without it the
// tier deadlocks, since you could only be invited to your first Circle
// by someone you already shared a Circle with. Evaluate it with the
// IsUserKnownTo query rather than by hand.
type AudiencePolicy string

const (
	AudiencePolicyAnyone AudiencePolicy = "anyone"
	AudiencePolicyKnown  AudiencePolicy = "known"
	AudiencePolicyNobody AudiencePolicy = "nobody"
)
