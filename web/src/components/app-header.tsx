import { CreateHeaderLink } from '@/features/activities';
import { NotificationHeaderLink } from '@/features/notifications';
import { UserHeaderMenu } from '@/features/users';
import AppHeaderTitle from './app-header-title';
import SearchAnythingInput from './search-anything-input';
import { SidebarTrigger } from './ui/sidebar';

export default function AppHeader() {
  return (
    <header className="h-16 fixed w-full gap-4 justify-between top-0 z-50 bg-background border-b px-4 flex items-center">
      <div className="flex items-center gap-2">
        <SidebarTrigger size="icon-lg" className="[&_svg]:size-5!" />
        <AppHeaderTitle />
      </div>

      <div className="flex items-center gap-3 pr-2 justify-end grow">
        <SearchAnythingInput />
        <CreateHeaderLink />
        <NotificationHeaderLink />
        <UserHeaderMenu />
      </div>
    </header>
  );
}
