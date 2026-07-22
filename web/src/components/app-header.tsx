import { Notification01Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';
import { SidebarTrigger } from './ui/sidebar';
import UserMenu from './user-menu';

export default function AppHeader() {
  return (
    <header className="h-16 fixed w-full top-0 z-50 bg-background border-b px-4 flex items-center">
      <div>
        <SidebarTrigger size="icon-lg" className="[&_svg]:size-5!" />
        <Link to="/" className="text-3xl px-3 py-2 font-semibold font-heading">
          Memoria
        </Link>
      </div>
      <div className="grow flex items-center gap-4 pr-2 justify-end">
        <Link
          to="/notifications"
          className="size-10 flex justify-center items-center bg-primary/10 rounded-full"
        >
          <HugeiconsIcon icon={Notification01Icon} strokeWidth={2} />
        </Link>

        <UserMenu />
      </div>
    </header>
  );
}
