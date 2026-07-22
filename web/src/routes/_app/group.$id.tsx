import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/group/$id')({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/_app/group/$id"!</div>;
}
