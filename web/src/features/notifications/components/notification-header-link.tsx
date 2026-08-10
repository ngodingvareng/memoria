import { Badge } from '@/components/ui/badge';
import { useGetNotifications } from '@/lib/api/generated/notifications/notifications';
import { Notification01Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';

export function NotificationHeaderLink() {
  const notificationsQuery = useGetNotifications();
  const unreadCount = notificationsQuery.data?.unread_count ?? 0;

  return (
    <Link
      to="/notifications"
      className="relative size-10 flex justify-center items-center bg-primary/10 rounded-full"
    >
      <HugeiconsIcon icon={Notification01Icon} strokeWidth={2} />
      {unreadCount > 0 && (
        <Badge className="absolute -top-1 -right-1 min-w-4 h-4 px-1 text-[10px]">
          {unreadCount > 99 ? '99+' : unreadCount}
        </Badge>
      )}
    </Link>
  );
}
