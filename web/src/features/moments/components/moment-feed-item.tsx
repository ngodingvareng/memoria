import { MomentCard } from './moment-card';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMomentResponse } from '@/lib/api/generated/models';
import { useGetThreadsId } from '@/lib/api/generated/threads/threads';
import { useGetUserByID } from '@/lib/api/generated/users/users';
import { useSession } from '@/lib/session';

interface MomentFeedItemProps {
  moment: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMomentResponse;
  // The creator/thread-name row — only Home, a Circle overview, and a
  // Circle-owned thread's own page show it (callers decide per page
  // context).
  showHeader?: boolean;
  onEditNote?: (momentId: string, note: string) => void | Promise<void>;
  onEditColor?: (momentId: string, colorHex: string) => void | Promise<void>;
  onDelete?: (momentId: string) => void | Promise<void>;
}

// A feed row that resolves what MomentResponse alone doesn't carry —
// the actual author (a Circle thread can hold Moments from several
// members, not just the viewer) and the thread's name — before handing
// off to MomentCard. Used by the home feed, a Circle's Album feed, the
// Mentioned Moments feed, and a single Thread's own moment list.
export function MomentFeedItem({
  moment,
  showHeader = false,
  onEditNote,
  onEditColor,
  onDelete,
}: MomentFeedItemProps) {
  const session = useSession();
  const isOwn = moment.user_id === session?.user.id;

  // Fetched regardless of ownership — a Circle thread's own moments can
  // include the viewer's own, and GetUserByID works just as well for
  // "myself" as anyone else, so there's no need for a separate
  // session-only branch here (session.user carries no image_path).
  const userQuery = useGetUserByID(moment.user_id ?? '', {
    query: { enabled: !!moment.user_id },
  });
  const threadQuery = useGetThreadsId(moment.thread_id ?? '', {
    query: { enabled: !!moment.thread_id },
  });

  const user = {
    name: userQuery.data?.name ?? '',
    username: userQuery.data?.username ?? '',
    imageSrc: userQuery.data?.image_path ?? '',
    imageAlt: userQuery.data?.name ?? '',
  };

  return (
    <MomentCard
      id={moment.id}
      user={user}
      thread={{ id: moment.thread_id, name: threadQuery.data?.name ?? '' }}
      colorHex={moment.color_hex}
      content={moment.note ?? ''}
      createdAt={new Date(moment.created_at ?? moment.occurred_at ?? 0)}
      capturedAt={new Date(moment.occurred_at ?? 0)}
      showHeader={showHeader}
      isOwnedByCurrentUser={isOwn}
      onEditNote={
        moment.id && onEditNote
          ? (note) => onEditNote(moment.id!, note)
          : undefined
      }
      onEditColor={
        moment.id && onEditColor
          ? (colorHex) => onEditColor(moment.id!, colorHex)
          : undefined
      }
      onDelete={moment.id && onDelete ? () => onDelete(moment.id!) : undefined}
    />
  );
}
