import { Alert, AlertDescription } from '@/components/ui/alert';
import { Skeleton } from '@/components/ui/skeleton';
import {
  CircleImageUploader,
  DissolveCircleDialog,
  EditCircleDetailsForm,
  InviteLinkPanel,
  JoinRequestsPanel,
} from '@/features/circles';
import {
  useGetCirclesId,
  useGetCirclesIdMembers,
} from '@/lib/api/generated/circles/circles';
import { useSession } from '@/lib/session';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_circle/c/$id/settings')({
  component: RouteComponent,
});

function RouteComponent() {
  const { id } = Route.useParams();
  const session = useSession();
  const circleQuery = useGetCirclesId(id);
  const membersQuery = useGetCirclesIdMembers(id);

  const viewerUserId = session?.user.id ?? '';
  const viewer = membersQuery.data?.members?.find(
    (m) => m.user_id === viewerUserId
  );
  const viewerIsAdmin = viewer?.role === 'admin';
  const canManageInvites = viewerIsAdmin || viewer?.can_invite === true;

  if (circleQuery.isPending) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-40 w-full" />
      </div>
    );
  }

  if (circleQuery.isError || !circleQuery.data) {
    return (
      <Alert variant="destructive">
        <AlertDescription>
          Couldn't load this circle's settings. Please try again.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="flex max-w-xl flex-col gap-12">
      <div className="flex flex-col gap-4">
        <h2 className="text-xl font-semibold">Details</h2>
        {viewerIsAdmin && (
          <CircleImageUploader
            circleId={id}
            imagePath={circleQuery.data.image_path}
            name={circleQuery.data.name}
          />
        )}
        <EditCircleDetailsForm circleId={id} circle={circleQuery.data} />
      </div>

      {canManageInvites && (
        <div className="flex flex-col gap-4">
          <h2 className="text-xl font-semibold">Invite link</h2>
          <InviteLinkPanel circleId={id} />
        </div>
      )}

      {viewerIsAdmin && (
        <div className="flex flex-col gap-4">
          <h2 className="text-xl font-semibold">Join requests</h2>
          <JoinRequestsPanel circleId={id} />
        </div>
      )}

      {viewerIsAdmin && (
        <div className="flex flex-col gap-4">
          <h2 className="text-xl font-semibold">Danger Zone</h2>
          <DissolveCircleDialog
            circleId={id}
            circleName={circleQuery.data.name ?? ''}
          />
        </div>
      )}
    </div>
  );
}
