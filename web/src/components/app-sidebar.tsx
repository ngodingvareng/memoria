import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from '@/components/ui/sidebar';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetThreads } from '@/lib/api/generated/threads/threads';
import {
  Activity01Icon,
  Album01Icon,
  AtIcon,
  CircleIcon,
  Home05Icon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link, useRouterState } from '@tanstack/react-router';
import { Avatar, AvatarFallback, AvatarImage } from './ui/avatar';

// How many threads to surface in the sidebar's "Recent Threads" section.
const RECENT_THREADS_LIMIT = 5;

// This is sample data.
const data = {
  main: [
    {
      title: 'Home',
      icon: Home05Icon,
      url: '/',
    },
    {
      title: 'Threads',
      icon: Activity01Icon,
      url: '/thread',
    },
    {
      title: 'Circles',
      icon: CircleIcon,
      url: '/circle',
    },
    {
      title: 'Mentions',
      icon: AtIcon,
      url: '/mentions',
    },
    {
      title: 'Album',
      icon: Album01Icon,
      url: '/album',
    },
  ],

  circle: [
    {
      name: 'NgodingVareng',
      imageSrc: 'https://github.com/shadcn.png',
      imageAlt: 'hello',
      url: '/g/1',
    },
  ],
};

export default function AppSidebar({
  ...props
}: React.ComponentProps<typeof Sidebar>) {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });
  const { data: threadsData, isPending: isThreadsPending } = useGetThreads({
    page_size: RECENT_THREADS_LIMIT,
  });
  const recentThreads = threadsData?.threads ?? [];

  return (
    <Sidebar
      className="top-(--header-height) h-[calc(100svh-var(--header-height))]!"
      {...props}
    >
      <SidebarContent>
        <SidebarGroup>
          <SidebarMenu>
            {data.main.map((item) => (
              <SidebarMenuItem key={item.title}>
                <SidebarMenuButton
                  size="lg"
                  className="text-lg font-medium [&_svg]:size-6"
                  isActive={pathname === item.url}
                  render={
                    <Link to={item.url}>
                      <HugeiconsIcon icon={item.icon} strokeWidth={2} />
                      {item.title}
                    </Link>
                  }
                />
              </SidebarMenuItem>
            ))}
          </SidebarMenu>
        </SidebarGroup>

        <SidebarGroup className="group-data-[collapsible=icon]:hidden">
          <SidebarGroupLabel>Circles</SidebarGroupLabel>
          <SidebarMenu>
            {data.circle.map((item) => (
              <SidebarMenuItem key={item.name}>
                <SidebarMenuButton
                  render={
                    <Link to={item.url}>
                      <Avatar size="sm" className="rounded-sm after:rounded-sm">
                        <AvatarImage
                          src={item.imageSrc}
                          alt={item.imageAlt}
                          className="rounded-sm"
                        />
                        <AvatarFallback>CN</AvatarFallback>
                      </Avatar>
                      <span className="min-w-0 truncate">{item.name}</span>
                    </Link>
                  }
                />
              </SidebarMenuItem>
            ))}
          </SidebarMenu>
        </SidebarGroup>

        {(isThreadsPending || recentThreads.length > 0) && (
          <SidebarGroup className="group-data-[collapsible=icon]:hidden">
            <SidebarGroupLabel>Recent Threads</SidebarGroupLabel>
            <SidebarMenu>
              {isThreadsPending
                ? Array.from({ length: RECENT_THREADS_LIMIT }).map((_, i) => (
                    <SidebarMenuItem key={i}>
                      <Skeleton className="h-8 w-full" />
                    </SidebarMenuItem>
                  ))
                : recentThreads.map((thread) => (
                    <SidebarMenuItem key={thread.id}>
                      <SidebarMenuButton
                        isActive={pathname === `/thread/${thread.id}`}
                        render={
                          <Link to="/thread/$id" params={{ id: thread.id! }}>
                            <span className="min-w-0 truncate">
                              {thread.name}
                            </span>
                          </Link>
                        }
                      />
                    </SidebarMenuItem>
                  ))}
            </SidebarMenu>
          </SidebarGroup>
        )}
      </SidebarContent>
      <SidebarFooter>
        <div className="text-xs flex flex-col gap-2 px-2 py-1">
          <div className="flex items-center flex-wrap gap-2">
            <Link to="/about" className="hover:underline">
              About
            </Link>
            <Link to="/privacy" className="hover:underline">
              Privacy
            </Link>
            <Link to="/terms" className="hover:underline">
              Terms
            </Link>
          </div>
          <p className="text-muted-foreground">&copy;2026 Memoria</p>
        </div>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
