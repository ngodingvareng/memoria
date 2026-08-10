import { useDebouncedValue } from '@/hooks/use-debounced-value';
import { useGetMentionsUsers } from '@/lib/api/generated/mentions/mentions';

// Debounced typeahead over GET /mentions/users, shared by the inline
// "@" autocomplete and the "Add mention" search popover.
export function useMentionSearch(query: string) {
  const debouncedQuery = useDebouncedValue(query.trim(), 300);
  const enabled = debouncedQuery.length > 0;

  const searchQuery = useGetMentionsUsers(
    { q: debouncedQuery },
    { query: { enabled } }
  );

  return {
    users: searchQuery.data?.users ?? [],
    isLoading: enabled && searchQuery.isPending,
  };
}
