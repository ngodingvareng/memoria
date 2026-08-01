package repository

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

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

// pgUUIDToPtr converts a nullable pgtype.UUID into *uuid.UUID. Needed for
// ReplacedByID, which is NULL until a refresh token has been rotated out.
func pgUUIDToPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}
