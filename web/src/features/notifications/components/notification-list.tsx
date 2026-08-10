import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty';
import { ItemGroup, ItemSeparator } from '@/components/ui/item';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetNotifications } from '@/lib/api/generated/notifications/notifications';
import { Notification01Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Fragment } from 'react';
import { NotificationItem } from './notification-item';

export function NotificationList() {
  const notificationsQuery = useGetNotifications();
  const notifications = notificationsQuery.data?.notifications ?? [];

  if (notificationsQuery.isPending) {
    return (
      <div className="flex flex-col gap-3">
        <Skeleton className="h-16 w-full rounded-2xl" />
        <Skeleton className="h-16 w-full rounded-2xl" />
        <Skeleton className="h-16 w-full rounded-2xl" />
      </div>
    );
  }

  if (notifications.length === 0) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <HugeiconsIcon icon={Notification01Icon} strokeWidth={2} />
          </EmptyMedia>
          <EmptyTitle>No notifications yet</EmptyTitle>
          <EmptyDescription>
            You&apos;ll see mentions, circle invites, and other updates here.
          </EmptyDescription>
        </EmptyHeader>
        <EmptyContent />
      </Empty>
    );
  }

  return (
    <ItemGroup role="list">
      {notifications.map((notification, index) => (
        <Fragment key={notification.id}>
          <NotificationItem notification={notification} />
          {index !== notifications.length - 1 && <ItemSeparator />}
        </Fragment>
      ))}
    </ItemGroup>
  );
}
