package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

// AlbumRepository is the minimal capability AlbumUsecase needs: the two
// flattened, denormalized image-row queries that back the Personal and
// Circle Album feeds (FEATURES.md, Looking Back). Implemented by
// momentImageRepository — both queries are moment_images reads, just
// scoped by a wider visibility set than
// MomentImageRepository.ListByMomentID.
type AlbumRepository interface {
	// ListPersonalImages is every photo attachment from every Moment
	// userID owns, newest occurrence first.
	ListPersonalImages(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.AlbumImage, error)
	// ListCircleImages is every photo attachment from every Moment
	// visible in circleID — its own collaborative Threads' Moments plus
	// personal Moments shared into it — newest occurrence first.
	// Membership must be checked by the caller.
	ListCircleImages(ctx context.Context, circleID uuid.UUID, limit, offset int32) ([]*entity.AlbumImage, error)
}

// AlbumImageWithURL pairs the stored record with a freshly presigned,
// time-limited download URL — the URL itself is never persisted, it's
// generated on read. Mirrors MomentImageWithURL/ThreadImageWithURL.
type AlbumImageWithURL struct {
	*entity.AlbumImage

	URL string
}

type ListAlbumInput struct {
	UserID   uuid.UUID
	Page     int32
	PageSize int32
}

type ListCircleAlbumInput struct {
	UserID   uuid.UUID
	CircleID uuid.UUID
	Page     int32
	PageSize int32
}

// AlbumListResult mirrors MomentListResult: a page of Album rows, no
// Total — no COUNT query backs this, same pagination shape as
// ListMoments/ListCircleMoments/ListThreadMoments.
type AlbumListResult struct {
	Images   []AlbumImageWithURL
	Page     int32
	PageSize int32
}

type AlbumUsecase interface {
	// ListAlbum is the Personal Album: every photo attachment from
	// every Moment the caller owns.
	ListAlbum(ctx context.Context, input ListAlbumInput) (*AlbumListResult, error)
	// ListCircleAlbum is circleID's Album feed. Returns errs.ErrNotFound
	// if userID isn't an active member.
	ListCircleAlbum(ctx context.Context, input ListCircleAlbumInput) (*AlbumListResult, error)
}

type albumUsecase struct {
	repo    AlbumRepository
	storage Storage
	circles CircleAccessChecker
}

func NewAlbumUsecase(repo AlbumRepository, storage Storage, circles CircleAccessChecker) AlbumUsecase {
	return &albumUsecase{repo: repo, storage: storage, circles: circles}
}

func (u *albumUsecase) presignAll(ctx context.Context, images []*entity.AlbumImage) ([]AlbumImageWithURL, error) {
	out := make([]AlbumImageWithURL, 0, len(images))
	for _, img := range images {
		url, err := u.storage.PresignGet(ctx, img.ImagePath, presignedURLTTL)
		if err != nil {
			return nil, fmt.Errorf("presigning album image %s: %w", img.ID, err)
		}
		out = append(out, AlbumImageWithURL{AlbumImage: img, URL: url})
	}
	return out, nil
}

// ListAlbum implements [AlbumUsecase].
func (u *albumUsecase) ListAlbum(ctx context.Context, input ListAlbumInput) (*AlbumListResult, error) {
	page, pageSize := normalizedPage(input.Page, input.PageSize)

	images, err := u.repo.ListPersonalImages(ctx, input.UserID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("listing album images: %w", err)
	}
	withURLs, err := u.presignAll(ctx, images)
	if err != nil {
		return nil, err
	}
	return &AlbumListResult{Images: withURLs, Page: page, PageSize: pageSize}, nil
}

// ListCircleAlbum implements [AlbumUsecase]. Membership is checked
// here, before the repository is ever asked for its images — same
// pattern as MomentUsecase.ListCircleMoments.
func (u *albumUsecase) ListCircleAlbum(ctx context.Context, input ListCircleAlbumInput) (*AlbumListResult, error) {
	if _, err := u.circles.GetActiveMember(ctx, input.CircleID, input.UserID); err != nil {
		return nil, fmt.Errorf("checking circle membership: %w", err)
	}

	page, pageSize := normalizedPage(input.Page, input.PageSize)

	images, err := u.repo.ListCircleImages(ctx, input.CircleID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("listing circle album images: %w", err)
	}
	withURLs, err := u.presignAll(ctx, images)
	if err != nil {
		return nil, err
	}
	return &AlbumListResult{Images: withURLs, Page: page, PageSize: pageSize}, nil
}
