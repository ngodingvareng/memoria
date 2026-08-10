import { Fragment } from 'react';
import { MentionBadge } from '../components/mention-badge';
import { MENTION_PATTERN } from './mention-text';

export function renderTextWithMentions(text: string): React.ReactNode {
  return text
    .split(MENTION_PATTERN)
    .map((part, index) =>
      index % 2 === 1 ? (
        <MentionBadge key={`${index}-${part}`} username={part} />
      ) : (
        <Fragment key={index}>{part}</Fragment>
      )
    );
}
