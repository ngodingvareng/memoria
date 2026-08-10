import { NotificationHeaderLink } from '@/features/notifications';
import { UserHeaderMenu } from '@/features/users';
import AppHeaderTitle from './app-header-title';
import SearchAnythingInput from './search-anything-input';
import { SidebarTrigger } from './ui/sidebar';
import { CreateMenu } from './create-menu';

export default function AppHeader() {
  return (
    <header className="bg-background fixed top-0 z-50 grid h-16 w-full grid-cols-3 items-center gap-4 border-b px-4">
      <div className="flex items-center gap-2">
        <SidebarTrigger size="icon-lg" className="[&_svg]:size-5!" />
        <AppHeaderTitle />
      </div>

      <div className="flex justify-center">
        <SearchAnythingInput />
      </div>

      <div className="flex items-center justify-end gap-3 pr-2">
        <CreateMenu />
        <NotificationHeaderLink />
        <UserHeaderMenu />
      </div>
    </header>
  );
}
