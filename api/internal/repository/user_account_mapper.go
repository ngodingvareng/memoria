package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
)

func toEntityUserAccount(row db.UserAccount) *entity.UserAccount {
	return &entity.UserAccount{
		ID:           row.ID,
		UserID:       row.UserID,
		AccountID:    row.AccountID,
		ProviderID:   enum.AuthProvider(row.ProviderID),
		PasswordHash: pgTextToPtr(row.PasswordHash),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}
