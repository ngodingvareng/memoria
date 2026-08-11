package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

// ListMentions implements [MentionUsecase].
func (u *mentionUsecase) ListMentions(ctx context.Context, momentID, ownerUserID uuid.UUID) ([]*entity.MomentMention, error) {
	if _, err := u.moments.GetByID(ctx, momentID, ownerUserID); err != nil {
		return nil, fmt.Errorf("checking moment ownership: %w", err)
	}
	mentions, err := u.repo.ListByMomentID(ctx, momentID)
	if err != nil {
		return nil, fmt.Errorf("listing mentions: %w", err)
	}
	return mentions, nil
}

// ListMentionedMoments implements [MentionUsecase]: "Mentioned Moments"
// (FEATURES.md, Looking Back).
func (u *mentionUsecase) ListMentionedMoments(ctx context.Context, input ListMentionedMomentsInput) (*MomentListResult, error) {
	page, pageSize := normalizedPage(input.Page, input.PageSize)
	moments, err := u.repo.ListMentionedMoments(ctx, input.UserID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("listing mentioned moments: %w", err)
	}
	return &MomentListResult{Moments: moments, Page: page, PageSize: pageSize}, nil
}

// ListSharedCircles implements [MentionUsecase].
func (u *mentionUsecase) ListSharedCircles(ctx context.Context, momentID, ownerUserID uuid.UUID) ([]uuid.UUID, error) {
	if _, err := u.moments.GetByID(ctx, momentID, ownerUserID); err != nil {
		return nil, fmt.Errorf("checking moment ownership: %w", err)
	}
	ids, err := u.repo.ListSharedCircleIDs(ctx, momentID)
	if err != nil {
		return nil, fmt.Errorf("listing shared circles: %w", err)
	}
	return ids, nil
}
