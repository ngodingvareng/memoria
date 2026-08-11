import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  REACTION_EMOJI,
  REACTION_KINDS,
  type ReactionKind,
} from '@/features/moments/lib/reaction-emoji';
import {
  toCircleId,
  useMomentAudience,
  type AudienceContext,
} from '@/features/moments/lib/use-moment-audience';
import { getApiErrorMessage } from '@/lib/api-client';
import {
  getGetMomentsIdReactionsQueryKey,
  useDeleteMomentsIdReactions,
  useGetMomentsIdReactions,
  usePutMomentsIdReactions,
} from '@/lib/api/generated/reactions/reactions';
import { queryClient } from '@/lib/query-client';
import { useSession } from '@/lib/session';
import { SmileIcon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { useState } from 'react';
import { AudiencePicker } from './audience-picker';

interface ReactionPickerProps {
  momentId: string;
}

export function ReactionPicker({ momentId }: ReactionPickerProps) {
  const session = useSession();
  const audience = useMomentAudience(momentId);
  const reactionsQuery = useGetMomentsIdReactions(momentId);
  const setReaction = usePutMomentsIdReactions();
  const removeReaction = useDeleteMomentsIdReactions();

  const [selectedContext, setSelectedContext] = useState<
    AudienceContext | undefined
  >(undefined);
  const [error, setError] = useState<string | null>(null);

  const myReaction = reactionsQuery.data?.reactions?.find(
    (r) => r.user_id === session?.user.id
  );
  const activeContext = selectedContext ?? audience.defaultContext;
  const showAudiencePicker = audience.contexts.length > 1;

  const react = async (kind: ReactionKind) => {
    if (!activeContext) return;
    setError(null);
    try {
      if (myReaction?.kind === kind) {
        await removeReaction.mutateAsync({
          id: momentId,
          params: { circle_id: toCircleId(activeContext) },
        });
      } else {
        await setReaction.mutateAsync({
          id: momentId,
          data: { circle_id: toCircleId(activeContext), kind },
        });
      }
      await queryClient.invalidateQueries({
        queryKey: getGetMomentsIdReactionsQueryKey(momentId),
      });
    } catch (err) {
      setError(getApiErrorMessage(err, 'Failed to react. Please try again.'));
    }
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger render={<Button size="icon-sm" variant="ghost" />}>
        <HugeiconsIcon
          strokeWidth={2}
          icon={SmileIcon}
          className="size-5"
          // className={myReaction ? 'fill-yellow-500/50' : undefined}
        />
      </DropdownMenuTrigger>
      <DropdownMenuContent className="flex flex-col gap-2 p-2" align="start">
        {showAudiencePicker && (
          <AudiencePicker
            contexts={audience.contexts}
            value={activeContext}
            onChange={setSelectedContext}
          />
        )}
        <div className="flex gap-1">
          {REACTION_KINDS.map((kind) => (
            <Button
              key={kind}
              type="button"
              size="icon"
              variant={myReaction?.kind === kind ? 'secondary' : 'ghost'}
              disabled={!activeContext}
              onClick={() => react(kind)}
            >
              <span className="text-lg">{REACTION_EMOJI[kind]}</span>
            </Button>
          ))}
        </div>
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
