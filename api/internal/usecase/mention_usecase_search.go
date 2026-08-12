package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// mentionSearchOverfetchLimit is how many discoverable candidates
// SearchMentionableUsers pulls before narrowing by mention_policy/
// blocking; mentionSearchResultLimit is what actually reaches the
// caller once that filtering is applied.
const (
	mentionSearchOverfetchLimit = 20
	mentionSearchResultLimit    = 8
)

// SearchMentionableUsers implements [MentionUsecase]. Applies the same
// per-candidate checks CreateMention applies to a single resolved
// username — blocking (either direction) and mention_policy, including
// resolving audience_policy = "known" via IsKnownTo — just stopping
// once mentionSearchResultLimit matches are found instead of failing on
// the first disallowed one.
func (u *mentionUsecase) SearchMentionableUsers(
	ctx context.Context,
	requestingUserID uuid.UUID,
	query string,
) ([]*entity.User, error) {
	if strings.TrimSpace(query) == "" {
		return []*entity.User{}, nil
	}

	candidates, err := u.searcher.SearchByUsernamePrefix(ctx, requestingUserID, query, mentionSearchOverfetchLimit)
	if err != nil {
		return nil, fmt.Errorf("searching users by username: %w", err)
	}

	results := make([]*entity.User, 0, mentionSearchResultLimit)
	for _, candidate := range candidates {
		if len(results) >= mentionSearchResultLimit {
			break
		}

		blocked, err := u.blocks.IsBlockedEitherDirection(ctx, requestingUserID, candidate.ID)
		if err != nil {
			return nil, fmt.Errorf("checking block status: %w", err)
		}
		if blocked {
			continue
		}

		switch candidate.MentionPolicy {
		case enum.AudiencePolicyAnyone:
			results = append(results, candidate)
		case enum.AudiencePolicyKnown:
			known, err := u.knowns.IsKnownTo(ctx, candidate.ID, requestingUserID)
			if err != nil {
				return nil, fmt.Errorf("checking known status: %w", err)
			}
			if known {
				results = append(results, candidate)
			}
		default: // enum.AudiencePolicyNobody
		}
	}

	return results, nil
}
