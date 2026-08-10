import { ItemGroup, ItemSeparator } from '@/components/ui/item';
import { useGetMomentsIdMentions } from '@/lib/api/generated/mentions/mentions';
import { Fragment } from 'react';
import { MentionRow } from './mention-row';

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
