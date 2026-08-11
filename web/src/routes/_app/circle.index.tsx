import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader } from '@/components/ui/empty';
import {
  Item,
  ItemContent,
  ItemGroup,
  ItemMedia,
  ItemTitle,
} from '@/components/ui/item';
import Wrapper from '@/components/wrapper';
import { CircleProfileImage, PendingCircleInvites } from '@/features/circles';
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
            <div className="flex grow items-center justify-end gap-2">
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
                  render={<Link to="/c/$id" params={{ id: circle.id! }} />}
                >
                  <ItemMedia variant="image">
                    <CircleProfileImage circle={circle} size="xl" />
                  </ItemMedia>
                  <ItemContent className="flex gap-4">
                    <ItemTitle className="text-lg">{circle.name}</ItemTitle>
                  </ItemContent>
                </Item>
              ))}
            </ItemGroup>
          )}
        </div>
      </div>
    </Wrapper>
  );
}
