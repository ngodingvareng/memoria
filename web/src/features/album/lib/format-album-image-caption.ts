import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoAlbumImageResponse as AlbumImage } from '@/lib/api/generated/models';
import dayjs from '@/lib/dayjs';

// The lightbox caption for one Album image: when it was captured, and
// which Thread it belongs to (omitted for a threadless Moment — see
// AlbumImageResponse.thread_name).
export function formatAlbumImageCaption(image: AlbumImage): string | undefined {
  if (!image.occurred_at) return image.thread_name;
  const occurredAt = dayjs(image.occurred_at).format('MMM D, YYYY · h:mm A');
  return image.thread_name
    ? `${occurredAt} · ${image.thread_name}`
    : occurredAt;
}
