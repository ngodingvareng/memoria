import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import {
  Item,
  ItemContent,
  ItemDescription,
  ItemMedia,
  ItemTitle,
} from '@/components/ui/item';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoNotificationResponse } from '@/lib/api/generated/models';
import {
  getGetNotificationsQueryKey,
  usePatchNotificationsIdRead,
} from '@/lib/api/generated/notifications/notifications';
import { useGetUserByID } from '@/lib/api/generated/users/users';
import { queryClient } from '@/lib/query-client';
import { cn } from '@/lib/utils';
import {
  AtIcon,
  Clock01Icon,
  GiftIcon,
  Mail01Icon,
  Message01Icon,
  UserAdd01Icon,
  UserCheck01Icon,
  UserGroupIcon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon, type HugeiconsIconProps } from '@hugeicons/react';

interface NotificationItemProps {
  notification: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoNotificationResponse;
}

// Kind -> icon + copy template. `actor` is the mentioner/inviter/etc. and
// is nil for gifts/reminders, which have no actor by design (FEATURES.md,
// Notification).
const kindPresentation: Record<
  string,
  { icon: HugeiconsIconProps['icon']; describe: (actor?: string) => string }
> = {
  mentioned_in_moment: {
    icon: AtIcon,
    describe: (actor) => `${actor ?? 'Someone'} mentioned you in a moment`,
  },
  circle_invite_received: {
    icon: Mail01Icon,
    describe: (actor) => `${actor ?? 'Someone'} invited you to a circle`,
  },
  circle_join_request_received: {
    icon: UserAdd01Icon,
    describe: (actor) => `${actor ?? 'Someone'} asked to join your circle`,
  },
  added_to_circle: {
    icon: UserGroupIcon,
    describe: (actor) => `${actor ?? 'Someone'} added you to a circle`,
  },
  circle_join_request_approved: {
    icon: UserCheck01Icon,
    describe: () => 'Your request to join was approved',
  },
  commitment_upcoming: {
    icon: Clock01Icon,
    describe: () => 'A commitment is coming up',
  },
  moment_due: {
    icon: Clock01Icon,
    describe: () => 'A moment is due',
  },
  moment_missed: {
    icon: Clock01Icon,
    describe: () => 'A moment was missed',
  },
  echo_ready: {
    icon: GiftIcon,
    describe: () => 'An echo is ready to look back on',
  },
  recap_generated: {
    icon: GiftIcon,
    describe: () => 'Your recap is ready',
  },
  response_digest: {
    icon: Message01Icon,
    describe: () => 'People responded to your moments today',
  },
};

export function NotificationItem({ notification }: NotificationItemProps) {
  const isUnread = !notification.read_at;
  const presentation = notification.kind
    ? kindPresentation[notification.kind]
    : undefined;

  const actorQuery = useGetUserByID(notification.actor_user_id ?? '', {
    query: { enabled: Boolean(notification.actor_user_id) },
  });

  const markRead = usePatchNotificationsIdRead();

  const handleClick = () => {
    if (!isUnread || !notification.id) return;
    markRead.mutate(
      { id: notification.id },
      {
        onSuccess: () => {
          queryClient.invalidateQueries({
            queryKey: getGetNotificationsQueryKey(),
          });
        },
      }
    );
  };

  return (
    <Item
      role="listitem"
      onClick={handleClick}
      className={cn('cursor-pointer', isUnread && 'bg-primary/5')}
    >
      <ItemMedia variant="icon">
        {actorQuery.data ? (
          <Avatar size="sm">
            <AvatarImage
              src={actorQuery.data.image_path ?? undefined}
              alt={actorQuery.data.name}
            />
            <AvatarFallback>
              {actorQuery.data.name?.charAt(0).toUpperCase()}
            </AvatarFallback>
          </Avatar>
        ) : (
          <HugeiconsIcon
            icon={presentation?.icon ?? Message01Icon}
            strokeWidth={2}
          />
        )}
      </ItemMedia>
      <ItemContent>
        <ItemTitle>
          {presentation?.describe(actorQuery.data?.name) ?? 'New notification'}
        </ItemTitle>
        {notification.created_at && (
          <ItemDescription>
            {new Date(notification.created_at).toLocaleString()}
          </ItemDescription>
        )}
      </ItemContent>
    </Item>
  );
}
