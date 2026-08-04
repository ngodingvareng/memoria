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
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemHeader,
  ItemTitle,
} from '@/components/ui/item';
import { ArrowDown01Icon, StarIcon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute, Link } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_circle/c/$id/activity')({
  component: RouteComponent,
});

const activities = [
  {
    name: 'v0-1.5-sm',
    description: '5 days ago',
    image:
      'https://images.unsplash.com/photo-1650804068570-7fb2e3dbf888?q=80&w=640&auto=format&fit=crop',
    credit: 'Valeria Reverdo on Unsplash',
    isFavourited: true,
  },
  {
    name: 'v0-1.5-lg',
    description: '3 hours ago - needs confirmation',
    image:
      'https://images.unsplash.com/photo-1610280777472-54133d004c8c?q=80&w=640&auto=format&fit=crop',
    credit: 'Michael Oeser on Unsplash',
    isFavourited: false,
  },
  {
    name: 'v0-2.0-mini',
    description: '8 hour ago',
    image:
      'https://images.unsplash.com/photo-1602146057681-08560aee8cde?q=80&w=640&auto=format&fit=crop',
    credit: 'Cherry Laithang on Unsplash',
    isFavourited: false,
  },
];

function RouteComponent() {
  return (
    <div className="flex flex-col gap-8">
      <div className="flex items-center">
        <h1 className="text-2xl font-semibold">Activities</h1>
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
          <Button>New Activity</Button>
        </div>
      </div>
      <ItemGroup className="grid grid-cols-3 gap-4">
        {activities.map((model) => (
          <Item
            key={model.name}
            variant="outline"
            render={
              <Link to="/activity/$id" params={{ id: '1' }}>
                <ItemHeader>
                  <img
                    src={model.image}
                    alt={model.name}
                    width={128}
                    height={128}
                    className="aspect-video w-full rounded-sm object-cover"
                  />
                </ItemHeader>
                <ItemContent>
                  <ItemTitle>
                    {model.name}{' '}
                    {model.isFavourited && (
                      <HugeiconsIcon
                        icon={StarIcon}
                        size={12}
                        className="fill-muted-foreground text-muted-foreground"
                      />
                    )}
                  </ItemTitle>
                  <ItemDescription>{model.description}</ItemDescription>
                </ItemContent>
              </Link>
            }
          ></Item>
        ))}
      </ItemGroup>
    </div>
  );
}
