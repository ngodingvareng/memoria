import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/album')({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/_app/album"!</div>;
}
