package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
)

func toEntitySession(row db.UserSession) *entity.UserSession {
	return &entity.UserSession{
		ID:        row.ID,
		UserID:    row.UserID,
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt.Time,
		IPAddress: pgTextToPtr(row.IpAddress),
		UserAgent: pgTextToPtr(row.UserAgent),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
