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
import {
  Item,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemHeader,
  ItemTitle,
} from '@/components/ui/item';
import { Skeleton } from '@/components/ui/skeleton';
import Wrapper from '@/components/wrapper';
import { ThreadHero } from '@/features/threads';
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
    <Item
      variant="outline"
      render={<Link to="/thread/$id" params={{ id: thread.id! }} />}
    >
      <ItemHeader>
        <ThreadHero
          isReadMode={false}
          imageUrl={imagesQuery.data?.[0]?.url}
          imageAlt={thread.name ?? ''}
          colorHex={thread.color_hex}
          className=" rounded-xl"
        />
      </ItemHeader>
      <ItemContent>
        <ItemTitle>{thread.name}</ItemTitle>
        <ItemDescription>
          {thread.description ||
            (thread.updated_at &&
              formatDistanceToNow(thread.updated_at, {
                addSuffix: true,
              }))}
        </ItemDescription>
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
          <ItemGroup className="grid grid-cols-3 gap-4">
            {threads.map((thread) => (
              <ThreadListCard key={thread.id} thread={thread} />
            ))}
          </ItemGroup>
        )}
      </div>
    </Wrapper>
  );
}
