import {
  deleteMomentsIdMentionsMentionId,
  getGetMomentsIdMentionsQueryKey,
  getGetMomentsIdMentionsQueryOptions,
  postMomentsIdMentions,
} from '@/lib/api/generated/mentions/mentions';
import { getGetUserByIDQueryOptions } from '@/lib/api/generated/users/users';
import { queryClient } from '@/lib/query-client';
import { extractMentionUsernames } from './mention-text';

export interface MentionShareOffer {
  displayName: string;
  circleIds: string[];
}

// The note text is the source of truth for a Moment's mentions: this
// diffs the @usernames currently written in the text against the
// Mention records already stored, then creates/removes the difference.
// Used for both create (existing mentions is always empty) and edit.
// Anonymized mentions (mentioned_user_id is null — the account was
// deleted) are never touched, since a removed username can never match
// one anyway.
export async function syncMomentMentionsFromText(
  momentId: string,
  noteText: string
): Promise<{ shareOffers: MentionShareOffer[] }> {
  const desiredUsernames = new Set(extractMentionUsernames(noteText));

  const existing = await queryClient.fetchQuery(
    getGetMomentsIdMentionsQueryOptions(momentId)
  );
  const existingMentions = existing.mentions ?? [];

  const resolvedExisting = await Promise.all(
    existingMentions.map(async (mention) => {
      if (!mention.mentioned_user_id) {
        return { mention, username: null as string | null };
      }
      const user = await queryClient.fetchQuery(
        getGetUserByIDQueryOptions(mention.mentioned_user_id)
      );
      return { mention, username: user.username?.toLowerCase() ?? null };
    })
  );

  const existingUsernames = new Set(
    resolvedExisting
      .map((resolved) => resolved.username)
      .filter((username): username is string => username !== null)
  );

  const usernamesToAdd = [...desiredUsernames].filter(
    (username) => !existingUsernames.has(username)
  );
  const mentionsToRemove = resolvedExisting.filter(
    (resolved) =>
      resolved.username !== null && !desiredUsernames.has(resolved.username)
  );

  const [addResults] = await Promise.all([
    Promise.allSettled(
      usernamesToAdd.map((username) =>
        postMomentsIdMentions(momentId, { username })
      )
    ),
    Promise.allSettled(
      mentionsToRemove.map((resolved) =>
        deleteMomentsIdMentionsMentionId(momentId, resolved.mention.id!)
      )
    ),
  ]);

  if (usernamesToAdd.length > 0 || mentionsToRemove.length > 0) {
    await queryClient.invalidateQueries({
      queryKey: getGetMomentsIdMentionsQueryKey(momentId),
    });
  }

  const shareOffers: MentionShareOffer[] = addResults
    .filter(
      (
        result
      ): result is PromiseFulfilledResult<
        Awaited<ReturnType<typeof postMomentsIdMentions>>
      > => result.status === 'fulfilled'
    )
    .map((result) => result.value)
    .filter((mention) => (mention.shared_circle_ids?.length ?? 0) > 0)
    .map((mention) => ({
      displayName: mention.display_name ?? '',
      circleIds: mention.shared_circle_ids!,
    }));

  return { shareOffers };
}
