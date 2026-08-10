import { getAlbum } from '@/lib/api/generated/album/album';
import { useInfiniteQuery } from '@tanstack/react-query';

const PAGE_SIZE = 30; // higher than the 20 used for Moment feeds — Album is a photo grid, not full cards

export function useAlbumImages() {
  return useInfiniteQuery({
    queryKey: ['album'],
    queryFn: ({ pageParam }) =>
      getAlbum({ page: pageParam, page_size: PAGE_SIZE }),
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) =>
      (lastPage.images?.length ?? 0) < PAGE_SIZE
        ? undefined
        : allPages.length + 1,
  });
}
