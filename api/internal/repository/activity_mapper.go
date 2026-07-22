package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
)

func toEntityActivity(row db.Activity) *entity.Activity {
	return &entity.Activity{
		ID:                         row.ID,
		UserID:                     row.UserID,
		Name:                       row.Name,
		Description:                pgTextToPtr(row.Description),
		IsFixedSchedule:            row.IsFixedSchedule,
		ColorPalette:               pgTextToPtr(row.ColorPalette),
		ConfirmationTimeoutMinutes: pgInt4ToPtr(row.ConfirmationTimeoutMinutes),
		CreatedAt:                  row.CreatedAt.Time,
		UpdatedAt:                  row.UpdatedAt.Time,
		DeletedAt:                  pgTimestamptzToPtr(row.DeletedAt),
	}
}

func toCreateActivityParams(a *entity.Activity) db.CreateActivityParams {
	return db.CreateActivityParams{
		UserID:                     a.UserID,
		Name:                       a.Name,
		Description:                ptrToPgText(a.Description),
		IsFixedSchedule:            a.IsFixedSchedule,
		ColorPalette:               ptrToPgText(a.ColorPalette),
		ConfirmationTimeoutMinutes: ptrToPgInt4(a.ConfirmationTimeoutMinutes),
	}
}
