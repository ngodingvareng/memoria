import { Item, ItemContent, ItemTitle } from '@/components/ui/item';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse } from '@/lib/api/generated/models';
import { useGetThreadsIdImages } from '@/lib/api/generated/threads/threads';
import dayjs from '@/lib/dayjs';
import { Link } from '@tanstack/react-router';
import { ThreadCircleBadge } from './thread-circle-badge';
import { ThreadHero } from './thread-hero';

interface ThreadListItemProps {
  thread: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse;
}

export function ThreadListItem({ thread }: ThreadListItemProps) {
  const imagesQuery = useGetThreadsIdImages(thread.id!);

  return (
    <Item className="relative overflow-hidden p-0">
      <ThreadHero
        isReadMode={false}
        imageUrl={imagesQuery.data?.[0]?.url}
        imageAlt={thread.name ?? ''}
        colorHex={thread.color_hex}
        className="absolute right-0 z-8 h-full rounded-xl"
      />

      <ItemContent className="from-muted via-muted to-muted/60 z-9 bg-linear-to-r p-4">
        <Link
          to="/thread/$id"
          params={{ id: thread.id! }}
          className="contents w-full"
        >
          <ItemTitle className="text-lg">{thread.name}</ItemTitle>
          <div className="flex items-center gap-2">
            {thread.updated_at && (
              <p className="text-muted-foreground">
                {dayjs(thread.updated_at).fromNow()}
              </p>
            )}

            {thread.circle_id && (
              <ThreadCircleBadge circleId={thread.circle_id} />
            )}
          </div>
        </Link>
      </ItemContent>
    </Item>
  );
}
