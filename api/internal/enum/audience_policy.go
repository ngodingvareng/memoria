package enum

// AudiencePolicy controls who may reach a user through a given social
// affordance (see users.mention_policy / users.circle_invite_policy).
type AudiencePolicy string

const (
	AudiencePolicyAnyone        AudiencePolicy = "anyone"
	AudiencePolicyCircleMembers AudiencePolicy = "circle_members"
	AudiencePolicyNobody        AudiencePolicy = "nobody"
)
