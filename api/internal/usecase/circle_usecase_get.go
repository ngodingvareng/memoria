package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

// GetCircle implements [CircleUsecase].
func (u *circleUsecase) GetCircle(ctx context.Context, id, userID uuid.UUID) (*entity.Circle, error) {
	circle, err := u.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("getting circle: %w", err)
	}
	return circle, nil
}
