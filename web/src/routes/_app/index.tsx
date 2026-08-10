import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Empty, EmptyDescription, EmptyHeader } from '@/components/ui/empty';
import { InfiniteScrollSentinel } from '@/components/infinite-scroll-sentinel';
import Wrapper from '@/components/wrapper';
import { MomentFeedList } from '@/features/moments';
import { getMoments } from '@/lib/api/generated/moments/moments';
import { useGetUserByID } from '@/lib/api/generated/users/users';
import { useSession } from '@/lib/session';
import { useInfiniteQuery } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';

const PAGE_SIZE = 20;

export const Route = createFileRoute('/_app/')({
  component: Index,
});

function Index() {
  const session = useSession();
  const userQuery = useGetUserByID(session?.user.id ?? '', {
    query: { enabled: !!session?.user.id },
  });

  const momentsQuery = useInfiniteQuery({
    queryKey: ['moments', 'feed'],
    queryFn: ({ pageParam }) =>
      getMoments({ page: pageParam, page_size: PAGE_SIZE }),
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
      <div className="flex flex-col gap-12">
        <div className="flex items-center">
          <div className="flex gap-2 items-center">
            <Avatar size="xl">
              <AvatarImage
                src={userQuery.data?.image_path ?? undefined}
                alt={session?.user.name}
              />
              <AvatarFallback>
                {session?.user.name?.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <div>
              <p className="font-medium text-2xl/5">{session?.user.name}</p>
              {session?.user.username && (
                <p className="text-lg/5 text-muted-foreground">
                  @{session.user.username}
                </p>
              )}
            </div>
          </div>
        </div>

        <div className="flex flex-col gap-4">
          {momentsQuery.isError && (
            <Alert variant="destructive">
              <AlertDescription>
                Couldn't load your moments. Please try again.
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

          {moments.length > 0 && (
            <MomentFeedList moments={moments} showHeader />
          )}

          <InfiniteScrollSentinel
            enabled={
              Boolean(momentsQuery.hasNextPage) &&
              !momentsQuery.isFetchingNextPage
            }
            isLoading={momentsQuery.isFetchingNextPage}
            onIntersect={() => momentsQuery.fetchNextPage()}
          />
        </div>
      </div>
    </Wrapper>
  );
}
