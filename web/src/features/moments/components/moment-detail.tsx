import { MentionList } from '@/features/mentions';
import {
  CommentInput,
  CommentList,
  ShareToCirclePicker,
} from '@/features/threads';
import { MomentFeedItem } from './moment-feed-item';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMomentResponse } from '@/lib/api/generated/models';
import { useGetThreadsId } from '@/lib/api/generated/threads/threads';
import { useSession } from '@/lib/session';
import { Separator } from '@/components/ui/separator';

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
      <Separator />
      <div className="flex min-h-[50dvh] flex-col gap-6">
        <div className="flex flex-col gap-4">
          <CommentInput momentId={moment.id!} />
          <CommentList momentId={moment.id!} momentOwnerId={moment.user_id} />
        </div>

        {isPersonal && isOwner && (
          <div className="flex flex-col gap-4">
            <h2 className="text-lg font-semibold">Mentions</h2>
            <MentionList momentId={moment.id!} />
            <ShareToCirclePicker momentId={moment.id!} />
          </div>
        )}
      </div>
    </div>
  );
}
