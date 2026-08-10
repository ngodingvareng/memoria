import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetUserByID } from '@/lib/api/generated/users/users';
import { cn } from '@/lib/utils';

interface CircleUserChipProps {
  userId: string;
  avatarSize?: 'sm' | 'default' | 'lg' | 'xl';
  className?: string;
}

export function CircleUserChip({
  userId,
  avatarSize = 'default',
  className,
}: CircleUserChipProps) {
  const { data: user, isPending } = useGetUserByID(userId);

  if (isPending) {
    return (
      <div className={cn('flex items-center gap-2', className)}>
        <Skeleton className="size-8 shrink-0 rounded-full" />
        <Skeleton className="h-4 w-24" />
      </div>
    );
  }

  const name = user?.name ?? 'Unknown user';

  return (
    <div className={cn('flex min-w-0 items-center gap-2', className)}>
      <Avatar size={avatarSize}>
        <AvatarImage src={user?.image_path ?? undefined} alt={name} />
        <AvatarFallback>{name.charAt(0).toUpperCase()}</AvatarFallback>
      </Avatar>
      <div className="min-w-0">
        <p className="truncate leading-tight font-medium">{name}</p>
        {user?.username && (
          <p className="text-muted-foreground truncate text-sm leading-tight">
            @{user.username}
          </p>
        )}
      </div>
    </div>
  );
}
