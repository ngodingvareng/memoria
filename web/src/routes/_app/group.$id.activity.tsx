import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/group/$id/activity')({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/_app/group/$id/activity"!</div>;
}
