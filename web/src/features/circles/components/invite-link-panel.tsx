import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { ApiError } from '@/lib/api-client';
import {
  getGetCirclesIdInviteLinkQueryKey,
  useGetCirclesIdInviteLink,
  usePatchCirclesIdInviteLinkApproval,
  usePostCirclesIdInviteLink,
} from '@/lib/api/generated/circle-invites/circle-invites';
import { queryClient } from '@/lib/query-client';
import { Copy01Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { useState } from 'react';

interface InviteLinkPanelProps {
  circleId: string;
}

export function InviteLinkPanel({ circleId }: InviteLinkPanelProps) {
  const [freshToken, setFreshToken] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const inviteLinkQuery = useGetCirclesIdInviteLink(circleId);
  const createOrRotate = usePostCirclesIdInviteLink();
  const setApproval = usePatchCirclesIdInviteLinkApproval();

  const invite = inviteLinkQuery.data;
  const requiresApproval = invite?.requires_approval ?? true;

  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: getGetCirclesIdInviteLinkQueryKey(circleId),
    });

  const handleCreateOrRotate = async () => {
    setError(null);
    try {
      const result = await createOrRotate.mutateAsync({
        id: circleId,
        data: { requires_approval: requiresApproval },
      });
      setFreshToken(result.token ?? null);
      await invalidate();
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Failed to create the invite link.'
      );
    }
  };

  const handleApprovalChange = async (next: boolean) => {
    setError(null);
    try {
      await setApproval.mutateAsync({
        id: circleId,
        data: { requires_approval: next },
      });
      await invalidate();
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Failed to update the invite link.'
      );
    }
  };

  const linkUrl = freshToken
    ? `${window.location.origin}/c/${circleId}/join/${freshToken}`
    : null;

  return (
    <div className="flex flex-col gap-3">
      {invite && !inviteLinkQuery.isError && (
        <div className="flex items-center gap-2">
          <Checkbox
            id={`invite-link-approval-${circleId}`}
            checked={requiresApproval}
            disabled={setApproval.isPending}
            onCheckedChange={handleApprovalChange}
          />
          <Label
            htmlFor={`invite-link-approval-${circleId}`}
            className="text-sm text-muted-foreground"
          >
            Require approval to join via link
          </Label>
        </div>
      )}

      {linkUrl ? (
        <div className="flex items-center gap-2">
          <Input readOnly value={linkUrl} className="font-mono text-sm" />
          <Button
            type="button"
            variant="secondary"
            size="icon"
            onClick={() => navigator.clipboard.writeText(linkUrl)}
          >
            <HugeiconsIcon icon={Copy01Icon} strokeWidth={2} />
            <span className="sr-only">Copy link</span>
          </Button>
        </div>
      ) : (
        <p className="text-sm text-muted-foreground">
          {invite && !inviteLinkQuery.isError
            ? "The link's full URL is only ever shown once, right after it's generated. Rotate it to get a fresh copy."
            : 'No invite link yet.'}
        </p>
      )}

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <Button
        type="button"
        variant="outline"
        className="self-start"
        disabled={createOrRotate.isPending}
        onClick={handleCreateOrRotate}
      >
        {invite && !inviteLinkQuery.isError
          ? 'Rotate link'
          : 'Generate invite link'}
      </Button>
    </div>
  );
}
