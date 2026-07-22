import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/activities/$id/schedules')({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/activities/$id/schedules"!</div>;
}
