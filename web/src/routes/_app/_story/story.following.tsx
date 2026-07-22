import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_story/story/following')({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/_app/stories/following"!</div>;
}
