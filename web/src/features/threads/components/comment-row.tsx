import { Alert, AlertDescription } from '@/components/ui/alert';
import { ConfirmDestructiveDialog } from '@/components/dialogs/confirm-destructive-dialog';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Textarea } from '@/components/ui/textarea';
import { CircleNameLabel } from '@/features/circles';
import { getApiErrorMessage } from '@/lib/api-client';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoCommentResponse } from '@/lib/api/generated/models';
import {
  getGetMomentsIdCommentsQueryKey,
  useDeleteMomentsIdCommentsCommentId,
  usePutMomentsIdCommentsCommentId,
} from '@/lib/api/generated/comments/comments';
import { useGetUserByID } from '@/lib/api/generated/users/users';
import dayjs from '@/lib/dayjs';
import { queryClient } from '@/lib/query-client';
import { useSession } from '@/lib/session';
import {
  Delete02Icon,
  Edit04Icon,
  MoreVerticalIcon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { useState } from 'react';

interface CommentRowProps {
  momentId: string;
  comment: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoCommentResponse;
  momentOwnerId?: string;
}

export function CommentRow({
  momentId,
  comment,
  momentOwnerId,
}: CommentRowProps) {
  const session = useSession();
  const isDeletedAuthor = !comment.user_id;
  const userQuery = useGetUserByID(comment.user_id ?? '', {
    query: { enabled: !isDeletedAuthor },
  });
  const updateComment = usePutMomentsIdCommentsCommentId();
  const deleteComment = useDeleteMomentsIdCommentsCommentId();

  const [isEditing, setIsEditing] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [draft, setDraft] = useState(comment.body ?? '');
  const [error, setError] = useState<string | null>(null);

  const isOwn = !isDeletedAuthor && comment.user_id === session?.user.id;
  const isAuthorMomentOwner = comment.user_id === momentOwnerId;
  const showBadge = !(comment.circle_id == null && isAuthorMomentOwner);
  const authorName = isDeletedAuthor
    ? 'Deleted user'
    : (userQuery.data?.name ?? '');

  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: getGetMomentsIdCommentsQueryKey(momentId),
    });

  const handleSave = async () => {
    setError(null);
    try {
      await updateComment.mutateAsync({
        id: momentId,
        commentId: comment.id!,
        data: { body: draft },
      });
      setIsEditing(false);
      await invalidate();
    } catch (err) {
      setError(getApiErrorMessage(err, 'Failed to save. Please try again.'));
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-2">
        <Avatar size="sm">
          <AvatarImage
            src={userQuery.data?.image_path ?? undefined}
            alt={authorName}
          />
          <AvatarFallback>{authorName.charAt(0).toUpperCase()}</AvatarFallback>
        </Avatar>
        <div className="flex min-w-0 flex-1 flex-col gap-1">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium">{authorName}</span>
            {showBadge && (
              <Badge variant="secondary">
                {comment.circle_id ? (
                  <CircleNameLabel circleId={comment.circle_id} />
                ) : (
                  'Mentioned'
                )}
              </Badge>
            )}
            {comment.created_at && (
              <span className="text-muted-foreground text-xs">
                {dayjs(comment.created_at).fromNow()}
              </span>
            )}
          </div>
        </div>
        {isOwn && (
          <DropdownMenu>
            <DropdownMenuTrigger
              render={<Button variant="ghost" size="icon" />}
            >
              <HugeiconsIcon icon={MoreVerticalIcon} strokeWidth={2} />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start">
              <DropdownMenuGroup>
                <DropdownMenuItem onClick={() => setIsEditing(true)}>
                  <HugeiconsIcon icon={Edit04Icon} strokeWidth={2} />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem
                  variant="destructive"
                  onClick={() => setIsDeleteOpen(true)}
                >
                  <HugeiconsIcon icon={Delete02Icon} strokeWidth={2} />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuGroup>
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </div>
      {isEditing ? (
        <div className="flex flex-col gap-1">
          <Textarea value={draft} onChange={(e) => setDraft(e.target.value)} />
          <div className="flex justify-end gap-1">
            <Button
              type="button"
              size="sm"
              variant="ghost"
              onClick={() => {
                setIsEditing(false);
                setDraft(comment.body ?? '');
              }}
            >
              Cancel
            </Button>
            <Button
              type="button"
              size="sm"
              onClick={handleSave}
              disabled={updateComment.isPending}
            >
              Save
            </Button>
          </div>
        </div>
      ) : (
        <p className="whitespace-pre-wrap">{comment.body}</p>
      )}
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      <ConfirmDestructiveDialog
        open={isDeleteOpen}
        onOpenChange={setIsDeleteOpen}
        title="Delete this comment?"
        description="This can't be undone."
        confirmLabel="Delete"
        errorFallback="Failed to delete. Please try again."
        onConfirm={async () => {
          await deleteComment.mutateAsync({
            id: momentId,
            commentId: comment.id!,
          });
          await invalidate();
        }}
      />
    </div>
  );
}
