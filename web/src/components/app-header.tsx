import { NotificationHeaderLink } from '@/features/notifications';
import { UserHeaderMenu } from '@/features/users';
import AppHeaderTitle from './app-header-title';
import SearchAnythingInput from './search-anything-input';
import { SidebarTrigger } from './ui/sidebar';
import { CreateMenu } from './create-menu';

export default function AppHeader() {
  return (
    <header className="h-16 fixed w-full gap-4 top-0 z-50 bg-background border-b px-4 grid grid-cols-3 items-center">
      <div className="flex items-center gap-2 ">
        <SidebarTrigger size="icon-lg" className="[&_svg]:size-5!" />
        <AppHeaderTitle />
      </div>

      <div className="flex justify-center">
        <SearchAnythingInput />
      </div>

      <div className="flex items-center gap-3 pr-2 justify-end ">
        <CreateMenu />
        <NotificationHeaderLink />
        <UserHeaderMenu />
      </div>
    </header>
  );
}
