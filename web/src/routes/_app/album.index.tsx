import { Button } from '@/components/ui/button';
import { CambioImage } from '@/components/ui/cambio-image';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import Wrapper from '@/components/wrapper';
import { AlbumTimeline } from '@/features/album';
import { ArrowDown01Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/album/')({
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
          <div className="flex grow items-center justify-end gap-2">
            <DropdownMenu>
              <DropdownMenuTrigger render={<Button variant="secondary" />}>
                <span className="text-muted-foreground font-semibold">
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
          </div>
        </div>
        <div>
          <AlbumTimeline
            albums={[
              {
                date: 'November 7, 2029',
                content: (
                  <div className="grid grid-cols-3 gap-x-1.5">
                    {album.map((item) => (
                      <CambioImage
                        key={item}
                        src="https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=800&h=600&fit=crop"
                        alt="Beautiful mountain landscape"
                        width={800}
                        height={600}
                        motion="smooth"
                        className="rounded-md"
                      />
                    ))}
                  </div>
                ),
              },
              {
                date: 'November 4, 2029',
                content: (
                  <div className="grid grid-cols-3 gap-x-1.5">
                    {album.map((item) => (
                      <CambioImage
                        key={item}
                        src="https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=800&h=600&fit=crop"
                        alt="Beautiful mountain landscape"
                        width={800}
                        height={600}
                        motion="smooth"
                        className="rounded-md"
                      />
                    ))}
                  </div>
                ),
              },
            ]}
          />
        </div>
      </div>
    </Wrapper>
  );
}
