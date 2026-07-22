import {
  MoreVerticalIcon,
  Notification01Icon,
  PlusSignIcon,
  Search01Icon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';
import { InputGroup, InputGroupAddon, InputGroupInput } from './ui/input-group';
import { SidebarTrigger } from './ui/sidebar';
import UserMenu from './user-menu';
import { ButtonGroup } from './ui/button-group';
import { Button } from './ui/button';

export default function AppHeader() {
  return (
    <header className="h-16 fixed w-full gap-4 justify-between top-0 z-50 bg-background border-b px-4 flex items-center">
      <div className="flex items-center gap-2">
        <SidebarTrigger size="icon-lg" className="[&_svg]:size-5!" />
        <Link to="/" className="text-3xl font-semibold font-heading">
          Memoria
        </Link>
      </div>

      <div className="flex items-center gap-3 pr-2 justify-end grow">
        <InputGroup className="max-w-xl w-full h-10 [&_svg]:size-5!">
          <InputGroupInput
            placeholder="Search for stories and activities..."
            type="search"
            className="text-base!"
          />
          <InputGroupAddon>
            <HugeiconsIcon strokeWidth={2} icon={Search01Icon} />
          </InputGroupAddon>
        </InputGroup>
        <div className="flex gap-2">
          <ButtonGroup>
            <Button size="lg">
              <HugeiconsIcon icon={PlusSignIcon} /> New Activity
            </Button>
            <Button size="icon-lg">
              <HugeiconsIcon icon={MoreVerticalIcon} />
            </Button>
          </ButtonGroup>
        </div>
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
