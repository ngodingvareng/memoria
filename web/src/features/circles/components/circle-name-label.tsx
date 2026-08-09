import { Skeleton } from '@/components/ui/skeleton';
import { useGetCirclesId } from '@/lib/api/generated/circles/circles';

interface CircleNameLabelProps {
  circleId: string;
  className?: string;
}

// A bare circle name, resolved by id — the text-only building block
// behind ThreadCircleBadge/AudiencePicker/comment badges, wherever a
// full linked badge isn't wanted.
export function CircleNameLabel({ circleId, className }: CircleNameLabelProps) {
  const { data: circle, isPending } = useGetCirclesId(circleId);

  if (isPending)
    return <Skeleton className={className ?? 'inline-block h-4 w-16'} />;

  return <span className={className}>{circle?.name}</span>;
}
