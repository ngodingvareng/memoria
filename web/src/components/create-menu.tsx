import {
  MoreVerticalSquare01Icon,
  PlusSignIcon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';
import { Button } from './ui/button';
import { ButtonGroup } from './ui/button-group';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from './ui/dropdown-menu';

const data = {
  menuItems: [
    {
      label: 'New Activity',
      link: '/activity/new',
    },
    {
      label: 'New Circle',
      link: '/circle/new',
    },
  ],
};

export function CreateMenu() {
  return (
    <ButtonGroup>
      <Button
        render={
          <Link to="/capture/new">
            <HugeiconsIcon icon={PlusSignIcon} strokeWidth={2} /> New Capture
          </Link>
        }
      />
      <DropdownMenu>
        <DropdownMenuTrigger render={<Button />}>
          <HugeiconsIcon icon={MoreVerticalSquare01Icon} strokeWidth={2} />
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuGroup>
            {data.menuItems.map((item, index) => (
              <DropdownMenuItem
                key={index}
                render={<Link to={item.link}>{item.label}</Link>}
              />
            ))}
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </ButtonGroup>
  );
}
