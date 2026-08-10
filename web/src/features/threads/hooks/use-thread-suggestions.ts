import { useDebouncedValue } from '@/hooks/use-debounced-value';
import { useGetThreads } from '@/lib/api/generated/threads/threads';

// Before the user types anything, the Thread picker shows this many of
// their most-recently-updated threads (personal + circle-owned) as
// default suggestions.
export const RECENT_THREAD_SUGGESTIONS_LIMIT = 3;

// Once they start typing, results become a name search instead — capped
// separately since a search result list can reasonably show more.
const THREAD_SEARCH_RESULTS_LIMIT = 8;

// Backs the moment-creation Thread picker: default-shows the most
// recently updated threads, then switches to a debounced name search
// once the user types (mirrors useMentionSearch's shape).
export function useThreadSuggestions(query: string) {
  const debouncedQuery = useDebouncedValue(query.trim(), 300);
  const isSearching = debouncedQuery.length > 0;

  const recentQuery = useGetThreads(
    { sort: 'updated_at', page_size: RECENT_THREAD_SUGGESTIONS_LIMIT },
    { query: { enabled: !isSearching } }
  );
  const searchQuery = useGetThreads(
    {
      name: debouncedQuery,
      sort: 'updated_at',
      page_size: THREAD_SEARCH_RESULTS_LIMIT,
    },
    { query: { enabled: isSearching } }
  );

  const active = isSearching ? searchQuery : recentQuery;

  return {
    threads: active.data?.threads ?? [],
    isLoading: active.isPending,
  };
}
