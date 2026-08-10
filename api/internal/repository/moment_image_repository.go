package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.MomentImageRepository = (*momentImageRepository)(nil)
var _ usecase.AlbumRepository = (*momentImageRepository)(nil)

type momentImageRepository struct {
	q *db.Queries
}

func NewMomentImageRepository(pool *pgxpool.Pool) *momentImageRepository {
	return &momentImageRepository{q: db.New(pool)}
}

// Create implements [usecase.MomentImageRepository].
func (r *momentImageRepository) Create(ctx context.Context, image *entity.MomentImage) (*entity.MomentImage, error) {
	row, err := r.q.AddMomentImage(ctx, db.AddMomentImageParams{
		MomentID:         image.MomentID,
		ImagePath:        image.ImagePath,
		ImageAlt:         ptrToPgText(image.ImageAlt),
		ContentType:      ptrToPgText(image.ContentType),
		ByteSize:         ptrToPgInt8(image.ByteSize),
		Width:            ptrToPgInt4(image.Width),
		Height:           ptrToPgInt4(image.Height),
		MetadataStripped: image.MetadataStripped,
	})
	if err != nil {
		return nil, fmt.Errorf("insert moment image: %w", err)
	}
	return toEntityMomentImage(row), nil
}

// ListByMomentID implements [usecase.MomentImageRepository].
func (r *momentImageRepository) ListByMomentID(ctx context.Context, momentID uuid.UUID) ([]*entity.MomentImage, error) {
	rows, err := r.q.ListMomentImagesByMomentID(ctx, momentID)
	if err != nil {
		return nil, fmt.Errorf("list moment images: %w", err)
	}
	return toEntityMomentImages(rows), nil
}

// Delete implements [usecase.MomentImageRepository]. Returns the deleted
// image's path so the caller can remove the matching storage object; a
// mismatched id/momentID is a silent no-op (nil, nil), matching the
// ownership-check semantics used elsewhere in this package.
func (r *momentImageRepository) Delete(ctx context.Context, momentID, imageID uuid.UUID) (*entity.MomentImage, error) {
	row, err := r.q.DeleteMomentImage(ctx, db.DeleteMomentImageParams{ID: imageID, MomentID: momentID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("delete moment image: %w", err)
	}
	return toEntityMomentImage(row), nil
}

// ListPersonalImages implements [usecase.AlbumRepository].
func (r *momentImageRepository) ListPersonalImages(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.AlbumImage, error) {
	rows, err := r.q.ListAlbumImages(ctx, db.ListAlbumImagesParams{
		UserID:     userID,
		PageLimit:  limit,
		PageOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list album images: %w", err)
	}
	return toEntityAlbumImages(rows), nil
}

// ListCircleImages implements [usecase.AlbumRepository]. No membership
// filter at the query level — access is checked upstream by
// AlbumUsecase via CircleAccessChecker, same pattern as
// MomentUsecase.ListCircleMoments.
func (r *momentImageRepository) ListCircleImages(ctx context.Context, circleID uuid.UUID, limit, offset int32) ([]*entity.AlbumImage, error) {
	rows, err := r.q.ListCircleAlbumImages(ctx, db.ListCircleAlbumImagesParams{
		CircleID:   ptrToPgUUID(&circleID),
		PageLimit:  limit,
		PageOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list circle album images: %w", err)
	}
	return toEntityCircleAlbumImages(rows), nil
}
