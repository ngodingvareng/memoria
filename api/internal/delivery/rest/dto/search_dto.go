package dto

import (
	"time"

	"github.com/ngodingvareng/memoria/internal/entity"
)

type SearchSuggestionsQuery struct {
	Query string `query:"q" validate:"required,min=1,max=255"`
}

// MomentSuggestionResponse is a minimal suggestion row, not the full
// Moment payload — a live-typing popover only needs enough to identify
// and preview the match.
type MomentSuggestionResponse struct {
	ID         string  `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Note       *string `json:"note,omitempty"`
	PlaceName  *string `json:"place_name,omitempty"`
	OccurredAt string  `json:"occurred_at" example:"2026-08-04T14:30:00+09:00"`
}

type ThreadSuggestionResponse struct {
	ID       string  `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Name     string  `json:"name" example:"Morning workout"`
	ColorHex *string `json:"color_hex,omitempty"`
}

type SearchSuggestionsResponse struct {
	Moments []MomentSuggestionResponse `json:"moments"`
	Threads []ThreadSuggestionResponse `json:"threads"`
}

func NewMomentSuggestionResponse(e *entity.Moment) MomentSuggestionResponse {
	loc := time.FixedZone("", int(e.OccurredUTCOffsetMinutes)*60)
	occurredAt := time.Date(
		e.OccurredLocal.Year(), e.OccurredLocal.Month(), e.OccurredLocal.Day(),
		e.OccurredLocal.Hour(), e.OccurredLocal.Minute(), e.OccurredLocal.Second(), e.OccurredLocal.Nanosecond(),
		loc,
	)
	return MomentSuggestionResponse{
		ID:         e.ID.String(),
		Note:       e.Note,
		PlaceName:  e.PlaceName,
		OccurredAt: occurredAt.Format(time.RFC3339),
	}
}

func NewThreadSuggestionResponse(e *entity.Thread) ThreadSuggestionResponse {
	return ThreadSuggestionResponse{ID: e.ID.String(), Name: e.Name, ColorHex: e.ColorHex}
}
