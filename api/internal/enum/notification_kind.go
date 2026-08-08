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
	NotificationKindCircleInviteReceived NotificationKind = "circle_invite_received"
	NotificationKindMentionedInMoment    NotificationKind = "mentioned_in_moment"

	// Responses, batched into a single daily digest.
	NotificationKindResponseDigest NotificationKind = "response_digest"
)
