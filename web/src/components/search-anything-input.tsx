import { HugeiconsIcon } from '@hugeicons/react';
import { InputGroup, InputGroupAddon, InputGroupInput } from './ui/input-group';
import { Search01Icon } from '@hugeicons/core-free-icons';

export default function SearchAnythingInput() {
  return (
    <InputGroup className="max-w-xl w-full h-10 [&_svg]:size-5!">
      <InputGroupInput
        placeholder="Search for stories and activities..."
        type="search"
        className="text-base!"
      />
      <InputGroupAddon>
        <HugeiconsIcon strokeWidth={2} icon={Search01Icon} />
      </InputGroupAddon>
    </InputGroup>
  );
}
