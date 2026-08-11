package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// DissolveCircle implements [CircleUsecase].
func (u *circleUsecase) DissolveCircle(ctx context.Context, id, userID uuid.UUID) error {
	if err := u.repo.Dissolve(ctx, id, userID); err != nil {
		return fmt.Errorf("dissolving circle: %w", err)
	}
	return nil
}
