package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

// GetInviteLink implements [CircleInviteUsecase]. Viewing the existing
// link is available to any active member, not just invite-privileged
// ones — generating/rotating/toggling approval is the privileged action.
func (u *circleInviteUsecase) GetInviteLink(ctx context.Context, circleID, userID uuid.UUID) (*entity.CircleInvite, error) {
	if _, err := u.circles.GetActiveMember(ctx, circleID, userID); err != nil {
		return nil, fmt.Errorf("checking circle membership: %w", err)
	}
	invite, err := u.repo.GetActiveLinkByCircleID(ctx, circleID)
	if err != nil {
		return nil, fmt.Errorf("getting circle invite link: %w", err)
	}
	return invite, nil
}

// CreateOrRotateInviteLink implements [CircleInviteUsecase]. Rotation is
// the only revocation mechanism a link has (FEATURES.md, Circle
// Invite): revoking the old one and creating the new one run in the
// same transaction, since uq_circle_invites_active_link permits only
// one live link per Circle at a time.
func (u *circleInviteUsecase) CreateOrRotateInviteLink(ctx context.Context, input CreateOrRotateInviteLinkInput) (*CreateOrRotateInviteLinkResult, error) {
	if err := u.requireCanInvite(ctx, input.CircleID, input.UserID); err != nil {
		return nil, err
	}

	rawToken, err := u.tokens.Generate()
	if err != nil {
		return nil, fmt.Errorf("generating invite link token: %w", err)
	}
	tokenHash := u.tokens.Hash(rawToken)

	var invite *entity.CircleInvite
	err = u.repo.WithTransaction(ctx, func(tx CircleInviteRepository) error {
		if _, err := tx.RevokeActiveLink(ctx, input.CircleID); err != nil {
			return err
		}
		var createErr error
		invite, createErr = tx.CreateLink(ctx, input.CircleID, input.UserID, tokenHash, input.RequiresApproval)
		return createErr
	})
	if err != nil {
		return nil, fmt.Errorf("creating circle invite link: %w", err)
	}
	return &CreateOrRotateInviteLinkResult{Invite: invite, RawToken: rawToken}, nil
}

// SetInviteLinkRequiresApproval implements [CircleInviteUsecase].
// Requests already open when this toggles stay open either way —
// turning approval off does not admit them retroactively (FEATURES.md,
// Circle Invite).
func (u *circleInviteUsecase) SetInviteLinkRequiresApproval(ctx context.Context, circleID, userID uuid.UUID, requiresApproval bool) (*entity.CircleInvite, error) {
	if err := u.requireCanInvite(ctx, circleID, userID); err != nil {
		return nil, err
	}
	invite, err := u.repo.SetLinkRequiresApproval(ctx, circleID, requiresApproval)
	if err != nil {
		return nil, fmt.Errorf("setting circle invite link approval requirement: %w", err)
	}
	return invite, nil
}
