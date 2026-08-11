import Wrapper from '@/components/wrapper';
import { NotificationList } from '@/features/notifications';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/notifications')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <Wrapper>
      <h1 className="font-heading mb-6 text-2xl font-semibold tracking-tight">
        Notifications
      </h1>
      <NotificationList />
    </Wrapper>
  );
}
