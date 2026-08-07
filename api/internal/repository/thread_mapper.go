package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
)

func toEntityThread(row db.Thread) *entity.Thread {
	return &entity.Thread{
		ID:                         row.ID,
		UserID:                     row.UserID,
		Name:                       row.Name,
		Description:                pgTextToPtr(row.Description),
		HasCommitment:              row.HasCommitment,
		ColorHex:                   pgTextToPtr(row.ColorHex),
		ConfirmationTimeoutMinutes: pgInt4ToPtr(row.ConfirmationTimeoutMinutes),
		CreatedAt:                  row.CreatedAt.Time,
		UpdatedAt:                  row.UpdatedAt.Time,
		DeletedAt:                  pgTimestamptzToPtr(row.DeletedAt),
	}
}

func toCreateThreadParams(a *entity.Thread) db.CreateThreadParams {
	return db.CreateThreadParams{
		UserID:                     a.UserID,
		Name:                       a.Name,
		Description:                ptrToPgText(a.Description),
		HasCommitment:              a.HasCommitment,
		ColorHex:                   ptrToPgText(a.ColorHex),
		ConfirmationTimeoutMinutes: ptrToPgInt4(a.ConfirmationTimeoutMinutes),
	}
}

// toUpdateThreadParams deliberately has no HasCommitment field to
// set — UpdateThreadParams doesn't carry one either, since that flag
// has its own dedicated query (SetThreadHasCommitment).
func toUpdateThreadParams(a *entity.Thread) db.UpdateThreadParams {
	return db.UpdateThreadParams{
		ID:                         a.ID,
		UserID:                     a.UserID,
		Name:                       a.Name,
		Description:                ptrToPgText(a.Description),
		ColorHex:                   ptrToPgText(a.ColorHex),
		ConfirmationTimeoutMinutes: ptrToPgInt4(a.ConfirmationTimeoutMinutes),
	}
}
