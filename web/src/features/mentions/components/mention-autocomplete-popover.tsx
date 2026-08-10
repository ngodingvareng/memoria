import { Popover as PopoverPrimitive } from '@base-ui/react/popover';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoPublicUserResponse } from '@/lib/api/generated/models';
import { cn } from '@/lib/utils';
import type { RefObject } from 'react';
import { MentionSuggestions } from './mention-suggestions';

type MentionableUser =
  GithubComNgodingvarengMemoriaInternalDeliveryRestDtoPublicUserResponse;

interface MentionAutocompletePopoverProps {
  anchorRef: RefObject<HTMLTextAreaElement | null>;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  users: MentionableUser[];
  activeIndex: number;
  isLoading: boolean;
  onSelect: (user: MentionableUser) => void;
}

// Anchored to the textarea itself, not the caret — true caret-tracking
// in a plain <textarea> needs a mirror-div measurement technique, and
// anchoring below the textarea is the standard simplification (Slack's
// web compose box does the same). Non-modal so the textarea keeps
// keyboard focus while suggestions are open.
export function MentionAutocompletePopover({
  anchorRef,
  open,
  onOpenChange,
  users,
  activeIndex,
  isLoading,
  onSelect,
}: MentionAutocompletePopoverProps) {
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
              'bg-popover text-popover-foreground ring-foreground/5 dark:ring-foreground/10 z-50 flex max-h-64 w-72 origin-(--transform-origin) flex-col gap-0.5 overflow-y-auto rounded-2xl p-1.5 text-sm shadow-lg ring-1 outline-hidden'
            )}
          >
            <MentionSuggestions
              users={users}
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
