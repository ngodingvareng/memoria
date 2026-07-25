import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import Wrapper from '@/components/wrapper';
import { createFileRoute, Outlet } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_group')({
  component: RouteComponent,
});

function RouteComponent() {
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
              <p className="text-lg/5 text-muted-foreground">#ngodingvareng</p>
            </div>
          </div>
        </div>

        <Tabs defaultValue="overview" className="border-b-2">
          <TabsList variant="line">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="activities">Activities</TabsTrigger>
            <TabsTrigger value="members">Members</TabsTrigger>
            <TabsTrigger value="settings">Settings</TabsTrigger>
          </TabsList>
        </Tabs>
        <Outlet />
      </div>
    </Wrapper>
  );
}
