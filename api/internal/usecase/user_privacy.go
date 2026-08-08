package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

// This file holds the privacy-primitive interfaces shared by every
// usecase that needs to gate an interaction on Circle Invite, Mention,
// or Response's rules (mention_usecase.go, circle_invite_usecase.go,
// response_access.go). Declared once here rather than per-consumer
// because all three need the exact identical signature (CODING_STANDARDS.md
// §5) — see user_privacy_repository.go for the single implementation.
//
// Known/Block/Mute have no CRUD usecase of their own yet: only these
// read-only checks are wired in, which is enough to make
// mention_policy/circle_invite_policy and the blocking gate work.
// Letting a user actually manage their own Known/Block/Mute lists is a
// separate, not-yet-built "Privacy & Control" feature.

// UserBlockChecker gates every view/mention/comment/reaction access
// check (FEATURES.md, Privacy & Control): blocking is symmetric in
// effect, so this must always be checked both ways.
type UserBlockChecker interface {
	IsBlockedEitherDirection(ctx context.Context, userA, userB uuid.UUID) (bool, error)
}

// UserKnownChecker resolves audience_policy = "known" for
// users.mention_policy and users.circle_invite_policy. Argument order is
// not symmetric: ownerUserID is the person whose policy is being
// enforced, otherUserID is the one trying to reach them.
type UserKnownChecker interface {
	IsKnownTo(ctx context.Context, ownerUserID, otherUserID uuid.UUID) (bool, error)
}

// UserMuteChecker is a view filter only, never an access gate
// (FEATURES.md, Response Terms: "Muted users aren't restricted at
// all... Muting just hides their activity from the muter's own view").
type UserMuteChecker interface {
	IsMuted(ctx context.Context, muterUserID, mutedUserID uuid.UUID) (bool, error)
}

// UserPolicyReader is the minimal capability circle_invite_usecase.go
// and mention_usecase.go need from the user domain: resolving a
// username (both flows start from "search a username") and reading
// audience_policy settings before deciding whether an interaction is
// allowed. Satisfied structurally by the existing UserRepository (see
// auth_usecase.go) — no new repository needed.
type UserPolicyReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
}
