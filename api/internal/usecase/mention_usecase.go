package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

// --- Repository interface, defined here per the Dependency Rule ---

type MentionRepository interface {
	Create(ctx context.Context, momentID, mentionedUserID uuid.UUID, displayName string) (*entity.MomentMention, error)
	ListByMomentID(ctx context.Context, momentID uuid.UUID) ([]*entity.MomentMention, error)
	ListMentionedMoments(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.Moment, error)
	// Remove is a sqlc :exec query — a WHERE clause matching zero rows
	// (already removed, never mentioned) is a silent no-op.
	Remove(ctx context.Context, momentID, mentionedUserID uuid.UUID) error
	Delete(ctx context.Context, id, momentID uuid.UUID) error

	// ShareToCircle returns nil, nil when already shared — idempotent,
	// not an error (FEATURES.md, Mention).
	ShareToCircle(ctx context.Context, momentID, circleID, sharedByUserID uuid.UUID) (*entity.MomentCircle, error)
	ListSharedCircleIDs(ctx context.Context, momentID uuid.UUID) ([]uuid.UUID, error)
	UnshareFromCircle(ctx context.Context, momentID, circleID uuid.UUID) error
}

// UserSearcher backs the mention-search typeahead (FEATURES.md, Mention)
// — a candidate list of discoverable usernames matching a prefix, which
// SearchMentionableUsers then narrows by mention_policy/blocking, the
// same way CreateMention gates a single resolved username. Satisfied
// structurally by the existing UserRepository (see user_repository.go's
// SearchByUsernamePrefix) — no new repository needed.
type UserSearcher interface {
	SearchByUsernamePrefix(
		ctx context.Context,
		excludeUserID uuid.UUID,
		query string,
		limit int32,
	) ([]*entity.User, error)
}

// --- Inputs / outputs ---

type CreateMentionInput struct {
	MomentID uuid.UUID
	// OwnerUserID must own the Moment — only the owner adds mentions to
	// it (FEATURES.md, Mention: naming another person present).
	OwnerUserID uuid.UUID
	Username    string
}

type ShareMomentToCircleInput struct {
	MomentID uuid.UUID
	CircleID uuid.UUID
	// UserID must own the Moment and be an active member of the target
	// Circle.
	UserID uuid.UUID
}

type ListMentionedMomentsInput struct {
	UserID   uuid.UUID
	Page     int32
	PageSize int32
}

// CreateMentionResult is CreateMention's return: the new mention, plus
// which Circles the owner and the mentioned user both actively belong
// to — the candidate set for the mention flow's optional "Share to
// circle too?" step (FEATURES.md, Mention), computed in the same call
// so the frontend can offer it immediately with no extra request.
type CreateMentionResult struct {
	Mention         *entity.MomentMention
	SharedCircleIDs []uuid.UUID
}

// --- Usecase ---

type MentionUsecase interface {
	CreateMention(ctx context.Context, input CreateMentionInput) (*CreateMentionResult, error)
	ListMentions(ctx context.Context, momentID, ownerUserID uuid.UUID) ([]*entity.MomentMention, error)
	// LeaveMention is the mentioned user removing themselves — no
	// notification to the owner (FEATURES.md, Leaving a mention).
	LeaveMention(ctx context.Context, momentID, mentionedUserID uuid.UUID) error
	DeleteMention(ctx context.Context, mentionID, momentID, ownerUserID uuid.UUID) error
	ListMentionedMoments(ctx context.Context, input ListMentionedMomentsInput) (*MomentListResult, error)
	// SearchMentionableUsers backs the mention-search typeahead: a
	// prefix match over discoverable usernames the requester is
	// currently allowed to mention (FEATURES.md, Mention + Privacy &
	// Control). Empty query returns an empty slice — no "browse
	// everyone" mode.
	SearchMentionableUsers(ctx context.Context, requestingUserID uuid.UUID, query string) ([]*entity.User, error)

	ShareToCircle(ctx context.Context, input ShareMomentToCircleInput) (*entity.MomentCircle, error)
	UnshareFromCircle(ctx context.Context, momentID, circleID, userID uuid.UUID) error
	// ListSharedCircles is which Circles a personal Moment is currently
	// shared to — owner-only, backs the "manage sharing" surface.
	ListSharedCircles(ctx context.Context, momentID, ownerUserID uuid.UUID) ([]uuid.UUID, error)
}

type mentionUsecase struct {
	repo          MentionRepository
	moments       MomentAccessChecker
	circles       CircleAccessChecker
	circleShares  CircleShareChecker
	users         UserPolicyReader
	knowns        UserKnownChecker
	blocks        UserBlockChecker
	searcher      UserSearcher
	notifications NotificationCreator
}

func NewMentionUsecase(
	repo MentionRepository,
	moments MomentAccessChecker,
	circles CircleAccessChecker,
	circleShares CircleShareChecker,
	users UserPolicyReader,
	knowns UserKnownChecker,
	blocks UserBlockChecker,
	searcher UserSearcher,
	notifications NotificationCreator,
) MentionUsecase {
	return &mentionUsecase{
		repo:          repo,
		moments:       moments,
		circles:       circles,
		circleShares:  circleShares,
		users:         users,
		knowns:        knowns,
		blocks:        blocks,
		searcher:      searcher,
		notifications: notifications,
	}
}
