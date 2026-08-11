import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse } from '@/lib/api/generated/models';
import dayjs from '@/lib/dayjs';
import { cn } from '@/lib/utils';
import { ThreadCircleBadge } from './thread-circle-badge';

type ThreadOption =
  GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse;

interface ThreadSuggestionsProps {
  threads: ThreadOption[];
  activeIndex: number;
  isLoading?: boolean;
  onSelect: (thread: ThreadOption) => void;
  className?: string;
}

// Compact row list — same shape as MentionSuggestions, sized for a
// picker popover rather than ThreadListItem's full hero card.
export function ThreadSuggestions({
  threads,
  activeIndex,
  isLoading = false,
  onSelect,
  className,
}: ThreadSuggestionsProps) {
  if (threads.length === 0) {
    return (
      <p className="text-muted-foreground px-2 py-1.5 text-sm">
        {isLoading ? 'Loading…' : 'No matching threads.'}
      </p>
    );
  }

  return (
    <div className={cn('flex flex-col gap-0.5', className)} role="listbox">
      {threads.map((thread, index) => (
        <button
          key={thread.id}
          type="button"
          role="option"
          aria-selected={index === activeIndex}
          onMouseDown={(e) => e.preventDefault()}
          onClick={() => onSelect(thread)}
          className={cn(
            'flex items-center gap-2 rounded-xl px-2 py-1.5 text-left outline-none',
            index === activeIndex ? 'bg-accent' : 'hover:bg-accent/50'
          )}
        >
          <span
            className="border-foreground/20 size-2.5 shrink-0 rounded-full border"
            style={{
              backgroundColor: thread.color_hex || undefined,
            }}
          />
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium">{thread.name}</p>
            {thread.updated_at && (
              <p className="text-muted-foreground truncate text-xs">
                Updated {dayjs(thread.updated_at).fromNow()}
              </p>
            )}
          </div>
          {thread.circle_id && (
            <ThreadCircleBadge circleId={thread.circle_id} compact />
          )}
        </button>
      ))}
    </div>
  );
}
