import { HugeiconsIcon } from '@hugeicons/react';
import { InputGroup, InputGroupAddon, InputGroupInput } from './ui/input-group';
import { Search01Icon } from '@hugeicons/core-free-icons';

export default function SearchAnythingInput() {
  return (
    <InputGroup className="h-10 w-full max-w-2xl [&_svg]:size-5!">
      <InputGroupInput
        placeholder="Search anything..."
        type="search"
        className="text-base!"
      />
      <InputGroupAddon>
        <HugeiconsIcon strokeWidth={2} icon={Search01Icon} />
      </InputGroupAddon>
    </InputGroup>
  );
}
