import { Button } from '@/components/ui/button';
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemSeparator,
  ItemTitle,
} from '@/components/ui/item';
import {
  getGetCircleInvitesQueryKey,
  useGetCircleInvites,
  usePostCircleInvitesIdAccept,
  usePostCircleInvitesIdDecline,
} from '@/lib/api/generated/circle-invites/circle-invites';
import { getGetCirclesQueryKey } from '@/lib/api/generated/circles/circles';
import { queryClient } from '@/lib/query-client';

export function PendingCircleInvites() {
  const invitesQuery = useGetCircleInvites();
  const accept = usePostCircleInvitesIdAccept();
  const decline = usePostCircleInvitesIdDecline();

  const invalidate = () =>
    Promise.all([
      queryClient.invalidateQueries({
        queryKey: getGetCircleInvitesQueryKey(),
      }),
      queryClient.invalidateQueries({ queryKey: getGetCirclesQueryKey() }),
    ]);

  const invites = invitesQuery.data?.invites ?? [];
  if (invitesQuery.isPending || invites.length === 0) return null;

  return (
    <div className="flex flex-col gap-2">
      <h2 className="text-xl font-semibold">Circle invites</h2>
      <ItemGroup>
        {invites.map((invite, index) => (
          <div key={invite.id}>
            <Item size="sm">
              <ItemContent>
                {/* CircleInviteResponse only carries circle_id — GET
                    /circles/:id requires active membership, which a
                    pending invitee doesn't have yet, so the circle's
                    name can't be resolved here. */}
                <ItemTitle>You've been invited to a circle</ItemTitle>
                <ItemDescription>
                  {invite.invited_by_user_id ? 'by a member' : ''}
                </ItemDescription>
              </ItemContent>
              <ItemActions>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={async () => {
                    await decline.mutateAsync({ id: invite.id! });
                    await invalidate();
                  }}
                >
                  Decline
                </Button>
                <Button
                  size="sm"
                  onClick={async () => {
                    await accept.mutateAsync({ id: invite.id! });
                    await invalidate();
                  }}
                >
                  Accept
                </Button>
              </ItemActions>
            </Item>
            {index !== invites.length - 1 && <ItemSeparator />}
          </div>
        ))}
      </ItemGroup>
    </div>
  );
}
