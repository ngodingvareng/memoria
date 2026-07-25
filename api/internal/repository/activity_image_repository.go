package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.ActivityImageRepository = (*activityImageRepository)(nil)

type activityImageRepository struct {
	q *db.Queries
}

func NewActivityImageRepository(pool *pgxpool.Pool) *activityImageRepository {
	return &activityImageRepository{q: db.New(pool)}
}

// Create implements [usecase.ActivityImageRepository].
func (r *activityImageRepository) Create(ctx context.Context, image *entity.ActivityImage) (*entity.ActivityImage, error) {
	row, err := r.q.AddActivityImage(ctx, db.AddActivityImageParams{
		ActivityID: image.ActivityID,
		ImagePath:  image.ImagePath,
		ImageAlt:   ptrToPgText(image.ImageAlt),
	})
	if err != nil {
		return nil, fmt.Errorf("insert activity image: %w", err)
	}
	return toEntityActivityImage(row), nil
}

// ListByActivityID implements [usecase.ActivityImageRepository].
func (r *activityImageRepository) ListByActivityID(ctx context.Context, activityID uuid.UUID) ([]*entity.ActivityImage, error) {
	rows, err := r.q.ListActivityImagesByActivityID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("list activity images: %w", err)
	}
	return toEntityActivityImages(rows), nil
}

// Delete implements [usecase.ActivityImageRepository].
func (r *activityImageRepository) Delete(ctx context.Context, activityID uuid.UUID, imageID uuid.UUID) error {
	if err := r.q.DeleteActivityImage(ctx, db.DeleteActivityImageParams{
		ID:         imageID,
		ActivityID: activityID,
	}); err != nil {
		return fmt.Errorf("delete activity image: %w", err)
	}
	return nil
}
