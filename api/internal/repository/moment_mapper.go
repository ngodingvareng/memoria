package repository

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// toEntityMoment deliberately leaves out row.SearchDocument — it's a
// tsvector search index, not a domain value, and stays a repository
// concern (see the Moment entity's own header comment).
func toEntityMoment(row db.Moment) *entity.Moment {
	return &entity.Moment{
		ID:                       row.ID,
		UserID:                   row.UserID,
		ThreadID:                 pgUUIDToPtr(row.ThreadID),
		Origin:                   enum.MomentOrigin(row.Origin),
		OccurredAt:               row.OccurredAt.Time,
		OccurredLocal:            row.OccurredLocal.Time,
		OccurredUTCOffsetMinutes: row.OccurredUtcOffsetMinutes,
		OccurredOn:               row.OccurredOn.Time,
		RecordedAt:               row.RecordedAt.Time,
		SettlingTime:             pgIntervalToDuration(row.SettlingTime),
		Note:                     pgTextToPtr(row.Note),
		ColorHex:                 pgTextToPtr(row.ColorHex),
		PlaceName:                pgTextToPtr(row.PlaceName),
		Latitude:                 pgNumericToPtr(row.Latitude),
		Longitude:                pgNumericToPtr(row.Longitude),
		ClientID:                 pgUUIDToPtr(row.ClientID),
		LastViewedAt:             pgTimestamptzToPtr(row.LastViewedAt),
		CreatedAt:                row.CreatedAt.Time,
		UpdatedAt:                row.UpdatedAt.Time,
		DeletedAt:                pgTimestamptzToPtr(row.DeletedAt),
	}
}

func toEntityMoments(rows []db.Moment) []*entity.Moment {
	out := make([]*entity.Moment, len(rows))
	for i, row := range rows {
		out[i] = toEntityMoment(row)
	}
	return out
}

// toCreateMomentParams. A zero-value m.RecordedAt means "let the DB
// default to NOW()" (mirrors CreateMoment's own COALESCE) — a real
// recorded_at can never legitimately be the Go zero time.
func toCreateMomentParams(m *entity.Moment) db.CreateMomentParams {
	var recordedAt any
	if !m.RecordedAt.IsZero() {
		recordedAt = timeToPgTimestamptz(m.RecordedAt)
	}

	return db.CreateMomentParams{
		UserID:                   m.UserID,
		ThreadID:                 ptrToPgUUID(m.ThreadID),
		Origin:                   db.MomentOrigin(m.Origin),
		OccurredAt:               timeToPgTimestamptz(m.OccurredAt),
		OccurredLocal:            pgtype.Timestamp{Time: m.OccurredLocal, Valid: true},
		OccurredUtcOffsetMinutes: m.OccurredUTCOffsetMinutes,
		RecordedAt:               recordedAt,
		Note:                     ptrToPgText(m.Note),
		ColorHex:                 ptrToPgText(m.ColorHex),
		PlaceName:                ptrToPgText(m.PlaceName),
		Latitude:                 ptrToPgNumeric(m.Latitude),
		Longitude:                ptrToPgNumeric(m.Longitude),
		ClientID:                 ptrToPgUUID(m.ClientID),
	}
}

func toUpdateMomentParams(m *entity.Moment) db.UpdateMomentParams {
	return db.UpdateMomentParams{
		ID:                       m.ID,
		UserID:                   m.UserID,
		ThreadID:                 ptrToPgUUID(m.ThreadID),
		OccurredAt:               timeToPgTimestamptz(m.OccurredAt),
		OccurredLocal:            pgtype.Timestamp{Time: m.OccurredLocal, Valid: true},
		OccurredUtcOffsetMinutes: m.OccurredUTCOffsetMinutes,
		Note:                     ptrToPgText(m.Note),
		ColorHex:                 ptrToPgText(m.ColorHex),
		PlaceName:                ptrToPgText(m.PlaceName),
		Latitude:                 ptrToPgNumeric(m.Latitude),
		Longitude:                ptrToPgNumeric(m.Longitude),
	}
}
