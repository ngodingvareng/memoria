import { Button } from '@/components/ui/button';
import { CircleNameLabel } from '@/features/circles';
import {
  getGetMomentsIdSharesQueryKey,
  postMomentsIdShare,
} from '@/lib/api/generated/mentions/mentions';
import { queryClient } from '@/lib/query-client';
import { useState } from 'react';
import { toast } from 'sonner';
import type { MentionShareOffer } from '../lib/sync-mentions';

interface MentionShareOfferToastProps {
  momentId: string;
  offer: MentionShareOffer;
  toastId: string | number;
}

// Relocated from the old AddMentionForm's inline "Share to circle too?"
// step (FEATURES.md, Mention) — mentioning now happens by editing note
// text rather than through a form with room for this offer inline, so
// it surfaces as a toast right after the mention is created instead.
export function MentionShareOfferToast({
  momentId,
  offer,
  toastId,
}: MentionShareOfferToastProps) {
  const [circleIds, setCircleIds] = useState(offer.circleIds);
  const [pendingCircleId, setPendingCircleId] = useState<string | null>(null);

  const handleShare = async (circleId: string) => {
    setPendingCircleId(circleId);
    try {
      await postMomentsIdShare(momentId, { circle_id: circleId });
      await queryClient.invalidateQueries({
        queryKey: getGetMomentsIdSharesQueryKey(momentId),
      });
      const remaining = circleIds.filter((id) => id !== circleId);
      setCircleIds(remaining);
      if (remaining.length === 0) toast.dismiss(toastId);
    } finally {
      setPendingCircleId(null);
    }
  };

  return (
    <div className="bg-popover text-popover-foreground ring-foreground/5 dark:ring-foreground/10 flex w-full flex-col gap-2 rounded-2xl p-3 text-sm shadow-lg ring-1">
      <p>
        You and {offer.displayName} are both in these circles — share this
        moment there too?
      </p>
      <div className="flex flex-wrap gap-2">
        {circleIds.map((circleId) => (
          <Button
            key={circleId}
            type="button"
            size="sm"
            variant="outline"
            disabled={pendingCircleId === circleId}
            onClick={() => handleShare(circleId)}
          >
            <CircleNameLabel circleId={circleId} />
          </Button>
        ))}
      </div>
    </div>
  );
}
