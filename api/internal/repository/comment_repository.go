package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.CommentRepository = (*commentRepository)(nil)

type commentRepository struct {
	q *db.Queries
}

func NewCommentRepository(pool *pgxpool.Pool) *commentRepository {
	return &commentRepository{q: db.New(pool)}
}

// Create implements [usecase.CommentRepository].
func (r *commentRepository) Create(ctx context.Context, momentID uuid.UUID, userID uuid.UUID, circleID *uuid.UUID, body string) (*entity.Comment, error) {
	row, err := r.q.CreateComment(ctx, db.CreateCommentParams{
		MomentID: momentID, UserID: ptrToPgUUID(&userID), CircleID: ptrToPgUUID(circleID), Body: body,
	})
	if err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}
	return toEntityComment(row), nil
}

// GetByID implements [usecase.CommentRepository].
func (r *commentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Comment, error) {
	row, err := r.q.GetCommentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get comment by id: %w", err)
	}
	return toEntityComment(row), nil
}

// ListForMoment implements [usecase.CommentRepository].
func (r *commentRepository) ListForMoment(ctx context.Context, momentID, viewerID uuid.UUID, mentionAllowed bool, visibleCircleIDs []uuid.UUID) ([]*entity.Comment, error) {
	if visibleCircleIDs == nil {
		visibleCircleIDs = []uuid.UUID{}
	}
	rows, err := r.q.ListCommentsForMoment(ctx, db.ListCommentsForMomentParams{
		MomentID: momentID, MentionAllowed: mentionAllowed, VisibleCircleIds: visibleCircleIDs, ViewerID: viewerID,
	})
	if err != nil {
		return nil, fmt.Errorf("list comments for moment: %w", err)
	}
	return toEntityComments(rows), nil
}

// Update implements [usecase.CommentRepository].
func (r *commentRepository) Update(ctx context.Context, id, userID uuid.UUID, body string) (*entity.Comment, error) {
	row, err := r.q.UpdateComment(ctx, db.UpdateCommentParams{Body: body, ID: id, UserID: ptrToPgUUID(&userID)})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("update comment: %w", err)
	}
	return toEntityComment(row), nil
}

// SoftDelete implements [usecase.CommentRepository]. :exec — a WHERE
// clause matching zero rows (wrong owner, wrong id, already deleted) is
// a silent no-op here, same pattern as every other soft-delete in this
// codebase.
func (r *commentRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	if err := r.q.SoftDeleteComment(ctx, db.SoftDeleteCommentParams{ID: id, UserID: ptrToPgUUID(&userID)}); err != nil {
		return fmt.Errorf("soft delete comment: %w", err)
	}
	return nil
}

// AnonymizeByUserID implements [usecase.CommentRepository].
func (r *commentRepository) AnonymizeByUserID(ctx context.Context, userID uuid.UUID) error {
	if err := r.q.AnonymizeCommentsByUserID(ctx, ptrToPgUUID(&userID)); err != nil {
		return fmt.Errorf("anonymize comments by user id: %w", err)
	}
	return nil
}
