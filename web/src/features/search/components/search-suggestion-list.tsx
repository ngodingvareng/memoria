import { cn } from '@/lib/utils';
import type { SearchSuggestionItem } from '../hooks/use-search-autocomplete';
import { SearchSuggestionRow } from './search-suggestion-row';

interface SearchSuggestionListProps {
  items: SearchSuggestionItem[];
  activeIndex: number;
  isLoading?: boolean;
  onSelect: (item: SearchSuggestionItem) => void;
  className?: string;
}

// Pure list markup — sectioned by entity type, same shape as
// MentionSuggestions/ThreadSuggestions.
export function SearchSuggestionList({
  items,
  activeIndex,
  isLoading = false,
  onSelect,
  className,
}: SearchSuggestionListProps) {
  if (items.length === 0) {
    return (
      <p className="text-muted-foreground px-2 py-1.5 text-sm">
        {isLoading ? 'Searching…' : 'No matching results.'}
      </p>
    );
  }

  const threads = items.filter((item) => item.type === 'thread');
  const moments = items.filter((item) => item.type === 'moment');

  return (
    <div className={cn('flex flex-col gap-1', className)} role="listbox">
      {threads.length > 0 && (
        <div className="flex flex-col gap-0.5">
          <p className="text-muted-foreground px-2 pt-1 text-xs font-medium">
            Threads
          </p>
          {threads.map((item) => (
            <SearchSuggestionRow
              key={item.data.id}
              item={item}
              active={items.indexOf(item) === activeIndex}
              onSelect={onSelect}
            />
          ))}
        </div>
      )}

      {moments.length > 0 && (
        <div className="flex flex-col gap-0.5">
          <p className="text-muted-foreground px-2 pt-1 text-xs font-medium">
            Moments
          </p>
          {moments.map((item) => (
            <SearchSuggestionRow
              key={item.data.id}
              item={item}
              active={items.indexOf(item) === activeIndex}
              onSelect={onSelect}
            />
          ))}
        </div>
      )}
    </div>
  );
}
