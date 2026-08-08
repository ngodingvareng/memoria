package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
)

func toEntityUser(row db.User) *entity.User {
	return &entity.User{
		ID:                     row.ID,
		Name:                   row.Name,
		Username:               pgTextToPtr(row.Username),
		Email:                  row.Email,
		EmailVerified:          row.EmailVerified,
		ImagePath:              pgTextToPtr(row.ImagePath),
		Bio:                    pgTextToPtr(row.Bio),
		Timezone:               row.Timezone,
		MentionPolicy:          enum.AudiencePolicy(row.MentionPolicy),
		CircleInvitePolicy:     enum.AudiencePolicy(row.CircleInvitePolicy),
		DiscoverableByUsername: row.DiscoverableByUsername,
		StripPhotoMetadata:     row.StripPhotoMetadata,
		CreatedAt:              row.CreatedAt.Time,
		UpdatedAt:              row.UpdatedAt.Time,
		DeletedAt:              pgTimestamptzToPtr(row.DeletedAt),
	}
}
