import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { CircleNameLabel } from '@/features/circles';
import { ApiError } from '@/lib/api-client';
import {
  getGetMomentsIdMentionsQueryKey,
  getGetMomentsIdSharesQueryKey,
  usePostMomentsIdMentions,
  usePostMomentsIdShare,
} from '@/lib/api/generated/mentions/mentions';
import { queryClient } from '@/lib/query-client';
import { useState } from 'react';

interface AddMentionFormProps {
  momentId: string;
}

// Owner-only. Naming someone present on a personal Moment
// automatically shares it with them (FEATURES.md, Mention) — if that
// also puts the owner and the mentioned user in the same Circle(s),
// the response tells us, and we offer sharing there too. There is no
// standalone "share to circle" button; this offer is the only place
// sharing ever starts.
export function AddMentionForm({ momentId }: AddMentionFormProps) {
  const [username, setUsername] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [offer, setOffer] = useState<{
    displayName: string;
    circleIds: string[];
  } | null>(null);
  const createMention = usePostMomentsIdMentions();
  const shareToCircle = usePostMomentsIdShare();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username.trim()) return;

    setError(null);
    setOffer(null);
    try {
      const mention = await createMention.mutateAsync({
        id: momentId,
        data: { username: username.trim() },
      });
      setUsername('');
      await queryClient.invalidateQueries({
        queryKey: getGetMomentsIdMentionsQueryKey(momentId),
      });
      if (mention.shared_circle_ids && mention.shared_circle_ids.length > 0) {
        setOffer({
          displayName: mention.display_name ?? username,
          circleIds: mention.shared_circle_ids,
        });
      }
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Failed to mention this user. Please try again.'
      );
    }
  };

  const handleShare = async (circleId: string) => {
    await shareToCircle.mutateAsync({
      id: momentId,
      data: { circle_id: circleId },
    });
    await queryClient.invalidateQueries({
      queryKey: getGetMomentsIdSharesQueryKey(momentId),
    });
    setOffer((prev) =>
      prev
        ? { ...prev, circleIds: prev.circleIds.filter((id) => id !== circleId) }
        : prev
    );
  };

  return (
    <div className="flex flex-col gap-3">
      <form onSubmit={handleSubmit} className="flex gap-2">
        <Input
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          placeholder="username"
          autoComplete="off"
        />
        <Button
          type="submit"
          disabled={!username.trim() || createMention.isPending}
        >
          Mention
        </Button>
      </form>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {offer && offer.circleIds.length > 0 && (
        <div className="flex flex-col gap-2 rounded-lg border p-3">
          <p className="text-sm text-muted-foreground">
            You and {offer.displayName} are both in these circles — share this
            moment there too?
          </p>
          <div className="flex flex-wrap gap-2">
            {offer.circleIds.map((circleId) => (
              <Button
                key={circleId}
                type="button"
                size="sm"
                variant="outline"
                onClick={() => handleShare(circleId)}
                disabled={shareToCircle.isPending}
              >
                <CircleNameLabel circleId={circleId} />
              </Button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
