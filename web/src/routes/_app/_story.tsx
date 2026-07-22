import Wrapper from '@/components/wrapper';
import { createFileRoute, Outlet } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_story')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <Wrapper>
      <Outlet />
    </Wrapper>
  );
}
