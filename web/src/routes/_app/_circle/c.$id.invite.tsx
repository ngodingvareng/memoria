import { Button } from '@/components/ui/button';
import { InviteDirectForm, InviteLinkPanel } from '@/features/circles';
import { useGetCirclesIdMembers } from '@/lib/api/generated/circles/circles';
import { useSession } from '@/lib/session';
import { ArrowLeft02Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute, Link } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_circle/c/$id/invite')({
  component: RouteComponent,
});

function RouteComponent() {
  const { id } = Route.useParams();
  const session = useSession();
  const membersQuery = useGetCirclesIdMembers(id);

  const viewerUserId = session?.user.id ?? '';
  const viewer = membersQuery.data?.members?.find(
    (m) => m.user_id === viewerUserId
  );
  const canInvite = viewer?.role === 'admin' || viewer?.can_invite === true;

  return (
    <div className="flex flex-col gap-8 max-w-xl">
      <div className="-ml-4">
        <Button
          variant="ghost"
          className="text-muted-foreground"
          render={<Link to="/c/$id/member" params={{ id }} />}
        >
          <HugeiconsIcon icon={ArrowLeft02Icon} strokeWidth={2} /> Back
        </Button>
      </div>

      <h1 className="text-2xl font-semibold">Add member</h1>

      {!canInvite ? (
        <p className="text-muted-foreground">
          You don't have permission to invite members to this circle.
        </p>
      ) : (
        <>
          <div className="flex flex-col gap-4">
            <h2 className="text-lg font-semibold">By username</h2>
            <InviteDirectForm circleId={id} />
          </div>

          <div className="flex flex-col gap-4">
            <h2 className="text-lg font-semibold">By link</h2>
            <InviteLinkPanel circleId={id} />
          </div>
        </>
      )}
    </div>
  );
}
