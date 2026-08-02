package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
)

func toEntityRefreshToken(row db.RefreshToken) *entity.RefreshToken {
	return &entity.RefreshToken{
		ID:           row.ID,
		UserID:       row.UserID,
		FamilyID:     row.FamilyID,
		TokenHash:    row.TokenHash,
		ExpiresAt:    row.ExpiresAt.Time,
		RevokedAt:    pgTimestamptzToPtr(row.RevokedAt),
		ReplacedByID: pgUUIDToPtr(row.ReplacedByID),
		IPAddress:    pgTextToPtr(row.IpAddress),
		UserAgent:    pgTextToPtr(row.UserAgent),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}
