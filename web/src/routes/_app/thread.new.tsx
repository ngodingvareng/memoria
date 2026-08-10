import Wrapper from '@/components/wrapper';
import { CreateThreadForm } from '@/features/threads';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/thread/new')({
  validateSearch: (search: Record<string, unknown>): { circleId?: string } => ({
    circleId: typeof search.circleId === 'string' ? search.circleId : undefined,
  }),
  component: RouteComponent,
});

function RouteComponent() {
  const { circleId } = Route.useSearch();

  return (
    <Wrapper>
      <div className="flex max-w-xl flex-col gap-8">
        <div className="flex items-center">
          <h1 className="text-2xl font-semibold">Create new thread</h1>
        </div>

        <CreateThreadForm circleId={circleId} />
      </div>
    </Wrapper>
  );
}
