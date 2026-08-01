package repository

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// This file holds pgtype <-> plain-Go-pointer conversions shared by
// every entity mapper in this package (activity_mapper.go,
// user_mapper.go, etc. as they're added). Keeping pgtype awareness
// confined to internal/repository is deliberate — entity and usecase
// never see pgtype directly, only plain Go types (string, *string,
// int32, time.Time, ...).

func pgTextToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return new(t.String)
}

func ptrToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func pgInt4ToPtr(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	return new(i.Int32)
}

func ptrToPgInt4(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

func pgTimestamptzToPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return new(t.Time)
}

func ptrToPgTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func pgBoolToPtr(b pgtype.Bool) *bool {
	if !b.Valid {
		return nil
	}
	return new(b.Bool)
}

func ptrToPgBool(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}
