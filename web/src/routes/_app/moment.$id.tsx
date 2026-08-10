import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import Wrapper from '@/components/wrapper';
import { MomentDetail } from '@/features/moments';
import { useGetMomentsId } from '@/lib/api/generated/moments/moments';
import { ArrowLeft02Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute, useRouter } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/moment/$id')({
  component: RouteComponent,
});

function RouteComponent() {
  const { id } = Route.useParams();
  const router = useRouter();
  const momentQuery = useGetMomentsId(id);

  return (
    <Wrapper>
      <div className="-ml-4 pb-4">
        <Button
          variant="ghost"
          size="lg"
          className="[&_svg]:text-muted-foreground! text-xl [&_svg]:size-6!"
          onClick={() => router.history.back()}
        >
          <HugeiconsIcon icon={ArrowLeft02Icon} strokeWidth={2} /> Back
        </Button>
      </div>

      <h1 className="pb-4 text-2xl font-semibold">Moment Comments</h1>

      {momentQuery.isPending && (
        <div className="flex flex-col gap-4">
          <Skeleton className="h-40 w-full rounded-2xl" />
          <Skeleton className="h-24 w-full" />
        </div>
      )}

      {momentQuery.isError && (
        <Alert variant="destructive">
          <AlertDescription>
            Couldn't find this moment. It may not exist, or you may not have
            access to it.
          </AlertDescription>
        </Alert>
      )}

      {momentQuery.data && <MomentDetail moment={momentQuery.data} />}
    </Wrapper>
  );
}
