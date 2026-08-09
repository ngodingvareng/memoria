import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import Wrapper from '@/components/wrapper';
import { useGetCirclesId } from '@/lib/api/generated/circles/circles';
import {
  createFileRoute,
  Link,
  Outlet,
  useParams,
  useRouterState,
} from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_circle')({
  component: RouteComponent,
});

function RouteComponent() {
  const { id } = useParams({ strict: false }) as { id: string };
  const circleQuery = useGetCirclesId(id);
  const circle = circleQuery.data;

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
            {circleQuery.isPending ? (
              <>
                <Skeleton className="size-12 rounded-xl" />
                <Skeleton className="h-6 w-40" />
              </>
            ) : (
              <>
                <Avatar size="xl" className="rounded-xl after:rounded-xl">
                  <AvatarImage
                    src={circle?.image_path ?? undefined}
                    alt={circle?.name}
                    className="rounded-xl"
                  />
                  <AvatarFallback>
                    {circle?.name?.charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div>
                  <p className="font-medium text-2xl/5">{circle?.name}</p>
                  {circle?.description && (
                    <p className="text-lg/5 text-muted-foreground">
                      {circle.description}
                    </p>
                  )}
                </div>
              </>
            )}
          </div>
        </div>

        <Tabs value={activeTab} className="border-b-2">
          <TabsList variant="line">
            <TabsTrigger
              value="overview"
              render={
                <Link to="/c/$id" params={{ id }}>
                  Overview
                </Link>
              }
            />
            <TabsTrigger
              value="threads"
              render={
                <Link to="/c/$id/thread" params={{ id }}>
                  Threads
                </Link>
              }
            />
            <TabsTrigger
              value="members"
              render={
                <Link to="/c/$id/member" params={{ id }}>
                  Members
                </Link>
              }
            />
            <TabsTrigger
              value="settings"
              render={
                <Link to="/c/$id/settings" params={{ id }}>
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
