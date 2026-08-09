import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from '@/components/ui/empty';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_circle/c/$id/album')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <Empty>
      <EmptyHeader>
        <EmptyTitle>Coming soon</EmptyTitle>
        <EmptyDescription>
          A circle's shared photo album is on the way.
        </EmptyDescription>
      </EmptyHeader>
    </Empty>
  );
}
