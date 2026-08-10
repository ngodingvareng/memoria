import { cn } from '@/lib/utils';
import { AtIcon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';

interface MentionBadgeProps {
  username: string;
  className?: string;
}

export function MentionBadge({ username, className }: MentionBadgeProps) {
  return (
    <Link
      to="/@{$username}"
      params={{ username }}
      onClick={(e) => e.stopPropagation()}
      className={cn(
        'font-bold text-blue-600 underline-offset-2 hover:underline dark:text-blue-400',
        className
      )}
    >
      <HugeiconsIcon icon={AtIcon} className="inline" />
      {username}
    </Link>
  );
}
