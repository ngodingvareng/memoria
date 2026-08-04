import Wrapper from '@/components/wrapper';
import { CreateCircleForm } from '@/features/circles';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/circle/new')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <Wrapper>
      <div className="flex flex-col gap-8 max-w-xl">
        <div className="flex items-center">
          <h1 className="text-2xl font-semibold">Create circle</h1>
        </div>

        <CreateCircleForm />
      </div>
    </Wrapper>
  );
}
