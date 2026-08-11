import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoAlbumImageResponse as AlbumImage } from '@/lib/api/generated/models';
import dayjs from '@/lib/dayjs';

export interface AlbumDateBucket {
  date: string;
  images: AlbumImage[];
}

// The server already returns images newest-occurrence-first (see
// ListAlbumImages/ListCircleAlbumImages' ORDER BY) — this only groups
// consecutive same-local-day images into buckets, it never re-sorts.
export function groupAlbumImagesByDate(
  images: AlbumImage[]
): AlbumDateBucket[] {
  const buckets: AlbumDateBucket[] = [];
  for (const image of images) {
    if (!image.occurred_at) continue;
    const occurredAt = dayjs(image.occurred_at);
    const last = buckets[buckets.length - 1];
    const lastOccurredAt = last?.images[0]?.occurred_at;
    if (last && lastOccurredAt && occurredAt.isSame(lastOccurredAt, 'day')) {
      last.images.push(image);
    } else {
      buckets.push({
        date: occurredAt.format('MMMM D, YYYY'),
        images: [image],
      });
    }
  }
  return buckets;
}
