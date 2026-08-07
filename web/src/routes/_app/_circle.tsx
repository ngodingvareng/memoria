import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import Wrapper from '@/components/wrapper';
import {
  createFileRoute,
  Link,
  Outlet,
  useRouterState,
} from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_circle')({
  component: RouteComponent,
});

function RouteComponent() {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

  const activeTab = pathname.includes('/settings')
    ? 'settings'
    : pathname.includes('/member')
      ? 'members'
      : pathname.includes('/thread')
        ? 'threads'
        : 'overview';

  return (
    <Wrapper>
      <div className="flex flex-col gap-6">
        <div className="flex items-center">
          <div className="flex gap-2 items-center">
            <Avatar size="xl" className="rounded-xl after:rounded-xl">
              <AvatarImage
                src="https://github.com/shadcn.png"
                alt="@shadcn"
                className="rounded-xl"
              />
              <AvatarFallback>CN</AvatarFallback>
            </Avatar>
            <div>
              <p className="font-medium text-2xl/5">NgodingVareng</p>
              <p className="text-lg/5 text-muted-foreground">g/ngodingvareng</p>
            </div>
          </div>
        </div>

        <Tabs value={activeTab} className="border-b-2">
          <TabsList variant="line">
            <TabsTrigger
              value="overview"
              render={
                <Link to="/g/$id" params={{ id: '1' }}>
                  Overview
                </Link>
              }
            />
            <TabsTrigger
              value="threads"
              render={
                <Link to="/g/$id/thread" params={{ id: '1' }}>
                  Threads
                </Link>
              }
            />
            <TabsTrigger
              value="members"
              render={
                <Link to="/g/$id/member" params={{ id: '1' }}>
                  Members
                </Link>
              }
            />
            <TabsTrigger
              value="settings"
              render={
                <Link to="/g/$id/settings" params={{ id: '1' }}>
                  Settings
                </Link>
              }
            />
          </TabsList>
        </Tabs>
        <Outlet />
      </div>
    </Wrapper>
  );
}
