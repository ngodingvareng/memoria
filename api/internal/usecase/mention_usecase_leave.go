package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// LeaveMention implements [MentionUsecase].
func (u *mentionUsecase) LeaveMention(ctx context.Context, momentID, mentionedUserID uuid.UUID) error {
	if err := u.repo.Remove(ctx, momentID, mentionedUserID); err != nil {
		return fmt.Errorf("leaving mention: %w", err)
	}
	return nil
}
