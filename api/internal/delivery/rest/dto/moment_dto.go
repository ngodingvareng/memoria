package dto

import (
	"time"

	"github.com/ngodingvareng/memoria/internal/entity"
)

type CreateMomentRequest struct {
	ThreadID *string `json:"thread_id,omitempty" validate:"omitempty,uuid"`
	// Origin defaults to "manual" server-side when omitted. "commitment"
	// is deliberately not accepted here — Commitment-generated Moments
	// never go through this path (FEATURES.md).
	Origin *string `json:"origin,omitempty" validate:"omitempty,oneof=manual import"`

	OccurredAt string `json:"occurred_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00" example:"2026-08-04T14:30:00+09:00"`
	// The local UTC offset in effect where it happened, in minutes.
	// Omitted means UTC (0) — see the moments.occurred_utc_offset_minutes
	// column comment for the ±1020 (±17h) bound.
	OccurredUTCOffsetMinutes int16 `json:"occurred_utc_offset_minutes" validate:"gte=-1020,lte=1020" example:"540"`

	// RecordedAt is for offline sync only: the true local time the user
	// actually wrote the Moment down. Omitted means "now" (ordinary
	// online capture) — see FEATURES.md's Time table.
	RecordedAt *string `json:"recorded_at,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`

	Note      *string  `json:"note,omitempty"       validate:"omitempty,max=10000"`
	ColorHex  *string  `json:"color_hex,omitempty"  validate:"omitempty,hexcolor"`
	PlaceName *string  `json:"place_name,omitempty" validate:"omitempty,max=255"`
	Latitude  *float64 `json:"latitude,omitempty"   validate:"required_with=Longitude,omitempty,gte=-90,lte=90"`
	Longitude *float64 `json:"longitude,omitempty"  validate:"required_with=Latitude,omitempty,gte=-180,lte=180"`

	// ClientID lets an offline-captured Moment be synced twice (e.g. a
	// flaky retry) without creating a duplicate.
	ClientID *string `json:"client_id,omitempty" validate:"omitempty,uuid"`
}

type UpdateMomentRequest struct {
	ThreadID *string `json:"thread_id,omitempty" validate:"omitempty,uuid"`

	OccurredAt               string `json:"occurred_at"                 validate:"required,datetime=2006-01-02T15:04:05Z07:00" example:"2026-08-04T14:30:00+09:00"`
	OccurredUTCOffsetMinutes int16  `json:"occurred_utc_offset_minutes" validate:"gte=-1020,lte=1020"                          example:"540"`

	Note      *string  `json:"note,omitempty"       validate:"omitempty,max=10000"`
	ColorHex  *string  `json:"color_hex,omitempty"  validate:"omitempty,hexcolor"`
	PlaceName *string  `json:"place_name,omitempty" validate:"omitempty,max=255"`
	Latitude  *float64 `json:"latitude,omitempty"   validate:"required_with=Longitude,omitempty,gte=-90,lte=90"`
	Longitude *float64 `json:"longitude,omitempty"  validate:"required_with=Latitude,omitempty,gte=-180,lte=180"`
}

type ListMomentsQuery struct {
	Page     int32 `query:"page"      validate:"omitempty,gt=0"`
	PageSize int32 `query:"page_size" validate:"omitempty,gt=0,lte=100"`
}

type SearchMomentsQuery struct {
	Query    string `query:"q"         validate:"required,min=1,max=255"`
	Page     int32  `query:"page"      validate:"omitempty,gt=0"`
	PageSize int32  `query:"page_size" validate:"omitempty,gt=0,lte=100"`
}

type MomentResponse struct {
	ID       string  `json:"id"                  example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	UserID   string  `json:"user_id"             example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	ThreadID *string `json:"thread_id,omitempty" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Origin   string  `json:"origin"              example:"manual"`

	// OccurredAt is reconstructed from occurred_local +
	// occurred_utc_offset_minutes, so it round-trips exactly the wall
	// time and offset the Moment was captured with.
	OccurredAt               string `json:"occurred_at"                 example:"2026-08-04T14:30:00+09:00"`
	OccurredUTCOffsetMinutes int16  `json:"occurred_utc_offset_minutes" example:"540"`
	RecordedAt               string `json:"recorded_at"                 example:"2026-08-04T14:35:00Z"`
	SettlingTimeSeconds      int64  `json:"settling_time_seconds"       example:"300"`

	Note      *string  `json:"note,omitempty"`
	ColorHex  *string  `json:"color_hex,omitempty"`
	PlaceName *string  `json:"place_name,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	ClientID  *string  `json:"client_id,omitempty"`

	CreatedAt string `json:"created_at" example:"2026-08-04T14:35:00Z"`
	UpdatedAt string `json:"updated_at" example:"2026-08-04T14:35:00Z"`
}

type ListMomentsResponse struct {
	Moments    []MomentResponse   `json:"moments"`
	Pagination PaginationResponse `json:"pagination"`
}

func NewMomentResponse(e *entity.Moment) MomentResponse {
	var threadID *string
	if e.ThreadID != nil {
		id := e.ThreadID.String()
		threadID = &id
	}

	var clientID *string
	if e.ClientID != nil {
		id := e.ClientID.String()
		clientID = &id
	}

	loc := time.FixedZone("", int(e.OccurredUTCOffsetMinutes)*60)
	occurredAt := time.Date(
		e.OccurredLocal.Year(), e.OccurredLocal.Month(), e.OccurredLocal.Day(),
		e.OccurredLocal.Hour(), e.OccurredLocal.Minute(), e.OccurredLocal.Second(), e.OccurredLocal.Nanosecond(),
		loc,
	)

	return MomentResponse{
		ID:                       e.ID.String(),
		UserID:                   e.UserID.String(),
		ThreadID:                 threadID,
		Origin:                   string(e.Origin),
		OccurredAt:               occurredAt.Format(time.RFC3339),
		OccurredUTCOffsetMinutes: e.OccurredUTCOffsetMinutes,
		RecordedAt:               e.RecordedAt.Format(time.RFC3339),
		SettlingTimeSeconds:      int64(e.SettlingTime.Seconds()),
		Note:                     e.Note,
		ColorHex:                 e.ColorHex,
		PlaceName:                e.PlaceName,
		Latitude:                 e.Latitude,
		Longitude:                e.Longitude,
		ClientID:                 clientID,
		CreatedAt:                e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                e.UpdatedAt.Format(time.RFC3339),
	}
}
