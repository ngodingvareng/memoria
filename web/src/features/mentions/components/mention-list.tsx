import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { ItemGroup, ItemSeparator } from '@/components/ui/item';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMentionResponse } from '@/lib/api/generated/models';
import {
  getGetMomentsIdMentionsQueryKey,
  useDeleteMomentsIdMentionsMentionId,
  useGetMomentsIdMentions,
} from '@/lib/api/generated/mentions/mentions';
import { useGetUserByID } from '@/lib/api/generated/users/users';
import { queryClient } from '@/lib/query-client';
import { Fragment } from 'react';

interface MentionListProps {
  momentId: string;
}

// Only the Moment's owner can list its mentions (ListMentions is
// owner-scoped server-side, same as ShareToCircle/ListSharedCircles) —
// this component is only ever rendered for the owner. See
// MomentDetail for how a mentioned, non-owner viewer leaves instead.
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
          <MentionRow momentId={momentId} mention={mention} />
          {index !== mentions.length - 1 && <ItemSeparator />}
        </Fragment>
      ))}
    </ItemGroup>
  );
}

interface MentionRowProps {
  momentId: string;
  mention: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMentionResponse;
}

function MentionRow({ momentId, mention }: MentionRowProps) {
  const isAnonymized = !mention.mentioned_user_id;
  const userQuery = useGetUserByID(mention.mentioned_user_id ?? '', {
    query: { enabled: !isAnonymized },
  });
  const deleteMention = useDeleteMomentsIdMentionsMentionId();

  const name = isAnonymized
    ? mention.display_name
    : (userQuery.data?.name ?? mention.display_name);

  const handleRemove = async () => {
    await deleteMention.mutateAsync({ id: momentId, mentionId: mention.id! });
    await queryClient.invalidateQueries({
      queryKey: getGetMomentsIdMentionsQueryKey(momentId),
    });
  };

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
      <Button
        type="button"
        size="sm"
        variant="ghost"
        onClick={handleRemove}
        disabled={deleteMention.isPending}
      >
        Remove
      </Button>
    </div>
  );
}
