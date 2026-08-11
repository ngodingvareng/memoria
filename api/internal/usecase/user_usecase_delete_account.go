package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/enum"
)

// DeleteAccount implements [UserUsecase].
func (u *userUsecase) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	user, err := u.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("loading user before deletion: %w", err)
	}

	var imagePaths []string
	err = u.accountDeletion.WithTransaction(ctx, func(repos AccountDeletionRepositories) error {
		// Gather storage cleanup targets first, while everything this
		// user owns is still visible — SoftDeleteAllByUserID below
		// makes both queries return nothing.
		momentImagePaths, err := repos.MomentImage.ListImagePathsByOwnerID(ctx, userID)
		if err != nil {
			return fmt.Errorf("listing moment image paths: %w", err)
		}
		threadImagePaths, err := repos.ThreadImage.ListImagePathsByOwnerID(ctx, userID)
		if err != nil {
			return fmt.Errorf("listing thread image paths: %w", err)
		}
		imagePaths = append(imagePaths, momentImagePaths...)
		imagePaths = append(imagePaths, threadImagePaths...)
		if user.ImagePath != nil {
			if key, ok := strings.CutPrefix(*user.ImagePath, u.storage.PublicURL("")); ok {
				imagePaths = append(imagePaths, key)
			}
		}

		// Anonymize before the Moments disappear — harmless for
		// comments/reactions on the user's own (about-to-vanish)
		// Moments, required for the ones on other people's (FEATURES.md,
		// Lifecycle & Deletion).
		if err := repos.Comment.AnonymizeByUserID(ctx, userID); err != nil {
			return fmt.Errorf("anonymizing comments: %w", err)
		}
		if err := repos.Reaction.AnonymizeByUserID(ctx, userID); err != nil {
			return fmt.Errorf("anonymizing reactions: %w", err)
		}

		if err := repos.Moment.SoftDeleteAllByUserID(ctx, userID); err != nil {
			return fmt.Errorf("soft-deleting moments: %w", err)
		}
		if err := repos.Thread.SoftDeleteAllByUserID(ctx, userID); err != nil {
			return fmt.Errorf("soft-deleting threads: %w", err)
		}

		if err := leaveOrHandoffCircles(ctx, repos.Circle, userID); err != nil {
			return fmt.Errorf("resolving circle memberships: %w", err)
		}

		// UpdateImagePath's query is guarded by deleted_at IS NULL like
		// every other write in this codebase, so it must run before
		// SoftDelete — after it, the same call would silently match
		// zero rows and surface as errs.ErrNotFound.
		if _, err := repos.User.UpdateImagePath(ctx, userID, nil); err != nil {
			return fmt.Errorf("clearing profile image: %w", err)
		}
		if err := repos.User.SoftDelete(ctx, userID); err != nil {
			return fmt.Errorf("soft-deleting user: %w", err)
		}

		if err := repos.RefreshToken.RevokeAllByUserID(ctx, userID); err != nil {
			return fmt.Errorf("revoking sessions: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("deleting account: %w", err)
	}

	// Storage cleanup happens after the transaction commits, best
	// effort — an orphaned file is a cleanup job's problem, not a
	// reason to fail an otherwise-successful account deletion (same
	// reasoning as UploadProfileImage's own best-effort old-image
	// cleanup).
	for _, key := range imagePaths {
		if err := u.storage.Delete(context.WithoutCancel(ctx), key); err != nil {
			slog.WarnContext(ctx, "failed to delete storage object during account deletion", "key", key, "error", err)
		}
	}

	return nil
}

// leaveOrHandoffCircles resolves every Circle userID is still an active
// member of, before their own membership ends. A plain member (or an
// admin who isn't the sole one) just leaves — nothing they contributed
// is removed (FEATURES.md, Circle Membership). A sole admin can't just
// leave: the longest-tenured other active member (ListMembers is
// ordered by joined_at) is promoted first so the Circle keeps
// functioning; if no other active member exists, the Circle is
// dissolved instead, since there's no one left to hand it to.
func leaveOrHandoffCircles(ctx context.Context, circles CircleRepository, userID uuid.UUID) error {
	memberCircles, err := circles.ListByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("listing circle memberships: %w", err)
	}

	for _, circle := range memberCircles {
		isSoleAdmin, err := circles.IsSoleActiveAdmin(ctx, circle.ID, userID)
		if err != nil {
			return fmt.Errorf("checking sole admin status for circle %s: %w", circle.ID, err)
		}
		if !isSoleAdmin {
			if err := circles.Leave(ctx, circle.ID, userID); err != nil {
				return fmt.Errorf("leaving circle %s: %w", circle.ID, err)
			}
			continue
		}

		members, err := circles.ListMembers(ctx, circle.ID)
		if err != nil {
			return fmt.Errorf("listing members for circle %s: %w", circle.ID, err)
		}
		var successor *uuid.UUID
		for _, member := range members {
			if member.UserID != userID {
				id := member.UserID
				successor = &id
				break
			}
		}

		if successor == nil {
			if err := circles.Dissolve(ctx, circle.ID, userID); err != nil {
				return fmt.Errorf("dissolving circle %s: %w", circle.ID, err)
			}
			continue
		}

		if _, err := circles.UpdateMemberRole(ctx, circle.ID, *successor, userID, enum.CircleRoleAdmin); err != nil {
			return fmt.Errorf("promoting successor in circle %s: %w", circle.ID, err)
		}
		if err := circles.Leave(ctx, circle.ID, userID); err != nil {
			return fmt.Errorf("leaving circle %s after handoff: %w", circle.ID, err)
		}
	}

	return nil
}
