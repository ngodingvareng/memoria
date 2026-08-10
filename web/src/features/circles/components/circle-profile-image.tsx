import type { ComponentProps } from 'react';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { cn } from '@/lib/utils';

type CircleProfileImageProps = Omit<
  ComponentProps<typeof Avatar>,
  'children'
> & {
  circle: {
    name?: string | null;
    image_path?: string | null;
  };
};

export function CircleProfileImage({
  circle,
  className,
  size = 'default',
  ...props
}: CircleProfileImageProps) {
  const initial = circle.name?.trim().charAt(0).toUpperCase() ?? '?';
  const isSmall = size === 'sm';
  const roundedClass = isSmall
    ? 'rounded-sm after:rounded-sm'
    : 'rounded-xl after:rounded-xl';
  const imageClass = isSmall ? 'rounded-sm' : 'rounded-xl';

  return (
    <Avatar size={size} className={cn(roundedClass, className)} {...props}>
      <AvatarImage
        src={circle.image_path ?? undefined}
        alt={circle.name ?? 'Circle profile'}
        className={imageClass}
      />
      <AvatarFallback className={imageClass}>{initial}</AvatarFallback>
    </Avatar>
  );
}
