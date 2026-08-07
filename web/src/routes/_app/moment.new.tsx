import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/moment/new')({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/_app/moment/new"!</div>;
}
