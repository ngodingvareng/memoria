import { getCirclesIdAlbum } from '@/lib/api/generated/album/album';
import { useInfiniteQuery } from '@tanstack/react-query';

const PAGE_SIZE = 30;

export function useCircleAlbumImages(circleId: string) {
  return useInfiniteQuery({
    queryKey: ['circles', circleId, 'album'],
    queryFn: ({ pageParam }) =>
      getCirclesIdAlbum(circleId, {
        page: pageParam,
        page_size: PAGE_SIZE,
      }),
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) =>
      (lastPage.images?.length ?? 0) < PAGE_SIZE
        ? undefined
        : allPages.length + 1,
  });
}
