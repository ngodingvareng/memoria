import { ConfirmDestructiveDialog } from '@/components/dialogs/confirm-destructive-dialog';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from '@/components/ui/empty';
import {
  InputGroup,
  InputGroupButton,
  InputGroupInput,
} from '@/components/ui/input-group';
import { Item, ItemContent, ItemTitle } from '@/components/ui/item';
import { Skeleton } from '@/components/ui/skeleton';
import { getApiErrorMessage } from '@/lib/api-client';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoPublicUserResponse } from '@/lib/api/generated/models';
import {
  getGetUsersMeKnownsQueryKey,
  useGetUsersMeKnowns,
  usePostUsersMeKnowns,
} from '@/lib/api/generated/users/users';
import { queryClient } from '@/lib/query-client';
import * as React from 'react';
import { useUserRelationship } from '../hooks/use-user-relationship';

// Marking someone known primarily happens from their own profile page
// (FEATURES.md, Known: "search a username, tap once") — this add form
// mirrors Blocked/Muted's add-by-username form for consistency, since
// Known now shares their exact settings-list shape.
export function KnownPeopleList() {
  const [username, setUsername] = React.useState('');
  const [addError, setAddError] = React.useState<string | null>(null);

  const knownQuery = useGetUsersMeKnowns();
  const addKnown = usePostUsersMeKnowns();

  const users = knownQuery.data?.users ?? [];

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username.trim()) return;
    setAddError(null);
    try {
      await addKnown.mutateAsync({ data: { username } });
      await queryClient.invalidateQueries({
        queryKey: getGetUsersMeKnownsQueryKey(),
      });
      setUsername('');
    } catch (err) {
      setAddError(getApiErrorMessage(err, 'Could not mark that user known.'));
    }
  };

  return (
    <div className="flex flex-col gap-4 pt-4">
      <form onSubmit={handleAdd}>
        <InputGroup>
          <InputGroupInput
            placeholder="Username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
          />
          <InputGroupButton type="submit" disabled={!username.trim()}>
            Mark known
          </InputGroupButton>
        </InputGroup>
        {addError && (
          <p className="text-destructive mt-1 text-sm">{addError}</p>
        )}
      </form>

      {knownQuery.isPending ? (
        <Skeleton className="h-12 w-full rounded-2xl" />
      ) : users.length === 0 ? (
        <Empty className="p-6">
          <EmptyHeader>
            <EmptyTitle>No known people yet</EmptyTitle>
            <EmptyDescription>
              People you mark as known may mention you or invite you to a
              circle, depending on your privacy settings.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <div className="flex flex-col gap-2">
          {users.map((u) => (
            <KnownPersonRow key={u.id} user={u} />
          ))}
        </div>
      )}
    </div>
  );
}

interface KnownPersonRowProps {
  user: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoPublicUserResponse;
}

// Block/Mute here (item 3's "from the known people list" surface) reuse
// the same shared relationship hook the profile page and moment card
// menu use, so state never drifts between surfaces.
function KnownPersonRow({ user }: KnownPersonRowProps) {
  const relationship = useUserRelationship(user.username ?? '');
  const [error, setError] = React.useState<string | null>(null);

  const handleUnmark = async () => {
    setError(null);
    try {
      await relationship.toggleKnown();
    } catch (err) {
      setError(getApiErrorMessage(err, 'Could not unmark that user.'));
    }
  };

  const handleToggleMute = async () => {
    setError(null);
    try {
      await relationship.toggleMute();
    } catch (err) {
      setError(getApiErrorMessage(err, 'Could not mute that user.'));
    }
  };

  return (
    <div className="flex flex-col gap-1">
      <Item>
        <Avatar size="sm">
          <AvatarImage src={user.image_path ?? undefined} alt={user.name} />
          <AvatarFallback>{user.name?.charAt(0).toUpperCase()}</AvatarFallback>
        </Avatar>
        <ItemContent>
          <ItemTitle>{user.name}</ItemTitle>
        </ItemContent>
        <div className="flex gap-2">
          {relationship.isBlocked ? (
            <Button
              variant="outline"
              size="sm"
              onClick={relationship.toggleBlock}
              disabled={relationship.isTogglingBlock}
            >
              Unblock
            </Button>
          ) : (
            <ConfirmDestructiveDialog
              triggerRender={<Button variant="outline" size="sm" />}
              triggerLabel="Block"
              title={`Block ${user.name}?`}
              description="Neither of you will be able to see, mention, comment on, or react to the other's Moments, in any context."
              confirmLabel="Block"
              errorFallback="Failed to block this user. Please try again."
              onConfirm={relationship.toggleBlock}
            />
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={handleToggleMute}
            disabled={relationship.isTogglingMute}
          >
            {relationship.isMuted ? 'Unmute' : 'Mute'}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={handleUnmark}
            disabled={relationship.isTogglingKnown}
          >
            Unmark
          </Button>
        </div>
      </Item>
      {error && <p className="text-destructive text-sm">{error}</p>}
    </div>
  );
}
