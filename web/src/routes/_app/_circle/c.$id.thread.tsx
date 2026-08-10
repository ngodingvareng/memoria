import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader } from '@/components/ui/empty';
import { ItemGroup } from '@/components/ui/item';
import { ThreadListItem } from '@/features/threads';
import { useGetCirclesIdThreads } from '@/lib/api/generated/threads/threads';
import { PlusSignIcon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute, Link } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_circle/c/$id/thread')({
  component: RouteComponent,
});

function RouteComponent() {
  const { id } = Route.useParams();
  const threadsQuery = useGetCirclesIdThreads(id);
  const threads = threadsQuery.data?.threads ?? [];

  return (
    <div className="flex flex-col gap-8">
      <div className="flex items-center">
        <h1 className="text-2xl font-semibold">Threads</h1>
        <div className="grow justify-end gap-2 flex items-center">
          <Button render={<Link to="/thread/new" search={{ circleId: id }} />}>
            <HugeiconsIcon icon={PlusSignIcon} /> New Thread
          </Button>
        </div>
      </div>

      {threadsQuery.isPending ? (
        <p className="text-muted-foreground">Loading threads…</p>
      ) : threads.length === 0 ? (
        <Empty>
          <EmptyHeader>
            <EmptyDescription>No threads yet</EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <ItemGroup className="grid grid-cols-1 gap-4">
          {threads.map((thread) => (
            <ThreadListItem key={thread.id} thread={thread} />
          ))}
        </ItemGroup>
      )}
    </div>
  );
}
