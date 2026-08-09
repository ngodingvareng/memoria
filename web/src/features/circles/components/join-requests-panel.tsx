import { Button } from '@/components/ui/button';
import {
  ItemGroup,
  Item,
  ItemActions,
  ItemContent,
  ItemSeparator,
} from '@/components/ui/item';
import { CircleUserChip } from './circle-user-chip';
import {
  getGetCirclesIdJoinRequestsQueryKey,
  useGetCirclesIdJoinRequests,
  usePostCirclesIdJoinRequestsRequestIdApprove,
  usePostCirclesIdJoinRequestsRequestIdReject,
} from '@/lib/api/generated/circle-join-requests/circle-join-requests';
import { getGetCirclesIdMembersQueryKey } from '@/lib/api/generated/circles/circles';
import { queryClient } from '@/lib/query-client';

interface JoinRequestsPanelProps {
  circleId: string;
}

export function JoinRequestsPanel({ circleId }: JoinRequestsPanelProps) {
  const joinRequestsQuery = useGetCirclesIdJoinRequests(circleId);
  const approve = usePostCirclesIdJoinRequestsRequestIdApprove();
  const reject = usePostCirclesIdJoinRequestsRequestIdReject();

  const invalidate = () =>
    Promise.all([
      queryClient.invalidateQueries({
        queryKey: getGetCirclesIdJoinRequestsQueryKey(circleId),
      }),
      queryClient.invalidateQueries({
        queryKey: getGetCirclesIdMembersQueryKey(circleId),
      }),
    ]);

  const requests = joinRequestsQuery.data?.join_requests ?? [];

  if (joinRequestsQuery.isPending) {
    return <p className="text-muted-foreground">Loading join requests…</p>;
  }

  if (requests.length === 0) {
    return <p className="text-muted-foreground">No pending join requests.</p>;
  }

  return (
    <ItemGroup>
      {requests.map((request, index) => (
        <div key={request.id}>
          <Item size="sm">
            <ItemContent>
              <CircleUserChip userId={request.user_id!} />
            </ItemContent>
            <ItemActions>
              <Button
                size="sm"
                variant="outline"
                onClick={async () => {
                  await reject.mutateAsync({
                    id: circleId,
                    requestId: request.id!,
                  });
                  await invalidate();
                }}
              >
                Reject
              </Button>
              <Button
                size="sm"
                onClick={async () => {
                  await approve.mutateAsync({
                    id: circleId,
                    requestId: request.id!,
                  });
                  await invalidate();
                }}
              >
                Approve
              </Button>
            </ItemActions>
          </Item>
          {index !== requests.length - 1 && <ItemSeparator />}
        </div>
      ))}
    </ItemGroup>
  );
}
