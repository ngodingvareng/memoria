import { format, isSameDay, parseISO } from 'date-fns';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoAlbumImageResponse as AlbumImage } from '@/lib/api/generated/models';

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
    const occurredAt = parseISO(image.occurred_at);
    const last = buckets[buckets.length - 1];
    const lastOccurredAt = last?.images[0]?.occurred_at;
    if (
      last &&
      lastOccurredAt &&
      isSameDay(parseISO(lastOccurredAt), occurredAt)
    ) {
      last.images.push(image);
    } else {
      buckets.push({
        date: format(occurredAt, 'MMMM d, yyyy'),
        images: [image],
      });
    }
  }
  return buckets;
}
