package usecase

import (
	"context"
	"fmt"

	"github.com/ngodingvareng/memoria/internal/entity"
)

// CreateCircle implements [CircleUsecase]. Seats the creator as admin in
// the same transaction as the insert (FEATURES.md, Circle Permissions:
// "The user who creates a circle becomes its admin").
func (u *circleUsecase) CreateCircle(ctx context.Context, input CreateCircleInput) (*entity.Circle, error) {
	circle := &entity.Circle{
		Name:        input.Name,
		Description: input.Description,
		ColorHex:    input.ColorHex,
		ImagePath:   input.ImagePath,
	}

	var created *entity.Circle
	err := u.repo.WithTransaction(ctx, func(tx CircleRepository) error {
		var txErr error
		created, txErr = tx.Create(ctx, circle)
		if txErr != nil {
			return txErr
		}
		if _, txErr = tx.AddCreatorAsAdmin(ctx, created.ID, input.UserID); txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("creating circle: %w", err)
	}
	return created, nil
}
