package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

type User struct {
	ID   uuid.UUID
	Name string
	// Username is nil until the post-register onboarding step claims one
	// (see UserUsecase.SetUsername).
	Username      *string
	Email         string
	EmailVerified bool
	ImagePath     *string
	Bio           *string
	Timezone      string

	// Social Interaction Controls (FEATURES.md, Privacy & Control).
	MentionPolicy          enum.AudiencePolicy
	CircleInvitePolicy     enum.AudiencePolicy
	DiscoverableByUsername bool

	// Data Controls.
	StripPhotoMetadata bool

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
