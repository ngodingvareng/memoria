import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Skeleton } from '@/components/ui/skeleton';
import Wrapper from '@/components/wrapper';
import { MomentHeatmap } from '@/features/moments';
import {
  DeleteThreadDialog,
  EditThreadDetailsSection,
  EditThreadHero,
  EditThreadNameDialog,
} from '@/features/threads';
import {
  useGetThreadsId,
  useGetThreadsIdImages,
} from '@/lib/api/generated/threads/threads';
import { ArrowDown01Icon, ArrowLeft02Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute, Link } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/thread/$id/manage')({
  component: RouteComponent,
});

function RouteComponent() {
  const { id } = Route.useParams();
  const threadQuery = useGetThreadsId(id);
  const imagesQuery = useGetThreadsIdImages(id);
  const thread = threadQuery.data;

  return (
    <Wrapper>
      <div className="-ml-4 pb-4">
        <Button
          variant="ghost"
          size="lg"
          className="[&_svg]:text-muted-foreground! text-xl [&_svg]:size-6!"
          render={<Link to="/thread/$id" params={{ id }} />}
        >
          <HugeiconsIcon icon={ArrowLeft02Icon} strokeWidth={2} /> Back
        </Button>
      </div>

      {threadQuery.isPending && (
        <div className="flex flex-col gap-4">
          <Skeleton className="h-40 w-full rounded-4xl" />
          <Skeleton className="h-8 w-64" />
        </div>
      )}

      {threadQuery.isError && (
        <Alert variant="destructive">
          <AlertDescription>
            Couldn't load this thread. Please try again.
          </AlertDescription>
        </Alert>
      )}

      {thread && (
        <div className="flex flex-col gap-12">
          <div className="flex flex-col gap-4">
            <EditThreadHero
              threadId={id}
              threadName={thread.name ?? ''}
              heroImage={imagesQuery.data?.[0]}
              colorHex={thread.color_hex}
            />
            <div className="flex items-center gap-1">
              <h1 className="text-2xl font-semibold">{thread.name}</h1>
              <EditThreadNameDialog threadId={id} thread={thread} />
            </div>
          </div>

          <div className="flex flex-col gap-4">
            <div className="flex items-center">
              <h2 className="text-xl font-semibold">
                336k moments over the last year
              </h2>

              <div className="flex grow justify-end">
                <DropdownMenu>
                  <DropdownMenuTrigger render={<Button variant="secondary" />}>
                    <span className="text-muted-foreground font-semibold">
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

          <EditThreadDetailsSection threadId={id} thread={thread} />

          <div className="flex flex-col gap-4">
            <h2 className="text-xl font-semibold">Danger Zone</h2>
            <DeleteThreadDialog threadId={id} threadName={thread.name ?? ''} />
          </div>
        </div>
      )}
    </Wrapper>
  );
}
