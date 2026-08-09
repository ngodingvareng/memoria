import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { ApiError } from '@/lib/api-client';
import { getGetCirclesIdMembersQueryKey } from '@/lib/api/generated/circles/circles';
import { usePostCirclesIdMembersDirect } from '@/lib/api/generated/circle-invites/circle-invites';
import { queryClient } from '@/lib/query-client';
import { useState } from 'react';

interface InviteDirectFormProps {
  circleId: string;
}

export function InviteDirectForm({ circleId }: InviteDirectFormProps) {
  const [value, setValue] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<{
    added: number;
    pending: number;
    skipped: string[];
  } | null>(null);
  const inviteDirect = usePostCirclesIdMembersDirect();

  const usernames = value
    .split(/[\s,]+/)
    .map((u) => u.trim())
    .filter(Boolean);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setResult(null);
    if (usernames.length === 0) return;

    try {
      const response = await inviteDirect.mutateAsync({
        id: circleId,
        data: { usernames },
      });
      setResult({
        added: response.added_members?.length ?? 0,
        pending: response.pending_invites?.length ?? 0,
        skipped: response.skipped_usernames ?? [],
      });
      setValue('');
      await queryClient.invalidateQueries({
        queryKey: getGetCirclesIdMembersQueryKey(circleId),
      });
    } catch (err) {
      setError(
        err instanceof ApiError ? err.message : 'Failed to invite these users.'
      );
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <FieldGroup>
        <Field>
          <FieldLabel htmlFor="invite-usernames">Usernames</FieldLabel>
          <Input
            id="invite-usernames"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            placeholder="gede, ujang, faisal"
            autoComplete="off"
          />
          <FieldDescription>
            Separate multiple usernames with a comma or space.
          </FieldDescription>
        </Field>

        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {result && (
          <Alert>
            <AlertDescription>
              {result.added} added directly, {result.pending} left as a pending
              invite
              {result.skipped.length > 0 &&
                `, couldn't reach: ${result.skipped.join(', ')}`}
              .
            </AlertDescription>
          </Alert>
        )}

        <Field orientation="horizontal" className="justify-end">
          <Button
            type="submit"
            disabled={usernames.length === 0 || inviteDirect.isPending}
          >
            Invite
          </Button>
        </Field>
      </FieldGroup>
    </form>
  );
}
