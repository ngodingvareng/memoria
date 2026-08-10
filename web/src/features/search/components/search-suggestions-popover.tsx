import { cn } from '@/lib/utils';
import { Popover as PopoverPrimitive } from '@base-ui/react/popover';
import type { RefObject } from 'react';
import type { SearchSuggestionItem } from '../hooks/use-search-autocomplete';
import { SearchSuggestionList } from './search-suggestion-list';

interface SearchSuggestionsPopoverProps {
  anchorRef: RefObject<HTMLInputElement | null>;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  items: SearchSuggestionItem[];
  activeIndex: number;
  isLoading: boolean;
  onSelect: (item: SearchSuggestionItem) => void;
}

// Anchored to the search input itself, non-modal so the input keeps
// keyboard focus while suggestions are open — same shape as
// MentionAutocompletePopover.
export function SearchSuggestionsPopover({
  anchorRef,
  open,
  onOpenChange,
  items,
  activeIndex,
  isLoading,
  onSelect,
}: SearchSuggestionsPopoverProps) {
  return (
    <PopoverPrimitive.Root
      open={open}
      onOpenChange={onOpenChange}
      modal={false}
    >
      <PopoverPrimitive.Portal>
        <PopoverPrimitive.Positioner
          anchor={anchorRef}
          side="bottom"
          align="start"
          sideOffset={4}
          className="isolate z-50"
        >
          <PopoverPrimitive.Popup
            initialFocus={anchorRef}
            finalFocus={anchorRef}
            className={cn(
              'bg-popover text-popover-foreground ring-foreground/5 dark:ring-foreground/10 z-50 flex max-h-96 w-(--anchor-width) min-w-72 origin-(--transform-origin) flex-col gap-0.5 overflow-y-auto rounded-2xl p-1.5 text-sm shadow-lg ring-1 outline-hidden'
            )}
          >
            <SearchSuggestionList
              items={items}
              activeIndex={activeIndex}
              isLoading={isLoading}
              onSelect={onSelect}
            />
          </PopoverPrimitive.Popup>
        </PopoverPrimitive.Positioner>
      </PopoverPrimitive.Portal>
    </PopoverPrimitive.Root>
  );
}
