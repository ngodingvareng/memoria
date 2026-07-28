import Wrapper from '@/components/wrapper';
import { CreateGroupForm } from '@/features/groups';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/group/new')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <Wrapper>
      <div className="flex flex-col gap-8 max-w-xl">
        <div className="flex items-center">
          <h1 className="text-2xl font-semibold">Create group</h1>
        </div>

        <CreateGroupForm />
      </div>
    </Wrapper>
  );
}
