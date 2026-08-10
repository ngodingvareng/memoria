import { Button } from '@/components/ui/button';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoPublicUserResponse } from '@/lib/api/generated/models';
import { AtIcon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { useState } from 'react';
import { useMentionSearch } from '../hooks/use-mention-search';
import { MentionSuggestions } from './mention-suggestions';
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from '@/components/ui/input-group';

type MentionableUser =
  GithubComNgodingvarengMemoriaInternalDeliveryRestDtoPublicUserResponse;

interface AddMentionButtonProps {
  onInsert: (user: MentionableUser) => void;
}

// The alternative to typing "@" inline: a search popover for picking a
// user to mention, opened from a button rather than triggered by text.
export function AddMentionButton({ onInsert }: AddMentionButtonProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const { users, isLoading } = useMentionSearch(query);

  const handleSelect = (user: MentionableUser) => {
    onInsert(user);
    setOpen(false);
    setQuery('');
  };

  return (
    <Popover
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (!next) setQuery('');
      }}
    >
      <PopoverTrigger render={<Button variant="secondary" size="icon-sm" />}>
        <HugeiconsIcon strokeWidth={2.5} icon={AtIcon} />
        <span className="sr-only">Add mention</span>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-72 gap-2 p-2">
        <InputGroup>
          <InputGroupAddon>
            <HugeiconsIcon icon={AtIcon} />
          </InputGroupAddon>
          <InputGroupInput
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search username"
            autoComplete="off"
            autoFocus
          />
        </InputGroup>
        <MentionSuggestions
          users={users}
          activeIndex={-1}
          isLoading={isLoading}
          onSelect={handleSelect}
        />
      </PopoverContent>
    </Popover>
  );
}
