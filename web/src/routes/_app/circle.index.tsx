import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/circle/')({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/_app/group/"!</div>;
}
