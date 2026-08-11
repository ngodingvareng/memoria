import Wrapper from '@/components/wrapper';
import { BlockedMutedUsersList, PrivacySettingsForm } from '@/features/users';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_user/user/privacy')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <Wrapper fullWidth>
      <div className="flex max-w-xl flex-col gap-10">
        <div>
          <h1 className="font-heading mb-1 text-xl font-semibold tracking-tight">
            Privacy
          </h1>
          <p className="text-muted-foreground mb-6 text-sm">
            Control who can reach you and what others can see.
          </p>
          <PrivacySettingsForm />
        </div>

        <div>
          <h2 className="font-heading mb-1 text-lg font-semibold tracking-tight">
            Blocked &amp; muted
          </h2>
          <p className="text-muted-foreground mb-2 text-sm">
            Blocking is mutual; muting only affects your own view.
          </p>
          <BlockedMutedUsersList />
        </div>
      </div>
    </Wrapper>
  );
}
