import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoPublicUserResponse } from '@/lib/api/generated/models';
import { cn } from '@/lib/utils';

type MentionableUser =
  GithubComNgodingvarengMemoriaInternalDeliveryRestDtoPublicUserResponse;

interface MentionSuggestionsProps {
  users: MentionableUser[];
  activeIndex: number;
  isLoading?: boolean;
  onSelect: (user: MentionableUser) => void;
  className?: string;
}

// Pure list markup — the popover chrome around it differs between the
// inline "@" autocomplete (anchored to the textarea) and the "Add
// mention" button's popover (anchored to its trigger), so callers
// supply their own wrapper.
export function MentionSuggestions({
  users,
  activeIndex,
  isLoading = false,
  onSelect,
  className,
}: MentionSuggestionsProps) {
  if (users.length === 0) {
    return (
      <p className="text-muted-foreground px-2 py-1.5 text-sm">
        {isLoading ? 'Searching…' : 'No matching users.'}
      </p>
    );
  }

  return (
    <div className={cn('flex flex-col gap-0.5', className)} role="listbox">
      {users.map((user, index) => (
        <button
          key={user.id}
          type="button"
          role="option"
          aria-selected={index === activeIndex}
          onMouseDown={(e) => e.preventDefault()}
          onClick={() => onSelect(user)}
          className={cn(
            'flex items-center gap-2 rounded-xl px-2 py-1.5 text-left outline-none',
            index === activeIndex ? 'bg-accent' : 'hover:bg-accent/50'
          )}
        >
          <Avatar size="sm">
            <AvatarImage src={user.image_path ?? undefined} alt={user.name} />
            <AvatarFallback>
              {user.name?.charAt(0).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium">{user.name}</p>
            {user.username && (
              <p className="text-muted-foreground truncate text-xs">
                @{user.username}
              </p>
            )}
          </div>
        </button>
      ))}
    </div>
  );
}
