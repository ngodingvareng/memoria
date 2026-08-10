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
    <InputGroup className="h-10 w-full max-w-2xl [&_svg]:size-5!">
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
        anchorRef={inputRef}
        open={isOpen}
        onOpenChange={setIsOpen}
        items={items}
        activeIndex={activeIndex}
        isLoading={isLoading}
        onSelect={handleSelect}
      />
    </InputGroup>
  );
}
