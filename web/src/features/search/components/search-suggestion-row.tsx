import dayjs from '@/lib/dayjs';
import { cn } from '@/lib/utils';
import type { SearchSuggestionItem } from '../hooks/use-search-autocomplete';

interface SearchSuggestionRowProps {
  item: SearchSuggestionItem;
  active: boolean;
  onSelect: (item: SearchSuggestionItem) => void;
}

// One suggestion row — Threads and Moments render differently (a
// color-dot title vs. a note preview + date), same idiom as
// ThreadSuggestions' color dot.
export function SearchSuggestionRow({
  item,
  active,
  onSelect,
}: SearchSuggestionRowProps) {
  return (
    <button
      type="button"
      role="option"
      aria-selected={active}
      onMouseDown={(e) => e.preventDefault()}
      onClick={() => onSelect(item)}
      className={cn(
        'flex w-full items-center gap-2 rounded-xl px-2 py-1.5 text-left outline-none',
        active ? 'bg-accent' : 'hover:bg-accent/50'
      )}
    >
      {item.type === 'thread' ? (
        <>
          <span
            className="border-foreground/20 size-2.5 shrink-0 rounded-full border"
            style={{ backgroundColor: item.data.color_hex || undefined }}
          />
          <p className="truncate text-sm font-medium">{item.data.name}</p>
        </>
      ) : (
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium">
            {item.data.note || item.data.place_name || 'Untitled moment'}
          </p>
          {item.data.occurred_at && (
            <p className="text-muted-foreground truncate text-xs">
              {dayjs(item.data.occurred_at).fromNow()}
            </p>
          )}
        </div>
      )}
    </button>
  );
}
