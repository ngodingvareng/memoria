import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_circle/c/$id/invite')({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/_app/_circle/c/$id/invite"!</div>;
}
