package dto

import (
	"time"

	"github.com/ngodingvareng/memoria/internal/usecase"
)

type ListAlbumQuery struct {
	Page     int32 `query:"page"      validate:"omitempty,gt=0"`
	PageSize int32 `query:"page_size" validate:"omitempty,gt=0,lte=100"`
}

// AlbumSharedByResponse is who shared an image's Moment into a Circle
// — only present on a Circle Album row where IsShared is true, backing
// the "from @username" attribution label (FEATURES.md, Looking Back).
// Username is nullable because users.username itself is (a sharer who
// hasn't claimed a handle yet, or who has since deleted their account —
// see NewAlbumImageResponse).
type AlbumSharedByResponse struct {
	ID       string  `json:"id"                 example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Username *string `json:"username,omitempty" example:"budisantoso"`
}

// AlbumImageResponse is one flattened photo row — moment_id and
// occurred_at are denormalized directly onto it (no nested Moment
// payload) so the client can date-bucket the Album into its timeline
// with no extra per-Moment fetch.
type AlbumImageResponse struct {
	ID         string  `json:"id"                  example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	MomentID   string  `json:"moment_id"           example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	URL        string  `json:"url"                 example:"https://storage.example.com/moments/.../abc123.jpg?X-Amz-..."`
	ImageAlt   *string `json:"image_alt,omitempty"`
	Width      *int32  `json:"width,omitempty"`
	Height     *int32  `json:"height,omitempty"`
	OccurredAt string  `json:"occurred_at"         example:"2026-08-04T14:30:00+09:00"`

	// ThreadName is omitted when the image's Moment has no parent
	// Thread.
	ThreadName *string `json:"thread_name,omitempty" example:"Morning Run"`

	// IsShared/SharedBy are always false/omitted on GET /album. On
	// GET /circles/{id}/album, IsShared is true and SharedBy is set iff
	// this image's Moment reached the Circle via the mention/share
	// flow rather than one of the Circle's own collaborative Threads.
	IsShared bool                   `json:"is_shared"`
	SharedBy *AlbumSharedByResponse `json:"shared_by,omitempty"`
}

type ListAlbumResponse struct {
	Images     []AlbumImageResponse `json:"images"`
	Pagination PaginationResponse   `json:"pagination"`
}

func NewAlbumImageResponse(img usecase.AlbumImageWithURL) AlbumImageResponse {
	loc := time.FixedZone("", int(img.OccurredUTCOffsetMinutes)*60)
	occurredAt := time.Date(
		img.OccurredLocal.Year(),
		img.OccurredLocal.Month(),
		img.OccurredLocal.Day(),
		img.OccurredLocal.Hour(),
		img.OccurredLocal.Minute(),
		img.OccurredLocal.Second(),
		img.OccurredLocal.Nanosecond(),
		loc,
	)

	resp := AlbumImageResponse{
		ID:         img.ID.String(),
		MomentID:   img.MomentID.String(),
		URL:        img.URL,
		ImageAlt:   img.ImageAlt,
		Width:      img.Width,
		Height:     img.Height,
		OccurredAt: occurredAt.Format(time.RFC3339),
		IsShared:   img.IsShared,
		ThreadName: img.ThreadName,
	}
	// SharedByUserID can be nil even when IsShared is true: shared_by_user_id
	// is ON DELETE SET NULL, so a sharer who deleted their account leaves
	// the attribution flag set but no user to point at — omit the label
	// rather than render a broken reference.
	if img.IsShared && img.SharedByUserID != nil {
		resp.SharedBy = &AlbumSharedByResponse{
			ID:       img.SharedByUserID.String(),
			Username: img.SharedByUsername,
		}
	}
	return resp
}

func NewListAlbumResponse(result *usecase.AlbumListResult) ListAlbumResponse {
	images := make([]AlbumImageResponse, len(result.Images))
	for i, img := range result.Images {
		images[i] = NewAlbumImageResponse(img)
	}
	return ListAlbumResponse{
		Images: images,
		Pagination: PaginationResponse{
			Page:     result.Page,
			PageSize: result.PageSize,
		},
	}
}
