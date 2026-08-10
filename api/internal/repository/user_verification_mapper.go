package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
)

func toEntityUserVerification(row db.UserVerification) *entity.UserVerification {
	return &entity.UserVerification{
		ID:         row.ID,
		Identifier: row.Identifier,
		Value:      row.Value,
		ExpiresAt:  row.ExpiresAt.Time,
		ConsumedAt: pgTimestamptzToPtr(row.ConsumedAt),
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}
}
