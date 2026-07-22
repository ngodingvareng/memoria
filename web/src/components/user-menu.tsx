import {
  HelpCircleIcon,
  LanguageSquareIcon,
  Logout01Icon,
  Settings02Icon,
  Sun01Icon,
} from '@hugeicons/core-free-icons';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuPortal,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from './ui/dropdown-menu';
import { Avatar, AvatarFallback, AvatarImage } from './ui/avatar';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';

const data = {
  user: [
    [
      {
        title: 'Display Language: English',
        icon: LanguageSquareIcon,
        submenu: [
          {
            title: 'English',
          },
          {
            title: 'Indonesia',
          },
        ],
      },
      {
        title: 'Appereance: Device',
        icon: Sun01Icon,
        submenu: [
          {
            title: 'Light theme',
          },
          {
            title: 'Dark theme',
          },
          {
            title: 'Use device theme',
          },
        ],
      },
      {
        title: 'Settings',
        icon: Settings02Icon,
        url: '/user',
      },
    ],
    [
      {
        title: 'Sign Out',
        icon: Logout01Icon,
        url: '.',
      },
    ],
    [
      {
        title: 'Help',
        icon: HelpCircleIcon,
        url: '/help',
      },
    ],
  ],
};

export default function UserMenu() {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Avatar size="lg">
            <AvatarImage
              src="https://github.com/shadcn.png"
              alt="@shadcn"
              className="grayscale"
            />
            <AvatarFallback>CN</AvatarFallback>
          </Avatar>
        }
      />
      <DropdownMenuContent className="w-72">
        <DropdownMenuGroup>
          <div className="flex gap-3 items-center px-3 py-1">
            <Avatar>
              <AvatarImage
                src="https://github.com/shadcn.png"
                alt="@shadcn"
                className="grayscale"
              />
              <AvatarFallback>CN</AvatarFallback>
            </Avatar>
            <div>
              <p className="font-medium text-base/3">Rahmat</p>
              <p>@rahmat</p>
            </div>
          </div>
        </DropdownMenuGroup>
        {data.user.map((group) => (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              {group.map((item) =>
                item.submenu ? (
                  <DropdownMenuSub>
                    <DropdownMenuSubTrigger>
                      <HugeiconsIcon icon={item.icon} strokeWidth={2} />{' '}
                      {item.title}
                    </DropdownMenuSubTrigger>
                    <DropdownMenuPortal>
                      <DropdownMenuSubContent>
                        {item.submenu.map((subitem) => (
                          <DropdownMenuItem>{subitem.title}</DropdownMenuItem>
                        ))}
                      </DropdownMenuSubContent>
                    </DropdownMenuPortal>
                  </DropdownMenuSub>
                ) : (
                  <DropdownMenuItem
                    render={
                      <Link to={item.url}>
                        <HugeiconsIcon icon={item.icon} strokeWidth={2} />{' '}
                        {item.title}
                      </Link>
                    }
                  />
                )
              )}
            </DropdownMenuGroup>
          </>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
