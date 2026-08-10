import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from '@/components/ui/empty';
import { ItemGroup } from '@/components/ui/item';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import Wrapper from '@/components/wrapper';
import { MomentFeedList } from '@/features/moments';
import { ThreadListItem } from '@/features/threads';
import { useGetMomentsSearch } from '@/lib/api/generated/moments/moments';
import { useGetThreads } from '@/lib/api/generated/threads/threads';
import { createFileRoute } from '@tanstack/react-router';
import { useState } from 'react';

export const Route = createFileRoute('/_app/search')({
  validateSearch: (search: Record<string, unknown>): { q: string } => ({
    q: typeof search.q === 'string' ? search.q : '',
  }),
  component: RouteComponent,
});

const PAGE_SIZE = 20;

function RouteComponent() {
  const { q } = Route.useSearch();
  const [momentsPage, setMomentsPage] = useState(1);
  const [threadsPage, setThreadsPage] = useState(1);

  const momentsQuery = useGetMomentsSearch(
    { q, page: momentsPage, page_size: PAGE_SIZE },
    { query: { enabled: q.length > 0 } }
  );
  const threadsQuery = useGetThreads(
    { name: q, page: threadsPage, page_size: PAGE_SIZE },
    { query: { enabled: q.length > 0 } }
  );

  const moments = momentsQuery.data?.moments ?? [];
  const threads = threadsQuery.data?.threads ?? [];

  return (
    <Wrapper>
      <div className="flex flex-col gap-8">
        <h1 className="text-2xl font-semibold">
          {q ? `Search results for "${q}"` : 'Search'}
        </h1>

        {q.length === 0 ? (
          <Empty>
            <EmptyHeader>
              <EmptyTitle>Nothing to search for yet</EmptyTitle>
              <EmptyDescription>
                Type in the search box above to find Moments and Threads.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <Tabs defaultValue="moments">
            <TabsList>
              <TabsTrigger value="moments">Moments</TabsTrigger>
              <TabsTrigger value="threads">Threads</TabsTrigger>
            </TabsList>

            <TabsContent value="moments" className="flex flex-col gap-4 pt-4">
              {momentsQuery.isPending && (
                <div className="flex flex-col gap-2">
                  {Array.from({ length: 3 }).map((_, i) => (
                    <Skeleton key={i} className="h-16 w-full" />
                  ))}
                </div>
              )}

              {momentsQuery.isError && (
                <Alert variant="destructive">
                  <AlertDescription>
                    Couldn't load matching moments. Please try again.
                  </AlertDescription>
                </Alert>
              )}

              {!momentsQuery.isPending &&
                !momentsQuery.isError &&
                moments.length === 0 && (
                  <Empty>
                    <EmptyHeader>
                      <EmptyTitle>No matching moments</EmptyTitle>
                      <EmptyDescription>
                        Try a different search term.
                      </EmptyDescription>
                    </EmptyHeader>
                  </Empty>
                )}

              {!momentsQuery.isPending &&
                !momentsQuery.isError &&
                moments.length > 0 && (
                  <>
                    <MomentFeedList moments={moments} showHeader />
                    <div className="flex items-center justify-between">
                      <Button
                        variant="secondary"
                        disabled={momentsPage <= 1}
                        onClick={() => setMomentsPage((p) => p - 1)}
                      >
                        Previous
                      </Button>
                      <Button
                        variant="secondary"
                        disabled={moments.length < PAGE_SIZE}
                        onClick={() => setMomentsPage((p) => p + 1)}
                      >
                        Next
                      </Button>
                    </div>
                  </>
                )}
            </TabsContent>

            <TabsContent value="threads" className="flex flex-col gap-4 pt-4">
              {threadsQuery.isPending && (
                <ItemGroup className="grid grid-cols-1 gap-4">
                  {Array.from({ length: 3 }).map((_, i) => (
                    <Skeleton key={i} className="h-24 w-full" />
                  ))}
                </ItemGroup>
              )}

              {threadsQuery.isError && (
                <Alert variant="destructive">
                  <AlertDescription>
                    Couldn't load matching threads. Please try again.
                  </AlertDescription>
                </Alert>
              )}

              {!threadsQuery.isPending &&
                !threadsQuery.isError &&
                threads.length === 0 && (
                  <Empty>
                    <EmptyHeader>
                      <EmptyTitle>No matching threads</EmptyTitle>
                      <EmptyDescription>
                        Try a different search term.
                      </EmptyDescription>
                    </EmptyHeader>
                  </Empty>
                )}

              {!threadsQuery.isPending &&
                !threadsQuery.isError &&
                threads.length > 0 && (
                  <>
                    <ItemGroup className="grid grid-cols-1 gap-4">
                      {threads.map((thread) => (
                        <ThreadListItem key={thread.id} thread={thread} />
                      ))}
                    </ItemGroup>
                    <div className="flex items-center justify-between">
                      <Button
                        variant="secondary"
                        disabled={threadsPage <= 1}
                        onClick={() => setThreadsPage((p) => p - 1)}
                      >
                        Previous
                      </Button>
                      <Button
                        variant="secondary"
                        disabled={threads.length < PAGE_SIZE}
                        onClick={() => setThreadsPage((p) => p + 1)}
                      >
                        Next
                      </Button>
                    </div>
                  </>
                )}
            </TabsContent>
          </Tabs>
        )}
      </div>
    </Wrapper>
  );
}
