import { Alert, AlertDescription } from '@/components/ui/alert';
import { Empty, EmptyDescription, EmptyHeader } from '@/components/ui/empty';
import { InfiniteScrollSentinel } from '@/components/infinite-scroll-sentinel';
import Wrapper from '@/components/wrapper';
import { MomentFeedList } from '@/features/moments';
import { getMentions } from '@/lib/api/generated/mentions/mentions';
import { useInfiniteQuery } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';

const PAGE_SIZE = 20;

export const Route = createFileRoute('/_app/mentions')({
  component: RouteComponent,
});

function RouteComponent() {
  const momentsQuery = useInfiniteQuery({
    queryKey: ['mentions', 'feed'],
    queryFn: ({ pageParam }) =>
      getMentions({ page: pageParam, page_size: PAGE_SIZE }),
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) =>
      (lastPage.moments?.length ?? 0) < PAGE_SIZE
        ? undefined
        : allPages.length + 1,
  });

  const moments =
    momentsQuery.data?.pages.flatMap((page) => page.moments ?? []) ?? [];

  return (
    <Wrapper>
      <div className="flex flex-col gap-8">
        <h1 className="text-2xl font-semibold">Mentions</h1>

        {momentsQuery.isError && (
          <Alert variant="destructive">
            <AlertDescription>
              Couldn't load your mentions. Please try again.
            </AlertDescription>
          </Alert>
        )}

        {!momentsQuery.isPending &&
          !momentsQuery.isError &&
          moments.length === 0 && (
            <Empty>
              <EmptyHeader>
                <EmptyDescription>
                  No one has mentioned you in a moment yet.
                </EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}

        {moments.length > 0 && <MomentFeedList moments={moments} />}

        <InfiniteScrollSentinel
          enabled={
            Boolean(momentsQuery.hasNextPage) &&
            !momentsQuery.isFetchingNextPage
          }
          isLoading={momentsQuery.isFetchingNextPage}
          onIntersect={() => momentsQuery.fetchNextPage()}
        />
      </div>
    </Wrapper>
  );
}
