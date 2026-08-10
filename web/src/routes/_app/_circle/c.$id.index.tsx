import { Alert, AlertDescription } from '@/components/ui/alert';
import { Empty, EmptyDescription, EmptyHeader } from '@/components/ui/empty';
import { InfiniteScrollSentinel } from '@/components/infinite-scroll-sentinel';
import { MomentFeedList } from '@/features/moments';
import { getCirclesIdMoments } from '@/lib/api/generated/moments/moments';
import { useInfiniteQuery } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';

const PAGE_SIZE = 20;

export const Route = createFileRoute('/_app/_circle/c/$id/')({
  component: RouteComponent,
});

function RouteComponent() {
  const { id } = Route.useParams();

  const momentsQuery = useInfiniteQuery({
    queryKey: ['circles', id, 'moments', 'feed'],
    queryFn: ({ pageParam }) =>
      getCirclesIdMoments(id, { page: pageParam, page_size: PAGE_SIZE }),
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) =>
      (lastPage.moments?.length ?? 0) < PAGE_SIZE
        ? undefined
        : allPages.length + 1,
  });

  const moments =
    momentsQuery.data?.pages.flatMap((page) => page.moments ?? []) ?? [];

  return (
    <div className="flex flex-col gap-4">
      {momentsQuery.isError && (
        <Alert variant="destructive">
          <AlertDescription>
            Couldn't load this circle's moments. Please try again.
          </AlertDescription>
        </Alert>
      )}

      {!momentsQuery.isPending &&
        !momentsQuery.isError &&
        moments.length === 0 && (
          <Empty>
            <EmptyHeader>
              <EmptyDescription>No moments yet.</EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}

      {moments.length > 0 && <MomentFeedList moments={moments} showHeader />}

      <InfiniteScrollSentinel
        enabled={
          Boolean(momentsQuery.hasNextPage) && !momentsQuery.isFetchingNextPage
        }
        isLoading={momentsQuery.isFetchingNextPage}
        onIntersect={() => momentsQuery.fetchNextPage()}
      />
    </div>
  );
}
