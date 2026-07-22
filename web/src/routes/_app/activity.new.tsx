import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/activity/new')({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/_app/activity/new"!</div>;
}
