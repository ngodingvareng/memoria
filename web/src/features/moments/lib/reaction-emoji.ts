// FEATURES.md, Reaction: a fixed emoji set, always shown in this order.
export const REACTION_KINDS = ['heart', 'joy', 'tender'] as const;

export type ReactionKind = (typeof REACTION_KINDS)[number];

export const REACTION_EMOJI: Record<ReactionKind, string> = {
  heart: '❤️',
  joy: '😂',
  tender: '🥹',
};

export function isReactionKind(value: string): value is ReactionKind {
  return (REACTION_KINDS as readonly string[]).includes(value);
}
