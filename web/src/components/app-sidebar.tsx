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
import {
  Activity01Icon,
  Album01Icon,
  BookOpen01Icon,
  Home05Icon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link, useRouterState } from '@tanstack/react-router';
import { Avatar, AvatarFallback, AvatarImage } from './ui/avatar';

// This is sample data.
const data = {
  main: [
    {
      title: 'Home',
      icon: Home05Icon,
      url: '/',
    },
    {
      title: 'Stories',
      icon: BookOpen01Icon,
      url: '/story',
    },
    {
      title: 'Activities',
      icon: Activity01Icon,
      url: '/activity',
    },
    {
      title: 'Album',
      icon: Album01Icon,
      url: '/album',
    },
  ],

  group: [
    {
      name: 'NgodingVareng',
      imageSrc: 'https://github.com/shadcn.png',
      imageAlt: 'hello',
      url: '/activity/1',
    },
  ],

  followed: [
    {
      name: 'ShadCN',
      imageSrc: 'https://github.com/shadcn.png',
      imageAlt: 'hello',
      url: '/@shadcn',
    },
    {
      name: 'Who Are You',
      imageSrc: 'https://github.com/shadcn.png',
      imageAlt: 'hello',
      url: '/@shadcn',
    },
  ],

  recent: [
    {
      title: 'Apa si ini?',
      url: '/activity/1',
    },
    {
      title: 'He is not here anymore, so I should go tomorrow',
      url: '/activity/2',
    },
    {
      title: 'What',
      url: '/activity/3',
    },
    {
      title: 'PHOBOS <- Holy gd reference',
      url: '/activity/4',
    },
    {
      title: "There's weird light in front of my house",
      url: '/activity/5',
    },
  ],
};

export default function AppSidebar({
  ...props
}: React.ComponentProps<typeof Sidebar>) {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

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
          <SidebarGroupLabel>Group</SidebarGroupLabel>
          <SidebarMenu>
            {data.group.map((item) => (
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

        <SidebarGroup className="group-data-[collapsible=icon]:hidden">
          <SidebarGroupLabel>Followed</SidebarGroupLabel>
          <SidebarMenu>
            {data.followed.map((item) => (
              <SidebarMenuItem key={item.name}>
                <SidebarMenuButton
                  render={
                    <Link to={item.url}>
                      <Avatar size="sm">
                        <AvatarImage
                          src={item.imageSrc}
                          alt={item.imageAlt}
                          className="grayscale"
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

        <SidebarGroup className="group-data-[collapsible=icon]:hidden">
          <SidebarGroupLabel>Recent Activities</SidebarGroupLabel>
          <SidebarMenu>
            {data.recent.map((item) => (
              <SidebarMenuItem key={item.title}>
                <SidebarMenuButton
                  render={
                    <Link to={item.url}>
                      <span className="min-w-0 truncate">{item.title}</span>
                    </Link>
                  }
                />
              </SidebarMenuItem>
            ))}
          </SidebarMenu>
        </SidebarGroup>
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
