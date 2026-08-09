import { Button } from '@/components/ui/button';
import { AddMentionForm, MentionList } from '@/features/mentions';
import {
  CommentInput,
  CommentList,
  ShareToCirclePicker,
} from '@/features/threads';
import { MomentFeedItem } from './moment-feed-item';
import { useMomentAudience } from '../lib/use-moment-audience';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMomentResponse } from '@/lib/api/generated/models';
import { usePostMomentsIdMentionsLeave } from '@/lib/api/generated/mentions/mentions';
import { useGetThreadsId } from '@/lib/api/generated/threads/threads';
import { useSession } from '@/lib/session';
import { useNavigate } from '@tanstack/react-router';
import { useState } from 'react';

interface MomentDetailProps {
  moment: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMomentResponse;
}

export function MomentDetail({ moment }: MomentDetailProps) {
  const session = useSession();
  const threadQuery = useGetThreadsId(moment.thread_id ?? '', {
    query: { enabled: !!moment.thread_id },
  });

  // Mentioning only applies to personal Moments (FEATURES.md, Mention)
  // — one already inside a collaborative Thread is shared to its
  // Circle by construction.
  const isPersonal = !moment.thread_id || !threadQuery.data?.circle_id;
  const isOwner = moment.user_id === session?.user.id;

  return (
    <div className="flex flex-col gap-6">
      <MomentFeedItem moment={moment} />
      <div className="flex flex-col gap-4">
        <CommentInput momentId={moment.id!} />
        <CommentList momentId={moment.id!} momentOwnerId={moment.user_id} />
      </div>

      {isPersonal && (
        <div className="flex flex-col gap-4">
          <h2 className="text-lg font-semibold">Mentions</h2>
          {isOwner ? (
            <>
              <MentionList momentId={moment.id!} />
              <AddMentionForm momentId={moment.id!} />
              <ShareToCirclePicker momentId={moment.id!} />
            </>
          ) : (
            <LeaveMentionSection momentId={moment.id!} />
          )}
        </div>
      )}
    </div>
  );
}

// Shown to a non-owner viewer — reuses the audience resolver (already
// built for Comment/Reaction) rather than a separate check: if the
// mention context is available to this viewer and they're not the
// owner, they must be the mentioned user (FEATURES.md, Leaving a
// mention).
function LeaveMentionSection({ momentId }: { momentId: string }) {
  const navigate = useNavigate();
  const audience = useMomentAudience(momentId);
  const leaveMention = usePostMomentsIdMentionsLeave();
  const [isLeaving, setIsLeaving] = useState(false);

  const canLeave = audience.contexts.some(
    (context) => context.type === 'mention'
  );
  if (!canLeave) return null;

  const handleLeave = async () => {
    setIsLeaving(true);
    try {
      await leaveMention.mutateAsync({ id: momentId });
      navigate({ to: '/mentions' });
    } finally {
      setIsLeaving(false);
    }
  };

  return (
    <div className="flex items-center gap-2">
      <p className="text-sm text-muted-foreground">
        You were mentioned in this moment.
      </p>
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={handleLeave}
        disabled={isLeaving}
      >
        Leave mention
      </Button>
    </div>
  );
}
