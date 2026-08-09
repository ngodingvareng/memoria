import { cn } from '@/lib/utils';
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
        'font-medium text-primary hover:underline underline-offset-2',
        className
      )}
    >
      @{username}
    </Link>
  );
}
