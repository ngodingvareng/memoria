import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from '@/components/ui/empty';
import { ItemGroup } from '@/components/ui/item';
import { Skeleton } from '@/components/ui/skeleton';
import Wrapper from '@/components/wrapper';
import { ThreadListItem } from '@/features/threads';
import { useGetThreads } from '@/lib/api/generated/threads/threads';
import { ArrowDown01Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute, Link } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/thread/')({
  component: RouteComponent,
});

function RouteComponent() {
  const { data, isPending, isError } = useGetThreads();
  const threads = data?.threads ?? [];

  return (
    <Wrapper>
      <div className="flex flex-col gap-8">
        <div className="flex items-center">
          <h1 className="text-2xl font-semibold">Threads</h1>
          <div className="flex grow items-center justify-end gap-2">
            <DropdownMenu>
              <DropdownMenuTrigger render={<Button variant="secondary" />}>
                <span className="text-muted-foreground font-semibold">
                  Sort by
                </span>{' '}
                Last updated
                <HugeiconsIcon icon={ArrowDown01Icon} strokeWidth={2} />
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuGroup>
                  <DropdownMenuItem>Last updated</DropdownMenuItem>
                  <DropdownMenuItem>Date created</DropdownMenuItem>
                  <DropdownMenuItem>Alphabetical</DropdownMenuItem>
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
            <Button render={<Link to="/thread/new" />}>New Thread</Button>
          </div>
        </div>

        {isPending && (
          <ItemGroup className="grid grid-cols-3 gap-4">
            {Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className="h-24 w-full" />
            ))}
          </ItemGroup>
        )}

        {isError && (
          <Alert variant="destructive">
            <AlertDescription>
              Couldn't load your threads. Please try again.
            </AlertDescription>
          </Alert>
        )}

        {!isPending && !isError && threads.length === 0 && (
          <Empty>
            <EmptyHeader>
              <EmptyTitle>No threads yet</EmptyTitle>
              <EmptyDescription>
                Create a thread to start capturing moments.
              </EmptyDescription>
            </EmptyHeader>
            <Button render={<Link to="/thread/new" />}>New Thread</Button>
          </Empty>
        )}

        {!isPending && !isError && threads.length > 0 && (
          <ItemGroup className="grid grid-cols-1 gap-4">
            {threads.map((thread) => (
              <ThreadListItem key={thread.id} thread={thread} />
            ))}
          </ItemGroup>
        )}
      </div>
    </Wrapper>
  );
}
