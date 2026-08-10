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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { getApiErrorMessage } from '@/lib/api-client';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoPublicUserResponse } from '@/lib/api/generated/models';
import {
  getGetUsersMeBlocksQueryKey,
  getGetUsersMeMutesQueryKey,
  useDeleteUsersMeBlocksUsername,
  useDeleteUsersMeMutesUsername,
  useGetUsersMeBlocks,
  useGetUsersMeMutes,
  usePostUsersMeBlocks,
  usePostUsersMeMutes,
} from '@/lib/api/generated/users/users';
import { queryClient } from '@/lib/query-client';
import * as React from 'react';

export function BlockedMutedUsersList() {
  return (
    <Tabs defaultValue="blocked">
      <TabsList>
        <TabsTrigger value="blocked">Blocked</TabsTrigger>
        <TabsTrigger value="muted">Muted</TabsTrigger>
      </TabsList>
      <TabsContent value="blocked">
        <UserRestrictionList kind="blocked" />
      </TabsContent>
      <TabsContent value="muted">
        <UserRestrictionList kind="muted" />
      </TabsContent>
    </Tabs>
  );
}

interface UserRestrictionListProps {
  kind: 'blocked' | 'muted';
}

// Block and Mute share the exact same "add by username, list, remove"
// shape (FEATURES.md, Privacy & Control) — only the underlying
// generated hooks differ, so this switches on `kind` once here rather
// than duplicating the list/add/remove UI twice.
function UserRestrictionList({ kind }: UserRestrictionListProps) {
  const [username, setUsername] = React.useState('');
  const [addError, setAddError] = React.useState<string | null>(null);

  const blockedQuery = useGetUsersMeBlocks({
    query: { enabled: kind === 'blocked' },
  });
  const mutedQuery = useGetUsersMeMutes({
    query: { enabled: kind === 'muted' },
  });

  const addBlock = usePostUsersMeBlocks();
  const addMute = usePostUsersMeMutes();
  const removeBlock = useDeleteUsersMeBlocksUsername();
  const removeMute = useDeleteUsersMeMutesUsername();

  const query = kind === 'blocked' ? blockedQuery : mutedQuery;
  const users = query.data?.users ?? [];
  const noun = kind === 'blocked' ? 'blocked' : 'muted';

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username.trim()) return;
    setAddError(null);
    try {
      if (kind === 'blocked') {
        await addBlock.mutateAsync({ data: { username } });
        await queryClient.invalidateQueries({
          queryKey: getGetUsersMeBlocksQueryKey(),
        });
      } else {
        await addMute.mutateAsync({ data: { username } });
        await queryClient.invalidateQueries({
          queryKey: getGetUsersMeMutesQueryKey(),
        });
      }
      setUsername('');
    } catch (err) {
      setAddError(
        getApiErrorMessage(err, `Could not ${noun.slice(0, -2)} that user.`)
      );
    }
  };

  const handleRemove = async (
    target: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoPublicUserResponse
  ) => {
    if (!target.username) return;
    if (kind === 'blocked') {
      await removeBlock.mutateAsync({ username: target.username });
      await queryClient.invalidateQueries({
        queryKey: getGetUsersMeBlocksQueryKey(),
      });
    } else {
      await removeMute.mutateAsync({ username: target.username });
      await queryClient.invalidateQueries({
        queryKey: getGetUsersMeMutesQueryKey(),
      });
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
            {kind === 'blocked' ? 'Block' : 'Mute'}
          </InputGroupButton>
        </InputGroup>
        {addError && (
          <p className="text-destructive text-sm mt-1">{addError}</p>
        )}
      </form>

      {query.isPending ? (
        <Skeleton className="h-12 w-full rounded-2xl" />
      ) : users.length === 0 ? (
        <Empty className="p-6">
          <EmptyHeader>
            <EmptyTitle>No {noun} users</EmptyTitle>
            <EmptyDescription>
              {kind === 'blocked'
                ? 'Blocked users can no longer see, mention, comment, or react to your moments.'
                : "Muting hides someone's activity from your own view only."}
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <div className="flex flex-col gap-2">
          {users.map((u) => (
            <Item key={u.id}>
              <Avatar size="sm">
                <AvatarImage src={u.image_path ?? undefined} alt={u.name} />
                <AvatarFallback>
                  {u.name?.charAt(0).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <ItemContent>
                <ItemTitle>{u.name}</ItemTitle>
              </ItemContent>
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleRemove(u)}
              >
                {kind === 'blocked' ? 'Unblock' : 'Unmute'}
              </Button>
            </Item>
          ))}
        </div>
      )}
    </div>
  );
}
