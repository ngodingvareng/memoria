import Wrapper from '@/components/wrapper';
import { createFileRoute, Outlet } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_stories')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <Wrapper>
      <Outlet />
    </Wrapper>
  );
}
