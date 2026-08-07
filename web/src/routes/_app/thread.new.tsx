import Wrapper from '@/components/wrapper';
import { CreateThreadForm } from '@/features/threads';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/thread/new')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <Wrapper>
      <div className="flex flex-col gap-8 max-w-xl">
        <div className="flex items-center">
          <h1 className="text-2xl font-semibold">Create thread</h1>
        </div>

        <CreateThreadForm />
      </div>
    </Wrapper>
  );
}
