import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMentionResponse } from '@/lib/api/generated/models';
import { useGetUserByID } from '@/lib/api/generated/users/users';

interface MentionRowProps {
  mention: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMentionResponse;
}

export function MentionRow({ mention }: MentionRowProps) {
  const isAnonymized = !mention.mentioned_user_id;
  const userQuery = useGetUserByID(mention.mentioned_user_id ?? '', {
    query: { enabled: !isAnonymized },
  });

  const name = isAnonymized
    ? mention.display_name
    : (userQuery.data?.name ?? mention.display_name);

  return (
    <div className="flex items-center gap-2 py-2">
      <Avatar size="sm">
        <AvatarImage src={userQuery.data?.image_path ?? undefined} alt={name} />
        <AvatarFallback>{name?.charAt(0).toUpperCase()}</AvatarFallback>
      </Avatar>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">{name}</p>
        {!isAnonymized && userQuery.data?.username && (
          <p className="truncate text-xs text-muted-foreground">
            @{userQuery.data.username}
          </p>
        )}
      </div>
    </div>
  );
}
