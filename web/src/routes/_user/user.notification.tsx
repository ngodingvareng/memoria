import Wrapper from '@/components/wrapper';
import { NotificationPreferencesForm } from '@/features/notifications';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_user/user/notification')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <Wrapper fullWidth>
      <div className="max-w-xl">
        <h1 className="font-heading mb-1 text-xl font-semibold tracking-tight">
          Notifications
        </h1>
        <p className="text-muted-foreground mb-6 text-sm">
          Choose what you hear about, and when.
        </p>
        <NotificationPreferencesForm />
      </div>
    </Wrapper>
  );
}
