package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
)

// toEntityThread deliberately leaves out row.SearchDocument — it's a
// tsvector search index, not a domain value, and stays a repository
// concern (see the Thread entity's own header comment).
func toEntityThread(row db.Thread) *entity.Thread {
	return &entity.Thread{
		ID:          row.ID,
		UserID:      row.UserID,
		CircleID:    pgUUIDToPtr(row.CircleID),
		Name:        row.Name,
		Description: pgTextToPtr(row.Description),
		ColorHex:    pgTextToPtr(row.ColorHex),
		SortOrder:   row.SortOrder,
		ArchivedAt:  pgTimestamptzToPtr(row.ArchivedAt),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		DeletedAt:   pgTimestamptzToPtr(row.DeletedAt),
	}
}

func toCreateThreadParams(a *entity.Thread) db.CreateThreadParams {
	return db.CreateThreadParams{
		UserID:      a.UserID,
		CircleID:    ptrToPgUUID(a.CircleID),
		Name:        a.Name,
		Description: ptrToPgText(a.Description),
		ColorHex:    ptrToPgText(a.ColorHex),
	}
}

func toUpdateThreadParams(a *entity.Thread) db.UpdateThreadParams {
	return db.UpdateThreadParams{
		ID:          a.ID,
		UserID:      a.UserID,
		Name:        a.Name,
		Description: ptrToPgText(a.Description),
		ColorHex:    ptrToPgText(a.ColorHex),
	}
}
