import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_group/g/$id/settings')({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/_app/_group/group/$id/settings"!</div>;
}
