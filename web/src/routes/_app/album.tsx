import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Item, ItemHeader } from '@/components/ui/item';
import Wrapper from '@/components/wrapper';
import { ArrowDown01Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/album')({
  component: RouteComponent,
});

const album = [
  'https://images.unsplash.com/photo-1650804068570-7fb2e3dbf888?q=80&w=640&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1650804068570-7fb2e3dbf888?q=80&w=640&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1650804068570-7fb2e3dbf888?q=80&w=640&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1650804068570-7fb2e3dbf888?q=80&w=640&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1650804068570-7fb2e3dbf888?q=80&w=640&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1650804068570-7fb2e3dbf888?q=80&w=640&auto=format&fit=crop',
];

function RouteComponent() {
  return (
    <Wrapper>
      <div className="flex flex-col gap-8">
        <div className="flex items-center">
          <h1 className="text-2xl font-semibold">Album</h1>
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
        <div className="grid grid-cols-3 gap-4">
          {album.map((item) => (
            <img
              src={item}
              width={128}
              height={128}
              className="aspect-video w-full rounded-sm object-cover"
            />
          ))}
        </div>
      </div>
    </Wrapper>
  );
}
