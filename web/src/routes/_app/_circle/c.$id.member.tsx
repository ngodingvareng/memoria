import { Button } from '@/components/ui/button';
import { CircleMemberList } from '@/features/circles';
import { useGetCirclesIdMembers } from '@/lib/api/generated/circles/circles';
import { useSession } from '@/lib/session';
import { createFileRoute, Link } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_circle/c/$id/member')({
  component: RouteComponent,
});

function RouteComponent() {
  const { id } = Route.useParams();
  const session = useSession();
  const membersQuery = useGetCirclesIdMembers(id);

  const viewerUserId = session?.user.id ?? '';
  const viewerIsAdmin =
    membersQuery.data?.members?.some(
      (m) => m.user_id === viewerUserId && m.role === 'admin'
    ) ?? false;

  return (
    <div className="flex flex-col gap-8">
      <div className="flex items-center">
        <h1 className="text-2xl font-semibold">Members</h1>
        <div className="flex grow items-center justify-end gap-2">
          <Button render={<Link to="/c/$id/invite" params={{ id }} />}>
            Add member
          </Button>
        </div>
      </div>
      <CircleMemberList
        circleId={id}
        viewerUserId={viewerUserId}
        viewerIsAdmin={viewerIsAdmin}
      />
    </div>
  );
}
