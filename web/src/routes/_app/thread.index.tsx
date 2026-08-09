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
import { Item, ItemContent, ItemGroup, ItemTitle } from '@/components/ui/item';
import { Skeleton } from '@/components/ui/skeleton';
import Wrapper from '@/components/wrapper';
import { ThreadCircleBadge, ThreadHero } from '@/features/threads';
import {
  useGetThreads,
  useGetThreadsIdImages,
} from '@/lib/api/generated/threads/threads';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse } from '@/lib/api/generated/models';
import { ArrowDown01Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute, Link } from '@tanstack/react-router';
import { formatDistanceToNow } from 'date-fns';

export const Route = createFileRoute('/_app/thread/')({
  component: RouteComponent,
});

function ThreadListCard({
  thread,
}: {
  thread: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse;
}) {
  const imagesQuery = useGetThreadsIdImages(thread.id!);

  return (
    <Item className="relative overflow-hidden p-0">
      <ThreadHero
        isReadMode={false}
        imageUrl={imagesQuery.data?.[0]?.url}
        imageAlt={thread.name ?? ''}
        colorHex={thread.color_hex}
        className="h-full right-0 absolute z-8 rounded-xl"
      />

      <ItemContent className='p-4 bg-linear-to-r from-muted via-muted to-muted/60 z-9'>
        <Link to="/thread/$id" params={{ id: thread.id! }} className="contents">
          <ItemTitle className="text-lg">{thread.name}</ItemTitle>
          <div className="flex gap-2 items-center">
            {thread.updated_at && (
              <p className="text-muted-foreground">
                {formatDistanceToNow(thread.updated_at, {
                  addSuffix: true,
                })}
              </p>
            )}

            {thread.circle_id && (
              <ThreadCircleBadge circleId={thread.circle_id} />
            )}
          </div>
        </Link>
      </ItemContent>
    </Item>
  );
}

function RouteComponent() {
  const { data, isPending, isError } = useGetThreads();
  const threads = data?.threads ?? [];

  return (
    <Wrapper>
      <div className="flex flex-col gap-8">
        <div className="flex items-center">
          <h1 className="text-2xl font-semibold">Threads</h1>
          <div className="grow justify-end gap-2 flex items-center">
            <DropdownMenu>
              <DropdownMenuTrigger render={<Button variant="secondary" />}>
                <span className="font-semibold text-muted-foreground">
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
              <ThreadListCard key={thread.id} thread={thread} />
            ))}
          </ItemGroup>
        )}
      </div>
    </Wrapper>
  );
}
