import { Button } from '@/components/ui/button';
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from '@/components/ui/input-group';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse } from '@/lib/api/generated/models';
import { ArrowDown01Icon, SearchIcon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { useState } from 'react';
import { useThreadSuggestions } from '../hooks/use-thread-suggestions';
import { ThreadCircleBadge } from './thread-circle-badge';
import { ThreadSuggestions } from './thread-suggestions';

type ThreadOption =
  GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse;

interface ThreadPickerFieldProps {
  selectedThread: ThreadOption | null;
  onSelect: (thread: ThreadOption) => void;
  onBlur?: () => void;
  className?: string;
}

// Required "which thread is this Moment part of" field for the create
// form — same search-popover pattern as AddMentionButton, but rendered
// as a full field showing the current selection rather than a bare
// icon button, and default-populated with the most recently updated
// threads instead of staying empty until typed into.
export function ThreadPickerField({
  selectedThread,
  onSelect,
  onBlur,
  className,
}: ThreadPickerFieldProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const { threads, isLoading } = useThreadSuggestions(query);

  const handleSelect = (thread: ThreadOption) => {
    onSelect(thread);
    setOpen(false);
    setQuery('');
  };

  return (
    <Popover
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (!next) {
          setQuery('');
          onBlur?.();
        }
      }}
    >
      <PopoverTrigger
        render={
          <Button type="button" variant="secondary" className={className} />
        }
      >
        <span className="flex min-w-0 flex-1 items-center gap-2 text-left">
          {selectedThread ? (
            <>
              <span
                className="size-2.5 shrink-0 rounded-full border border-foreground/20"
                style={{
                  backgroundColor: selectedThread.color_hex || undefined,
                }}
              />
              <span className="min-w-0 truncate">{selectedThread.name}</span>
              {selectedThread.circle_id && (
                <ThreadCircleBadge
                  circleId={selectedThread.circle_id}
                  compact
                />
              )}
            </>
          ) : (
            <span className="text-muted-foreground">Choose a thread</span>
          )}
        </span>
        <HugeiconsIcon
          strokeWidth={2.5}
          icon={ArrowDown01Icon}
          className="text-muted-foreground"
        />
      </PopoverTrigger>
      <PopoverContent align="start" className="w-80 gap-2 p-2">
        <InputGroup>
          <InputGroupAddon>
            <HugeiconsIcon icon={SearchIcon} />
          </InputGroupAddon>
          <InputGroupInput
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search threads"
            autoComplete="off"
            autoFocus
          />
        </InputGroup>
        <ThreadSuggestions
          threads={threads}
          activeIndex={-1}
          isLoading={isLoading}
          onSelect={handleSelect}
        />
      </PopoverContent>
    </Popover>
  );
}
