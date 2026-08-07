import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import Wrapper from '@/components/wrapper';
import { MomentHeatmap, ThreadHero } from '@/features/threads';
import { ArrowDown01Icon, ArrowLeft02Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/thread/$id/info')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <Wrapper>
      <div className="pb-4">
        <Button
          variant="ghost"
          size="lg"
          className="text-xl [&_svg]:size-6! [&_svg]:text-muted-foreground!"
        >
          <HugeiconsIcon icon={ArrowLeft02Icon} strokeWidth={2} /> Back
        </Button>
      </div>
      <div className="flex flex-col gap-12">
        <div className="flex flex-col gap-4">
          <ThreadHero
            isReadMode={false}
            imageUrl="https://images.unsplash.com/photo-1604076850742-4c7221f3101b?q=80&w=1887&auto=format&fit=crop&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D"
            imageAlt="Photo by mymind on Unsplash"
          />
          <h1 className="text-2xl font-semibold">Adalah Pokoknya</h1>
        </div>

        <div className="flex flex-col gap-4">
          <div className="flex items-center">
            <h2 className="text-xl font-semibold">
              336k threads over the last year
            </h2>

            <div className="grow flex justify-end">
              <DropdownMenu>
                <DropdownMenuTrigger render={<Button variant="secondary" />}>
                  <span className="font-semibold text-muted-foreground">
                    Year
                  </span>{' '}
                  2077
                  <HugeiconsIcon icon={ArrowDown01Icon} strokeWidth={2} />
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuGroup>
                    <DropdownMenuItem>2077</DropdownMenuItem>
                    <DropdownMenuItem>2078</DropdownMenuItem>
                    <DropdownMenuItem>2079</DropdownMenuItem>
                  </DropdownMenuGroup>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>

          <MomentHeatmap />
        </div>

        <div className="flex flex-col gap-6">
          <h2 className="text-xl font-semibold">Commitments</h2>
        </div>

        <div className="flex flex-col gap-6">
          <h2 className="text-xl font-semibold">Members</h2>
        </div>

        <div className="flex flex-col gap-6">
          <h2 className="text-xl font-semibold">Settings</h2>
        </div>
      </div>
    </Wrapper>
  );
}
