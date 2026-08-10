import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { useGetUserByID } from '@/lib/api/generated/users/users';

interface CommentAuthorAvatarProps {
  userId: string;
}

export function CommentAuthorAvatar({ userId }: CommentAuthorAvatarProps) {
  const { data: user } = useGetUserByID(userId);
  const name = user?.name ?? '';

  return (
    <Avatar size="sm">
      <AvatarImage src={user?.image_path ?? undefined} alt={name} />
      <AvatarFallback>{name.charAt(0).toUpperCase()}</AvatarFallback>
    </Avatar>
  );
}
