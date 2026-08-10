import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { useGetUserByID } from '@/lib/api/generated/users/users';

interface ReactorAvatarProps {
  userId: string;
}

export function ReactorAvatar({ userId }: ReactorAvatarProps) {
  const { data: user } = useGetUserByID(userId);
  const name = user?.name ?? '';

  return (
    <span title={name}>
      <Avatar size="sm" className="ring-background ring-2">
        <AvatarImage src={user?.image_path ?? undefined} alt={name} />
        <AvatarFallback>{name.charAt(0).toUpperCase()}</AvatarFallback>
      </Avatar>
    </span>
  );
}
