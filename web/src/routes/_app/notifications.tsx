import { NotificationList } from '@/features/notifications';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/notifications')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="mx-auto max-w-2xl px-4 py-6">
      <h1 className="font-heading mb-6 text-2xl font-semibold tracking-tight">
        Notifications
      </h1>
      <NotificationList />
    </div>
  );
}
