import { Fragment } from 'react';
import { MentionBadge } from '../components/mention-badge';

const MENTION_PATTERN = /@([a-zA-Z0-9_.]{3,30})/g;

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
