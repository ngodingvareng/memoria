import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader } from '@/components/ui/empty';
import {
  Item,
  ItemContent,
  ItemGroup,
  ItemHeader,
  ItemTitle,
} from '@/components/ui/item';
import Wrapper from '@/components/wrapper';
import { PendingCircleInvites } from '@/features/circles';
import { useGetCircles } from '@/lib/api/generated/circles/circles';
import { PlusSignIcon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute, Link } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/circle/')({
  component: RouteComponent,
});

function RouteComponent() {
  const circlesQuery = useGetCircles();
  const circles = circlesQuery.data?.circles ?? [];

  return (
    <Wrapper>
      <div className="flex flex-col gap-12">
        <PendingCircleInvites />

        <div className="flex flex-col gap-4">
          <div className="flex items-center">
            <h1 className="text-2xl font-semibold">My circles</h1>
            <div className="grow justify-end gap-2 flex items-center">
              <Button render={<Link to="/circle/new" />}>
                <HugeiconsIcon icon={PlusSignIcon} /> New circle
              </Button>
            </div>
          </div>

          {circlesQuery.isPending ? (
            <p className="text-muted-foreground">Loading circles…</p>
          ) : circles.length === 0 ? (
            <Empty>
              <EmptyHeader>
                <EmptyDescription>
                  You're not part of any circle yet.
                </EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <ItemGroup className="grid grid-cols-1 gap-4">
              {circles.map((circle) => (
                <Item
                  key={circle.id}
                  variant="outline"
                  render={
                    <Link to="/c/$id" params={{ id: circle.id! }}>
                      <ItemContent className='flex gap-4'>
                        <Avatar
                          size="xl"
                          className="rounded-xl after:rounded-xl"
                        >
                          <AvatarImage
                            src={circle.image_path ?? undefined}
                            alt={circle.name}
                            className="rounded-xl"
                          />
                          <AvatarFallback className="rounded-xl">
                            {circle.name?.charAt(0).toUpperCase()}
                          </AvatarFallback>
                        </Avatar>

                        <ItemTitle>{circle.name}</ItemTitle>
                      </ItemContent>
                    </Link>
                  }
                />
              ))}
            </ItemGroup>
          )}
        </div>
      </div>
    </Wrapper>
  );
}
