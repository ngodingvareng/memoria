import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoPublicUserResponse } from '@/lib/api/generated/models';
import * as React from 'react';
import {
  ACTIVE_MENTION_TOKEN_PATTERN,
  insertMentionToken,
} from '../lib/mention-text';
import { useMentionSearch } from './use-mention-search';

type MentionableUser =
  GithubComNgodingvarengMemoriaInternalDeliveryRestDtoPublicUserResponse;

interface ActiveToken {
  // Text offsets of the token, including the leading "@".
  start: number;
  end: number;
  query: string;
}

interface UseInlineMentionAutocompleteOptions {
  onChange: (value: string) => void;
  textareaRef: React.RefObject<HTMLTextAreaElement | null>;
}

// Drives the "typing @ triggers a live suggestion dropdown" behavior
// shared by every note textarea (create form, thread compose box, edit
// mode) — tracks the @token immediately before the cursor, feeds it to
// the mention search, and handles inserting a picked user back into the
// text at the right offsets.
export function useInlineMentionAutocomplete({
  onChange,
  textareaRef,
}: UseInlineMentionAutocompleteOptions) {
  const [activeToken, setActiveToken] = React.useState<ActiveToken | null>(
    null
  );
  const [activeIndex, setActiveIndex] = React.useState(0);

  const { users, isLoading } = useMentionSearch(activeToken?.query ?? '');

  const recomputeActiveToken = React.useCallback(() => {
    const el = textareaRef.current;
    if (!el || el.selectionStart !== el.selectionEnd) {
      setActiveToken(null);
      return;
    }

    const cursor = el.selectionStart ?? 0;
    const match = el.value.slice(0, cursor).match(ACTIVE_MENTION_TOKEN_PATTERN);
    if (!match) {
      setActiveToken(null);
      return;
    }

    const query = match[1];
    setActiveToken({ start: cursor - query.length - 1, end: cursor, query });
    setActiveIndex(0);
  }, [textareaRef]);

  const handleChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    onChange(event.target.value);
    recomputeActiveToken();
  };

  const handleSelect = (user: MentionableUser) => {
    const el = textareaRef.current;
    if (!activeToken || !user.username || !el) return;

    const { text, cursor } = insertMentionToken(
      el.value,
      activeToken.start,
      activeToken.end,
      user.username
    );
    onChange(text);
    setActiveToken(null);

    requestAnimationFrame(() => {
      el.focus();
      el.setSelectionRange(cursor, cursor);
    });
  };

  // Used by the "Add mention" button: inserts at the current cursor
  // rather than replacing an active "@token" (there isn't one).
  const insertAtCursor = (user: MentionableUser) => {
    const el = textareaRef.current;
    if (!user.username || !el) return;

    const cursor = el.selectionStart ?? el.value.length;
    const { text, cursor: nextCursor } = insertMentionToken(
      el.value,
      cursor,
      cursor,
      user.username
    );
    onChange(text);
    setActiveToken(null);

    requestAnimationFrame(() => {
      el.focus();
      el.setSelectionRange(nextCursor, nextCursor);
    });
  };

  const handleKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (!activeToken || users.length === 0) return;

    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault();
        setActiveIndex((i) => (i + 1) % users.length);
        break;
      case 'ArrowUp':
        event.preventDefault();
        setActiveIndex((i) => (i - 1 + users.length) % users.length);
        break;
      case 'Enter':
      case 'Tab':
        event.preventDefault();
        handleSelect(users[activeIndex]);
        break;
      case 'Escape':
        setActiveToken(null);
        break;
    }
  };

  return {
    isOpen: activeToken !== null,
    suggestions: users,
    isLoading,
    activeIndex,
    handleChange,
    handleKeyDown,
    handleSelect,
    insertAtCursor,
    handleSelectionChange: recomputeActiveToken,
    closeSuggestions: () => setActiveToken(null),
  };
}
