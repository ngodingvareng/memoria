package enum

type NotificationKind string

const (
	// Commitment reminders — only for Threads carrying a Commitment.
	NotificationKindCommitmentUpcoming NotificationKind = "commitment_upcoming"
	NotificationKindMomentDue          NotificationKind = "moment_due"
	NotificationKindMomentMissed       NotificationKind = "moment_missed"

	// Gifts, delivered when ready.
	NotificationKindEchoReady      NotificationKind = "echo_ready"
	NotificationKindRecapGenerated NotificationKind = "recap_generated"

	// Someone needs you, delivered immediately.
	NotificationKindCircleInviteReceived      NotificationKind = "circle_invite_received"
	NotificationKindCircleJoinRequestReceived NotificationKind = "circle_join_request_received"
	NotificationKindMentionedInMoment         NotificationKind = "mentioned_in_moment"

	// Circle changes: immediate too, but nothing is being asked. Both
	// mean the user is now in a Circle they were not in before, which
	// changes who can see their Moments.
	//
	// There is deliberately no rejected/declined counterpart to either:
	// Memoria does not manufacture a moment of rejection.
	NotificationKindAddedToCircle             NotificationKind = "added_to_circle"
	NotificationKindCircleJoinRequestApproved NotificationKind = "circle_join_request_approved"

	// Responses, batched into a single daily digest.
	NotificationKindResponseDigest NotificationKind = "response_digest"
)
