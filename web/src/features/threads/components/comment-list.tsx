import { useGetMomentsIdComments } from '@/lib/api/generated/comments/comments';
import { CommentRow } from './comment-row';

interface CommentListProps {
  momentId: string;
  // Used to skip the "Mentioned" badge for the Moment owner's own
  // mention-context comment (FEATURES.md, Response Terms) — their
  // authorship needs no explaining.
  momentOwnerId?: string;
}

export function CommentList({ momentId, momentOwnerId }: CommentListProps) {
  const commentsQuery = useGetMomentsIdComments(momentId);
  const comments = commentsQuery.data?.comments ?? [];

  if (commentsQuery.isPending) {
    return <p className="text-muted-foreground">Loading comments…</p>;
  }

  if (comments.length === 0) {
    return <p className="text-muted-foreground">No comments yet.</p>;
  }

  return (
    <div className="flex flex-col gap-4">
      {comments.map((comment) => (
        <CommentRow
          key={comment.id}
          momentId={momentId}
          comment={comment}
          momentOwnerId={momentOwnerId}
        />
      ))}
    </div>
  );
}
