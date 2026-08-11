import {
  useSearchAutocomplete,
  SearchSuggestionsPopover,
} from '@/features/search';
import { HugeiconsIcon } from '@hugeicons/react';
import { useRef } from 'react';
import { InputGroup, InputGroupAddon, InputGroupInput } from './ui/input-group';
import { Search01Icon } from '@hugeicons/core-free-icons';

export default function SearchAnythingInput() {
  const inputRef = useRef<HTMLInputElement>(null);
  // InputGroup itself doesn't forward a ref, so the popover anchors to
  // this wrapper instead — matching its width (input + icon addon),
  // not just the input's.
  const groupRef = useRef<HTMLDivElement>(null);
  const {
    value,
    isOpen,
    setIsOpen,
    items,
    isLoading,
    activeIndex,
    handleChange,
    handleKeyDown,
    handleSelect,
  } = useSearchAutocomplete();

  return (
    <div ref={groupRef} className="hidden w-full max-w-2xl sm:block">
      <InputGroup className="h-10 w-full [&_svg]:size-5!">
        <InputGroupInput
          ref={inputRef}
          placeholder="Search anything..."
          type="search"
          className="text-base!"
          value={value}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
        />
        <InputGroupAddon>
          <HugeiconsIcon strokeWidth={2} icon={Search01Icon} />
        </InputGroupAddon>

        <SearchSuggestionsPopover
          anchorRef={groupRef}
          inputRef={inputRef}
          open={isOpen}
          onOpenChange={setIsOpen}
          items={items}
          activeIndex={activeIndex}
          isLoading={isLoading}
          onSelect={handleSelect}
        />
      </InputGroup>
    </div>
  );
}
