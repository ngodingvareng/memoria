package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// DeleteMention implements [MentionUsecase].
func (u *mentionUsecase) DeleteMention(ctx context.Context, mentionID, momentID, ownerUserID uuid.UUID) error {
	if _, err := u.moments.GetByID(ctx, momentID, ownerUserID); err != nil {
		return fmt.Errorf("checking moment ownership: %w", err)
	}
	if err := u.repo.Delete(ctx, mentionID, momentID); err != nil {
		return fmt.Errorf("deleting mention: %w", err)
	}
	return nil
}
