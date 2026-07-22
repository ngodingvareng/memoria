import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/activity/$id/schedule')({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/activities/$id/schedules"!</div>;
}
