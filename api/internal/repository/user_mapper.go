package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
)

func toEntityUser(row db.User) *entity.User {
	return &entity.User{
		ID:            row.ID,
		Name:          row.Name,
		Email:         row.Email,
		EmailVerified: row.EmailVerified,
		ImagePath:     pgTextToPtr(row.ImagePath),
		Timezone:      row.Timezone,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		DeletedAt:     pgTimestamptzToPtr(row.DeletedAt),
	}
}
