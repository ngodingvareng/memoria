import Wrapper from '@/components/wrapper';
import { KnownPeopleList } from '@/features/users';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_user/user/known-people')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <Wrapper fullWidth>
      <div className="flex max-w-xl flex-col gap-10">
        <div>
          <h1 className="font-heading mb-1 text-xl font-semibold tracking-tight">
            Known people
          </h1>
          <p className="text-muted-foreground mb-6 text-sm">
            People you've marked as known may mention you and invite you to
            circles, depending on your privacy settings.
          </p>
          <KnownPeopleList />
        </div>
      </div>
    </Wrapper>
  );
}
