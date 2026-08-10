import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { ItemGroup, ItemSeparator } from '@/components/ui/item';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMentionResponse } from '@/lib/api/generated/models';
import { useGetMomentsIdMentions } from '@/lib/api/generated/mentions/mentions';
import { useGetUserByID } from '@/lib/api/generated/users/users';
import { Fragment } from 'react';

interface MentionListProps {
  momentId: string;
}

// Only the Moment's owner can list its mentions (ListMentions is
// owner-scoped server-side, same as ShareToCircle/ListSharedCircles) —
// this component is only ever rendered for the owner. See
// MomentDetail for how a mentioned, non-owner viewer leaves instead.
//
// Read-only: the note text is now the source of truth for mentions
// (added/removed by editing it, see syncMomentMentionsFromText), so
// there's no per-mention remove action here anymore.
export function MentionList({ momentId }: MentionListProps) {
  const mentionsQuery = useGetMomentsIdMentions(momentId);
  const mentions = mentionsQuery.data?.mentions ?? [];

  if (mentionsQuery.isPending) {
    return <p className="text-muted-foreground">Loading mentions…</p>;
  }

  if (mentions.length === 0) {
    return <p className="text-muted-foreground">No one mentioned yet.</p>;
  }

  return (
    <ItemGroup>
      {mentions.map((mention, index) => (
        <Fragment key={mention.id}>
          <MentionRow mention={mention} />
          {index !== mentions.length - 1 && <ItemSeparator />}
        </Fragment>
      ))}
    </ItemGroup>
  );
}

interface MentionRowProps {
  mention: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMentionResponse;
}

function MentionRow({ mention }: MentionRowProps) {
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
