export const MENTION_PATTERN = /@([a-zA-Z0-9_.]{3,30})/g;

// The active "@token" immediately before the cursor, e.g. typing
// "hi @ge" has an active token of "ge" — used to drive the inline
// autocomplete dropdown while composing a note.
export const ACTIVE_MENTION_TOKEN_PATTERN = /(?:^|\s)@([a-zA-Z0-9_.]{0,30})$/;

// Every distinct @username referenced in a note, lowercased for
// case-insensitive comparison against stored Mention records (usernames
// are case-insensitive, see uq_users_username_lower).
export function extractMentionUsernames(text: string): string[] {
  const usernames = new Set<string>();
  for (const match of text.matchAll(MENTION_PATTERN)) {
    usernames.add(match[1].toLowerCase());
  }
  return [...usernames];
}

// Splices "@username " into text at the given range (start === end for
// a plain cursor insert), returning the new text and where the cursor
// should land afterward. Shared by the inline autocomplete's token
// replacement and the "Add mention" button's cursor insert.
export function insertMentionToken(
  text: string,
  start: number,
  end: number,
  username: string
): { text: string; cursor: number } {
  const inserted = `@${username} `;
  return {
    text: text.slice(0, start) + inserted + text.slice(end),
    cursor: start + inserted.length,
  };
}
