import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetCirclesId } from '@/lib/api/generated/circles/circles';
import { cn } from '@/lib/utils';
import { CircleIcon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';

interface ThreadCircleBadgeProps {
  circleId: string;
  className?: string;
  // Icon-only, for tight spaces like the sidebar's Recent Threads row —
  // the circle's name still surfaces via the native title tooltip.
  compact?: boolean;
}

// Marks a Thread as owned by a Circle — the badge itself is the link
// to that Circle (FEATURES.md's collaborative Thread).
export function ThreadCircleBadge({
  circleId,
  className,
  compact = false,
}: ThreadCircleBadgeProps) {
  const { data: circle, isPending } = useGetCirclesId(circleId);

  if (isPending) {
    return (
      <Skeleton
        className={cn(
          compact ? 'size-4 rounded-full' : 'h-5 w-16 rounded-3xl',
          className
        )}
      />
    );
  }

  if (!circle) return null;

  if (compact) {
    return (
      <Link
        to="/c/$id"
        params={{ id: circleId }}
        title={circle.name}
        onClick={(e) => e.stopPropagation()}
        className={cn(
          'text-muted-foreground hover:text-foreground shrink-0',
          className
        )}
      >
        <HugeiconsIcon icon={CircleIcon} strokeWidth={3} className="size-3.5" />
        <span className="sr-only">{circle.name}</span>
      </Link>
    );
  }

  return (
    <Badge
      variant="secondary"
      className={className}
      render={
        <Link
          to="/c/$id"
          params={{ id: circleId }}
          onClick={(e) => e.stopPropagation()}
        />
      }
    >
      <HugeiconsIcon
        data-icon="inline-start"
        icon={CircleIcon}
        strokeWidth={3}
      />
      {circle.name}
    </Badge>
  );
}
