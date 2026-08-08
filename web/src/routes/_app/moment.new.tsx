import Wrapper from '@/components/wrapper';
import { CreateMomentForm } from '@/features/moments';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/moment/new')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <Wrapper>
      <div className="flex flex-col gap-8 max-w-xl">
        <div className="flex items-center">
          <h1 className="text-2xl font-semibold">Create new moment</h1>
        </div>

        <CreateMomentForm />
      </div>
    </Wrapper>
  );
}
