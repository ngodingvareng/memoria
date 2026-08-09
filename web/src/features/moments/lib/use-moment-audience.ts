import { useGetMomentsIdAudience } from '@/lib/api/generated/comments/comments';

// Which Response context (FEATURES.md) a comment/reaction is posted
// in — mirrors the api's circle_id: undefined/null means the mention
// context, a circle_id names that Circle's audience.
export type AudienceContext =
  { type: 'mention' } | { type: 'circle'; circleId: string };

export interface MomentAudience {
  isPending: boolean;
  contexts: AudienceContext[];
  // Auto-picked when exactly one context is valid. undefined when the
  // viewer has none (nothing to post/react with) or more than one
  // (the caller must show a picker — see AudiencePicker).
  defaultContext: AudienceContext | undefined;
}

// useMomentAudience resolves which contexts the viewer may post a
// comment/reaction in for a Moment — a viewer can have zero, one, or
// several (e.g. a Circle member who is also mentioned).
export function useMomentAudience(momentId: string): MomentAudience {
  const query = useGetMomentsIdAudience(momentId);

  const contexts: AudienceContext[] = [
    ...(query.data?.mention_allowed ? [{ type: 'mention' as const }] : []),
    ...(query.data?.circle_ids ?? []).map((circleId): AudienceContext => ({
      type: 'circle',
      circleId,
    })),
  ];

  return {
    isPending: query.isPending,
    contexts,
    defaultContext: contexts.length === 1 ? contexts[0] : undefined,
  };
}

// toCircleId converts an AudienceContext into the circle_id value the
// api expects: undefined for the mention context, the circle's id
// otherwise.
export function toCircleId(
  context: AudienceContext | undefined
): string | undefined {
  return context?.type === 'circle' ? context.circleId : undefined;
}

export function audienceContextKey(context: AudienceContext): string {
  return context.type === 'mention' ? 'mention' : `circle:${context.circleId}`;
}
