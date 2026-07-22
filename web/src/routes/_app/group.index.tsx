import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/group/')({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/_app/group/"!</div>;
}
