package usecase

import (
	"context"
	"fmt"

	"github.com/ngodingvareng/memoria/internal/entity"
)

// UpdateCircle implements [CircleUsecase]. Full-representation update,
// same PUT convention as ThreadUsecase.UpdateThread.
func (u *circleUsecase) UpdateCircle(ctx context.Context, input UpdateCircleInput) (*entity.Circle, error) {
	circle := &entity.Circle{
		ID:          input.ID,
		Name:        input.Name,
		Description: input.Description,
		ColorHex:    input.ColorHex,
		ImagePath:   input.ImagePath,
	}

	updated, err := u.repo.Update(ctx, circle, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("updating circle: %w", err)
	}
	return updated, nil
}
