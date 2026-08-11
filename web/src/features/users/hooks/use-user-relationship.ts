import {
  getGetUsersMeBlocksQueryKey,
  getGetUsersMeKnownsQueryKey,
  getGetUsersMeMutesQueryKey,
  useDeleteUsersMeBlocksUsername,
  useDeleteUsersMeKnownsUsername,
  useDeleteUsersMeMutesUsername,
  useGetUsersMeBlocks,
  useGetUsersMeKnowns,
  useGetUsersMeMutes,
  usePostUsersMeBlocks,
  usePostUsersMeKnowns,
  usePostUsersMeMutes,
} from '@/lib/api/generated/users/users';
import { queryClient } from '@/lib/query-client';

// Known, Block, and Mute are three independent unilateral stances one
// user can take toward another (FEATURES.md, Privacy & Control). The
// profile page, the moment feed's menu, and the Known People settings
// list all need the same derived state and the same toggle actions
// toward a given username, so it's derived here once rather than
// re-implemented per surface — keeping all three surfaces in sync.
export function useUserRelationship(username: string) {
  const knownQuery = useGetUsersMeKnowns();
  const blockedQuery = useGetUsersMeBlocks();
  const mutedQuery = useGetUsersMeMutes();

  const isKnown =
    knownQuery.data?.users?.some((u) => u.username === username) ?? false;
  const isBlocked =
    blockedQuery.data?.users?.some((u) => u.username === username) ?? false;
  const isMuted =
    mutedQuery.data?.users?.some((u) => u.username === username) ?? false;

  const markKnown = usePostUsersMeKnowns();
  const unmarkKnown = useDeleteUsersMeKnownsUsername();
  const block = usePostUsersMeBlocks();
  const unblock = useDeleteUsersMeBlocksUsername();
  const mute = usePostUsersMeMutes();
  const unmute = useDeleteUsersMeMutesUsername();

  const toggleKnown = async () => {
    if (isKnown) {
      await unmarkKnown.mutateAsync({ username });
    } else {
      await markKnown.mutateAsync({ data: { username } });
    }
    await queryClient.invalidateQueries({
      queryKey: getGetUsersMeKnownsQueryKey(),
    });
  };

  const toggleBlock = async () => {
    if (isBlocked) {
      await unblock.mutateAsync({ username });
    } else {
      await block.mutateAsync({ data: { username } });
    }
    await queryClient.invalidateQueries({
      queryKey: getGetUsersMeBlocksQueryKey(),
    });
  };

  const toggleMute = async () => {
    if (isMuted) {
      await unmute.mutateAsync({ username });
    } else {
      await mute.mutateAsync({ data: { username } });
    }
    await queryClient.invalidateQueries({
      queryKey: getGetUsersMeMutesQueryKey(),
    });
  };

  return {
    isKnown,
    isBlocked,
    isMuted,
    isLoading:
      knownQuery.isPending || blockedQuery.isPending || mutedQuery.isPending,
    isTogglingKnown: markKnown.isPending || unmarkKnown.isPending,
    isTogglingBlock: block.isPending || unblock.isPending,
    isTogglingMute: mute.isPending || unmute.isPending,
    toggleKnown,
    toggleBlock,
    toggleMute,
  };
}
