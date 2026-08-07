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

var _ usecase.ThreadImageRepository = (*threadImageRepository)(nil)

type threadImageRepository struct {
	q *db.Queries
}

func NewThreadImageRepository(pool *pgxpool.Pool) *threadImageRepository {
	return &threadImageRepository{q: db.New(pool)}
}

// Create implements [usecase.ThreadImageRepository].
func (r *threadImageRepository) Create(ctx context.Context, image *entity.ThreadImage) (*entity.ThreadImage, error) {
	row, err := r.q.AddThreadImage(ctx, db.AddThreadImageParams{
		ThreadID:  image.ThreadID,
		ImagePath: image.ImagePath,
		ImageAlt:  ptrToPgText(image.ImageAlt),
	})
	if err != nil {
		return nil, fmt.Errorf("insert thread image: %w", err)
	}
	return toEntityThreadImage(row), nil
}

// ListByThreadID implements [usecase.ThreadImageRepository].
func (r *threadImageRepository) ListByThreadID(ctx context.Context, threadID uuid.UUID) ([]*entity.ThreadImage, error) {
	rows, err := r.q.ListThreadImagesByThreadID(ctx, threadID)
	if err != nil {
		return nil, fmt.Errorf("list thread images: %w", err)
	}
	return toEntityThreadImages(rows), nil
}

// Delete implements [usecase.ThreadImageRepository].
func (r *threadImageRepository) Delete(ctx context.Context, threadID uuid.UUID, imageID uuid.UUID) error {
	if err := r.q.DeleteThreadImage(ctx, db.DeleteThreadImageParams{
		ID:       imageID,
		ThreadID: threadID,
	}); err != nil {
		return fmt.Errorf("delete thread image: %w", err)
	}
	return nil
}
