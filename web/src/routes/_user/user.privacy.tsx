import { BlockedMutedUsersList, PrivacySettingsForm } from '@/features/users';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_user/user/privacy')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="max-w-xl flex flex-col gap-10">
      <div>
        <h1 className="font-heading text-xl font-semibold tracking-tight mb-1">
          Privacy
        </h1>
        <p className="text-muted-foreground text-sm mb-6">
          Control who can reach you and what others can see.
        </p>
        <PrivacySettingsForm />
      </div>

      <div>
        <h2 className="font-heading text-lg font-semibold tracking-tight mb-1">
          Blocked &amp; muted
        </h2>
        <p className="text-muted-foreground text-sm mb-2">
          Blocking is mutual; muting only affects your own view.
        </p>
        <BlockedMutedUsersList />
      </div>
    </div>
  );
}
