import { MomentCard } from './moment-card';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMomentResponse } from '@/lib/api/generated/models';
import { useGetThreadsId } from '@/lib/api/generated/threads/threads';
import { useGetUserByID } from '@/lib/api/generated/users/users';
import { useSession } from '@/lib/session';

interface MomentFeedItemProps {
  moment: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMomentResponse;
}

// A feed row that resolves what MomentResponse alone doesn't carry —
// the author and the thread's name — before handing off to MomentCard.
// Used by both the home feed (always the viewer's own moments) and a
// Circle's Album feed (moments from any member).
export function MomentFeedItem({ moment }: MomentFeedItemProps) {
  const session = useSession();
  const isOwn = moment.user_id === session?.user.id;

  const userQuery = useGetUserByID(moment.user_id ?? '', {
    query: { enabled: !isOwn && !!moment.user_id },
  });
  const threadQuery = useGetThreadsId(moment.thread_id ?? '', {
    query: { enabled: !!moment.thread_id },
  });

  const user = isOwn
    ? {
        name: session?.user.name ?? '',
        username: session?.user.username ?? '',
        imageSrc: '',
        imageAlt: session?.user.name ?? '',
      }
    : {
        name: userQuery.data?.name ?? '',
        username: userQuery.data?.username ?? '',
        imageSrc: userQuery.data?.image_path ?? '',
        imageAlt: userQuery.data?.name ?? '',
      };

  return (
    <MomentCard
      id={moment.id}
      user={user}
      thread={{ name: threadQuery.data?.name ?? '' }}
      colorHex={moment.color_hex}
      content={moment.note ?? ''}
      createdAt={new Date(moment.created_at ?? moment.occurred_at ?? 0)}
      capturedAt={new Date(moment.occurred_at ?? 0)}
      isPublished
      isOwnedByCurrentUser={isOwn}
    />
  );
}
