package usecase

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

// suggestionLimit is the fixed top-N per entity returned by
// GET /search/suggestions — a live-typing popover, not a paginated page.
const suggestionLimit int32 = 5

// ThreadSuggestionSearcher is the minimal capability SearchUsecase needs
// from the Thread domain. Method name matches
// ThreadRepository.SearchSuggestions on purpose — same pattern as
// ThreadAccessChecker (thread_image_usecase.go).
type ThreadSuggestionSearcher interface {
	SearchSuggestions(
		ctx context.Context,
		userID uuid.UUID,
		query, prefixQuery string,
		limit int32,
	) ([]*entity.Thread, error)
}

// MomentSuggestionSearcher is the Moment-domain equivalent.
type MomentSuggestionSearcher interface {
	SearchSuggestions(
		ctx context.Context,
		userID uuid.UUID,
		query, prefixQuery string,
		limit int32,
	) ([]*entity.Moment, error)
}

type SearchSuggestionsInput struct {
	UserID uuid.UUID
	Query  string
}

type SearchSuggestionsResult struct {
	Moments []*entity.Moment
	Threads []*entity.Thread
}

type SearchUsecase interface {
	SearchSuggestions(ctx context.Context, input SearchSuggestionsInput) (*SearchSuggestionsResult, error)
}

type searchUsecase struct {
	threads ThreadSuggestionSearcher
	moments MomentSuggestionSearcher
}

func NewSearchUsecase(threads ThreadSuggestionSearcher, moments MomentSuggestionSearcher) SearchUsecase {
	return &searchUsecase{threads: threads, moments: moments}
}

// SearchSuggestions implements [SearchUsecase]. A blank query (after
// trimming) short-circuits to an empty result without querying either
// domain — there is nothing to suggest for an empty search box.
func (u *searchUsecase) SearchSuggestions(
	ctx context.Context,
	input SearchSuggestionsInput,
) (*SearchSuggestionsResult, error) {
	query := strings.TrimSpace(input.Query)
	if query == "" {
		return &SearchSuggestionsResult{}, nil
	}
	prefixQuery := buildPrefixTsQuery(query)

	threads, err := u.threads.SearchSuggestions(ctx, input.UserID, query, prefixQuery, suggestionLimit)
	if err != nil {
		return nil, fmt.Errorf("search thread suggestions: %w", err)
	}
	moments, err := u.moments.SearchSuggestions(ctx, input.UserID, query, prefixQuery, suggestionLimit)
	if err != nil {
		return nil, fmt.Errorf("search moment suggestions: %w", err)
	}
	return &SearchSuggestionsResult{Moments: moments, Threads: threads}, nil
}

var (
	tsqueryUnsafeChars = regexp.MustCompile(`[^\p{L}\p{N}\s]`)
	whitespaceRun      = regexp.MustCompile(`\s+`)
)

// buildPrefixTsQuery turns raw user input into a Postgres prefix tsquery
// string, e.g. "morning run" -> "morning:* & run:*", so the last
// (possibly incomplete) word matches as the user is still typing it.
// tsquery-special characters (&, |, !, (, ), :) are stripped first so
// arbitrary input never produces a malformed or unintended-boolean-logic
// query string. An input that is empty or punctuation-only after
// stripping yields an empty string, which to_tsquery('simple', ”)
// accepts and simply matches nothing — no extra guard needed.
func buildPrefixTsQuery(query string) string {
	cleaned := tsqueryUnsafeChars.ReplaceAllString(query, " ")
	words := whitespaceRun.Split(strings.TrimSpace(cleaned), -1)
	terms := make([]string, 0, len(words))
	for _, w := range words {
		if w == "" {
			continue
		}
		terms = append(terms, w+":*")
	}
	return strings.Join(terms, " & ")
}
