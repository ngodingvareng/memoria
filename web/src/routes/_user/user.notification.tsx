import { NotificationPreferencesForm } from '@/features/notifications';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_user/user/notification')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="max-w-xl">
      <h1 className="font-heading text-xl font-semibold tracking-tight mb-1">
        Notifications
      </h1>
      <p className="text-muted-foreground text-sm mb-6">
        Choose what you hear about, and when.
      </p>
      <NotificationPreferencesForm />
    </div>
  );
}
