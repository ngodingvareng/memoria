//go:build unit

package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSearchUsecase_SearchSuggestions_BlankQueryShortCircuits(t *testing.T) {
	threads := mocks.NewMockThreadSuggestionSearcher(t)
	moments := mocks.NewMockMomentSuggestionSearcher(t)
	uc := usecase.NewSearchUsecase(threads, moments)

	result, err := uc.SearchSuggestions(context.Background(), usecase.SearchSuggestionsInput{
		UserID: uuid.New(),
		Query:  "   ",
	})

	assert.NoError(t, err)
	assert.Empty(t, result.Moments)
	assert.Empty(t, result.Threads)
	// Neither mock is stubbed above; testify/mock fails the test if
	// either is called unexpectedly for a blank query.
}

func TestSearchUsecase_SearchSuggestions_CallsBothSearchersWithPrefixQuery(t *testing.T) {
	threads := mocks.NewMockThreadSuggestionSearcher(t)
	moments := mocks.NewMockMomentSuggestionSearcher(t)
	uc := usecase.NewSearchUsecase(threads, moments)

	userID := uuid.New()
	threadResult := []*entity.Thread{{ID: uuid.New(), Name: "Morning run"}}
	momentResult := []*entity.Moment{{ID: uuid.New()}}

	threads.EXPECT().SearchSuggestions(mock.Anything, userID, "morning", "morning:*", int32(5)).
		Return(threadResult, nil)
	moments.EXPECT().SearchSuggestions(mock.Anything, userID, "morning", "morning:*", int32(5)).
		Return(momentResult, nil)

	result, err := uc.SearchSuggestions(context.Background(), usecase.SearchSuggestionsInput{
		UserID: userID,
		Query:  "  morning  ",
	})

	assert.NoError(t, err)
	assert.Equal(t, threadResult, result.Threads)
	assert.Equal(t, momentResult, result.Moments)
}

func TestSearchUsecase_SearchSuggestions_ThreadRepoErrorPropagates(t *testing.T) {
	threads := mocks.NewMockThreadSuggestionSearcher(t)
	moments := mocks.NewMockMomentSuggestionSearcher(t)
	uc := usecase.NewSearchUsecase(threads, moments)

	userID := uuid.New()
	threads.EXPECT().SearchSuggestions(mock.Anything, userID, "gym", "gym:*", int32(5)).
		Return(nil, errors.New("db unavailable"))

	result, err := uc.SearchSuggestions(context.Background(), usecase.SearchSuggestionsInput{
		UserID: userID,
		Query:  "gym",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	// moments is never stubbed above; testify/mock fails the test if
	// it's called after the thread search already failed.
}

func TestSearchUsecase_SearchSuggestions_MomentRepoErrorPropagates(t *testing.T) {
	threads := mocks.NewMockThreadSuggestionSearcher(t)
	moments := mocks.NewMockMomentSuggestionSearcher(t)
	uc := usecase.NewSearchUsecase(threads, moments)

	userID := uuid.New()
	threads.EXPECT().SearchSuggestions(mock.Anything, userID, "gym", "gym:*", int32(5)).
		Return(nil, nil)
	moments.EXPECT().SearchSuggestions(mock.Anything, userID, "gym", "gym:*", int32(5)).
		Return(nil, errors.New("db unavailable"))

	result, err := uc.SearchSuggestions(context.Background(), usecase.SearchSuggestionsInput{
		UserID: userID,
		Query:  "gym",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
}
