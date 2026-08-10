import {
  REACTION_EMOJI,
  REACTION_KINDS,
} from '@/features/moments/lib/reaction-emoji';
import { useGetMomentsIdReactions } from '@/lib/api/generated/reactions/reactions';
import { ReactorAvatar } from './reactor-avatar';

interface ReactionSummaryProps {
  momentId: string;
}

// "Who reacted is visible, how many is not" (FEATURES.md, Reaction) —
// this renders reactor avatars per emoji, never a number.
export function ReactionSummary({ momentId }: ReactionSummaryProps) {
  const reactionsQuery = useGetMomentsIdReactions(momentId);
  const reactions = reactionsQuery.data?.reactions ?? [];

  const groups = REACTION_KINDS.map((kind) => ({
    kind,
    reactorIds: reactions
      .filter((r) => r.kind === kind && r.user_id)
      .map((r) => r.user_id!),
  })).filter((group) => group.reactorIds.length > 0);

  if (groups.length === 0) return null;

  return (
    <div className="flex items-center gap-3">
      {groups.map(({ kind, reactorIds }) => (
        <div key={kind} className="flex items-center gap-1">
          <span className="text-sm">{REACTION_EMOJI[kind]}</span>
          <div className="flex -space-x-2">
            {reactorIds.slice(0, 4).map((userId) => (
              <ReactorAvatar key={userId} userId={userId} />
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
