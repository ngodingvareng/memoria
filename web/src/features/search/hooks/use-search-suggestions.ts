import { useDebouncedValue } from '@/hooks/use-debounced-value';
import { useGetSearchSuggestions } from '@/lib/api/generated/search/search';

// Debounced typeahead over GET /search/suggestions — top-5 Threads and
// top-5 Moments matching the live-typing search box.
export function useSearchSuggestions(query: string) {
  const debouncedQuery = useDebouncedValue(query.trim(), 300);
  const enabled = debouncedQuery.length > 0;

  const searchQuery = useGetSearchSuggestions(
    { q: debouncedQuery },
    { query: { enabled } }
  );

  return {
    threads: searchQuery.data?.threads ?? [],
    moments: searchQuery.data?.moments ?? [],
    isLoading: enabled && searchQuery.isPending,
  };
}
