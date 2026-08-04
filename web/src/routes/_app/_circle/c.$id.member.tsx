import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemMedia,
  ItemSeparator,
  ItemTitle,
} from '@/components/ui/item';
import {
  ArrowDown01Icon,
  MoreHorizontalIcon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_circle/c/$id/member')({
  component: RouteComponent,
});

const people = [
  {
    username: 'shadcn',
    avatar: 'https://github.com/shadcn.png',
    email: 'shadcn@vercel.com',
  },
  {
    username: 'maxleiter',
    avatar: 'https://github.com/maxleiter.png',
    email: 'maxleiter@vercel.com',
  },
  {
    username: 'evilrabbit',
    avatar: 'https://github.com/evilrabbit.png',
    email: 'evilrabbit@vercel.com',
  },
];

function RouteComponent() {
  return (
    <div className="flex flex-col gap-8">
      <div className="flex items-center">
        <h1 className="text-2xl font-semibold">Members</h1>
        <div className="grow justify-end gap-2 flex items-center">
          <DropdownMenu>
            <DropdownMenuTrigger render={<Button variant="secondary" />}>
              <span className="font-semibold text-muted-foreground">
                Sort by
              </span>{' '}
              Last updated
              <HugeiconsIcon icon={ArrowDown01Icon} strokeWidth={2} />
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuGroup>
                <DropdownMenuItem>Last updated</DropdownMenuItem>
                <DropdownMenuItem>Date created</DropdownMenuItem>
                <DropdownMenuItem>Alphabetical</DropdownMenuItem>
              </DropdownMenuGroup>
            </DropdownMenuContent>
          </DropdownMenu>
          <Button>Add member</Button>
        </div>
      </div>
      <ItemGroup>
        {people.map((person) => (
          <>
            <Item key={person.username} size="xs">
              <ItemMedia>
                <Avatar>
                  <AvatarImage src={person.avatar} />
                  <AvatarFallback>{person.username.charAt(0)}</AvatarFallback>
                </Avatar>
              </ItemMedia>
              <ItemContent className="gap-1">
                <ItemTitle>{person.username}</ItemTitle>
                <ItemDescription>{person.email}</ItemDescription>
              </ItemContent>
              <ItemActions>
                <DropdownMenu>
                  <DropdownMenuTrigger
                    render={<Button variant="ghost" size="icon" />}
                  >
                    <HugeiconsIcon icon={MoreHorizontalIcon} strokeWidth={2} />
                  </DropdownMenuTrigger>
                  <DropdownMenuContent>
                    <DropdownMenuGroup>
                      <DropdownMenuItem>Last updated</DropdownMenuItem>
                      <DropdownMenuItem>Date created</DropdownMenuItem>
                      <DropdownMenuItem>Alphabetical</DropdownMenuItem>
                    </DropdownMenuGroup>
                  </DropdownMenuContent>
                </DropdownMenu>
              </ItemActions>
            </Item>
            {people[people.length - 1].username !== person.username && (
              <ItemSeparator />
            )}
          </>
        ))}
      </ItemGroup>
    </div>
  );
}
