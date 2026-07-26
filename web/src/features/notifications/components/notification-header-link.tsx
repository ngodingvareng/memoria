import { Notification01Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';

export function NotificationHeaderLink() {
  return (
    <Link
      to="/notifications"
      className="size-10 flex justify-center items-center bg-primary/10 rounded-full"
    >
      <HugeiconsIcon icon={Notification01Icon} strokeWidth={2} />
    </Link>
  );
}
