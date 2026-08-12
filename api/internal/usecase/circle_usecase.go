package usecase

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// --- Repository interface, defined here per the Dependency Rule ---

type CircleRepository interface {
	// Create inserts the Circle row only — the caller is responsible for
	// also seating the creator as admin via AddCreatorAsAdmin, in the
	// same transaction (see CircleUsecase.CreateCircle).
	Create(ctx context.Context, circle *entity.Circle) (*entity.Circle, error)
	AddCreatorAsAdmin(ctx context.Context, circleID, userID uuid.UUID) (*entity.CircleMember, error)
	// Update and Dissolve are admin-only, enforced in the WHERE clause —
	// a non-admin's attempt and a wrong/foreign id are indistinguishable
	// from here, both surface as errs.ErrNotFound.
	Update(ctx context.Context, circle *entity.Circle, userID uuid.UUID) (*entity.Circle, error)
	// UpdateImagePath is the same admin-only gate as Update, scoped to
	// just the profile image.
	UpdateImagePath(ctx context.Context, id, userID uuid.UUID, imagePath *string) (*entity.Circle, error)
	// Dissolve is a sqlc :exec query (no rows-affected count), so a
	// WHERE clause matching zero rows is a silent no-op here.
	Dissolve(ctx context.Context, id, userID uuid.UUID) error
	// GetByID returns errs.ErrNotFound unless userID is an active member.
	GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Circle, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Circle, error)

	ListMembers(ctx context.Context, circleID uuid.UUID) ([]*entity.CircleMember, error)
	GetActiveMember(ctx context.Context, circleID, userID uuid.UUID) (*entity.CircleMember, error)
	// IsSoleActiveAdmin reports whether userID is circleID's only active
	// admin — checked before Leave/UpdateMemberRole would otherwise let
	// that admin self-leave or self-demote and orphan the Circle.
	IsSoleActiveAdmin(ctx context.Context, circleID, userID uuid.UUID) (bool, error)
	Leave(ctx context.Context, circleID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, circleID, userID, removedByUserID uuid.UUID) error
	UpdateMemberRole(
		ctx context.Context,
		circleID, userID, changedByUserID uuid.UUID,
		role enum.CircleRole,
	) (*entity.CircleMember, error)
	UpdateMemberPermissions(
		ctx context.Context,
		circleID, userID, changedByUserID uuid.UUID,
		canInvite, canCapture bool,
	) (*entity.CircleMember, error)

	// ListSharedCircleIDs is unused by CircleUsecase itself — declared
	// here only because circleRepository is the single implementation
	// backing both CircleRepository and CircleShareChecker below.
	ListSharedCircleIDs(ctx context.Context, userA, userB uuid.UUID) ([]uuid.UUID, error)

	WithTransaction(ctx context.Context, fn func(CircleRepository) error) error
}

// CircleAccessChecker is the minimal capability CircleInviteUsecase and
// CircleJoinRequestUsecase need: confirming the requester is an active
// member (and, at the call site, checking their can_invite flag) before
// letting them act on a Circle's invites/join requests. Method name
// matches CircleRepository.GetActiveMember on purpose — see
// CODING_STANDARDS.md §5.
type CircleAccessChecker interface {
	GetActiveMember(ctx context.Context, circleID, userID uuid.UUID) (*entity.CircleMember, error)
}

// CircleShareChecker is the minimal capability MentionUsecase needs:
// which Circles two users both actively belong to, for the mention
// flow's optional "Share to circle too?" step (FEATURES.md, Mention).
type CircleShareChecker interface {
	ListSharedCircleIDs(ctx context.Context, userA, userB uuid.UUID) ([]uuid.UUID, error)
}

// --- Inputs / outputs ---

type CreateCircleInput struct {
	UserID      uuid.UUID
	Name        string
	Description *string
	ColorHex    *string
	ImagePath   *string
}

type UpdateCircleInput struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description *string
	ColorHex    *string
	ImagePath   *string
}

type UpdateCircleMemberRoleInput struct {
	CircleID        uuid.UUID
	UserID          uuid.UUID
	ChangedByUserID uuid.UUID
	Role            enum.CircleRole
}

type UpdateCircleMemberPermissionsInput struct {
	CircleID        uuid.UUID
	UserID          uuid.UUID
	ChangedByUserID uuid.UUID
	CanInvite       bool
	CanCapture      bool
}

// --- Usecase ---

type CircleUsecase interface {
	CreateCircle(ctx context.Context, input CreateCircleInput) (*entity.Circle, error)
	UpdateCircle(ctx context.Context, input UpdateCircleInput) (*entity.Circle, error)
	DissolveCircle(ctx context.Context, id, userID uuid.UUID) error
	GetCircle(ctx context.Context, id, userID uuid.UUID) (*entity.Circle, error)
	ListMyCircles(ctx context.Context, userID uuid.UUID) ([]*entity.Circle, error)

	ListMembers(ctx context.Context, circleID, userID uuid.UUID) ([]*entity.CircleMember, error)
	LeaveCircle(ctx context.Context, circleID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, circleID, targetUserID, removedByUserID uuid.UUID) error
	UpdateMemberRole(ctx context.Context, input UpdateCircleMemberRoleInput) (*entity.CircleMember, error)
	UpdateMemberPermissions(ctx context.Context, input UpdateCircleMemberPermissionsInput) (*entity.CircleMember, error)

	// UploadCircleImage replaces the Circle's profile photo, best-effort
	// deleting whichever one it replaces. Admin-only, enforced the same
	// way as UpdateCircle.
	UploadCircleImage(ctx context.Context, input UploadCircleImageInput) (*entity.Circle, error)
}

type UploadCircleImageInput struct {
	CircleID    uuid.UUID
	UserID      uuid.UUID
	FileName    string // client-supplied, used only for its extension
	ContentType string
	Size        int64
	Body        io.Reader
}

type circleUsecase struct {
	repo    CircleRepository
	storage ProfileImageStorage
}

func NewCircleUsecase(repo CircleRepository, storage ProfileImageStorage) CircleUsecase {
	return &circleUsecase{repo: repo, storage: storage}
}
