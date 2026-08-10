import { Item, ItemContent, ItemTitle } from '@/components/ui/item';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse } from '@/lib/api/generated/models';
import { useGetThreadsIdImages } from '@/lib/api/generated/threads/threads';
import { Link } from '@tanstack/react-router';
import { formatDistanceToNow } from 'date-fns';
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
        className="h-full right-0 absolute z-8 rounded-xl"
      />

      <ItemContent className="p-4 bg-linear-to-r from-muted via-muted to-muted/60 z-9">
        <Link
          to="/thread/$id"
          params={{ id: thread.id! }}
          className="contents w-full"
        >
          <ItemTitle className="text-lg">{thread.name}</ItemTitle>
          <div className="flex gap-2 items-center">
            {thread.updated_at && (
              <p className="text-muted-foreground">
                {formatDistanceToNow(thread.updated_at, {
                  addSuffix: true,
                })}
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
