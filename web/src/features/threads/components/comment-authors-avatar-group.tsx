import { AvatarGroup, AvatarGroupCount } from '@/components/ui/avatar';
import { useGetMomentsIdComments } from '@/lib/api/generated/comments/comments';
import { CommentAuthorAvatar } from './comment-author-avatar';

// How many distinct commenters' avatars show before collapsing the
// rest into a "+N" count.
const MAX_VISIBLE_COMMENT_AUTHORS = 4;

interface CommentAuthorsAvatarGroupProps {
  momentId: string;
  className?: string;
}

// Who's commented, at a glance, next to the Comment button — dedup'd
// by author, newest comment's author last. Deleted-account authors
// (no user_id) are skipped; they have nothing to show an avatar for.
export function CommentAuthorsAvatarGroup({
  momentId,
  className,
}: CommentAuthorsAvatarGroupProps) {
  const commentsQuery = useGetMomentsIdComments(momentId);
  const comments = commentsQuery.data?.comments ?? [];

  const authorIds = [
    ...new Set(
      comments
        .map((comment) => comment.user_id)
        .filter((userId): userId is string => !!userId)
    ),
  ];

  if (authorIds.length === 0) return null;

  const visibleAuthorIds = authorIds.slice(0, MAX_VISIBLE_COMMENT_AUTHORS);
  const overflowCount = authorIds.length - visibleAuthorIds.length;

  return (
    <AvatarGroup className={className}>
      {visibleAuthorIds.map((userId) => (
        <CommentAuthorAvatar key={userId} userId={userId} />
      ))}
      {overflowCount > 0 && (
        <AvatarGroupCount>+{overflowCount}</AvatarGroupCount>
      )}
    </AvatarGroup>
  );
}
