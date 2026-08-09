import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { InputGroup, InputGroupTextarea } from '@/components/ui/input-group';
import { AudiencePicker } from './audience-picker';
import {
  toCircleId,
  useMomentAudience,
  type AudienceContext,
} from '@/features/moments/lib/use-moment-audience';
import { ApiError } from '@/lib/api-client';
import {
  getGetMomentsIdCommentsQueryKey,
  usePostMomentsIdComments,
} from '@/lib/api/generated/comments/comments';
import { queryClient } from '@/lib/query-client';
import { useState } from 'react';

interface CommentInputProps {
  momentId: string;
}

// Comments are flat, never threaded (FEATURES.md, Comment) — this is
// the only compose surface, there is no reply-to-comment.
export function CommentInput({ momentId }: CommentInputProps) {
  const audience = useMomentAudience(momentId);
  const createComment = usePostMomentsIdComments();

  const [selectedContext, setSelectedContext] = useState<
    AudienceContext | undefined
  >(undefined);
  const [body, setBody] = useState('');
  const [error, setError] = useState<string | null>(null);

  const activeContext = selectedContext ?? audience.defaultContext;
  const showAudiencePicker = audience.contexts.length > 1;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!body.trim() || !activeContext) return;

    setError(null);
    try {
      await createComment.mutateAsync({
        id: momentId,
        data: { circle_id: toCircleId(activeContext), body },
      });
      setBody('');
      await queryClient.invalidateQueries({
        queryKey: getGetMomentsIdCommentsQueryKey(momentId),
      });
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Failed to post comment. Please try again.'
      );
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-2">
      {showAudiencePicker && (
        <AudiencePicker
          contexts={audience.contexts}
          value={activeContext}
          onChange={setSelectedContext}
        />
      )}
      <InputGroup>
        <InputGroupTextarea
          value={body}
          onChange={(e) => setBody(e.target.value)}
          placeholder="Add a comment..."
          maxLength={10000}
        />
      </InputGroup>
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      <div className="flex justify-end">
        <Button
          type="submit"
          size="sm"
          disabled={!body.trim() || !activeContext || createComment.isPending}
        >
          Post
        </Button>
      </div>
    </form>
  );
}
