import { Badge } from '@/components/ui/badge';
import { CircleNameLabel } from '@/features/circles';
import {
  getGetMomentsIdSharesQueryKey,
  useDeleteMomentsIdShareCircleId,
  useGetMomentsIdShares,
} from '@/lib/api/generated/mentions/mentions';
import { queryClient } from '@/lib/query-client';

interface ShareToCirclePickerProps {
  momentId: string;
}

// The persistent "manage sharing" surface — which Circles this
// personal Moment is currently shared to, each with an unshare button.
// There's no standalone "share" action here on purpose (FEATURES.md,
// Mention: "Why there is no 'share to circle' button") — sharing only
// ever starts as an offer right after mentioning someone
// (AddMentionForm); this just manages what already exists.
export function ShareToCirclePicker({ momentId }: ShareToCirclePickerProps) {
  const sharesQuery = useGetMomentsIdShares(momentId);
  const unshare = useDeleteMomentsIdShareCircleId();

  const circleIds = sharesQuery.data?.circle_ids ?? [];

  const handleUnshare = async (circleId: string) => {
    await unshare.mutateAsync({ id: momentId, circleId });
    await queryClient.invalidateQueries({
      queryKey: getGetMomentsIdSharesQueryKey(momentId),
    });
  };

  if (circleIds.length === 0) return null;

  return (
    <div className="flex flex-col gap-2">
      <p className="text-sm text-muted-foreground">Shared with</p>
      <div className="flex flex-wrap gap-2">
        {circleIds.map((circleId) => (
          <Badge key={circleId} variant="secondary" className="gap-1">
            <CircleNameLabel circleId={circleId} />
            <button
              type="button"
              onClick={() => handleUnshare(circleId)}
              disabled={unshare.isPending}
              className="text-muted-foreground hover:text-foreground"
              aria-label="Unshare from this circle"
            >
              ×
            </button>
          </Badge>
        ))}
      </div>
    </div>
  );
}
