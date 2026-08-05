import { Fragment } from 'react';
import {
  HelpCircleIcon,
  LanguageSquareIcon,
  Logout01Icon,
  Settings02Icon,
  Sun01Icon,
  Tick02Icon,
  UserCircleIcon,
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
} from '@/components/ui/dropdown-menu';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';
import { useTheme, type Theme } from '@/components/theme-provider';

export function UserHeaderMenu() {
  const { theme, setTheme } = useTheme();

  const data = {
    user: [
      [
        {
          title: 'Profile',
          icon: UserCircleIcon,
          url: '/@shadcn',
        },
      ],
      [
        {
          title: 'Display Language: English',
          icon: LanguageSquareIcon,
          submenu: [
            { title: 'English', value: 'english' },
            { title: 'Indonesia', value: 'indonesia' },
          ],
        },
        {
          title: `Appearance: ${theme}`,
          icon: Sun01Icon,
          submenu: [
            { title: 'Light theme', value: 'light' },
            { title: 'Dark theme', value: 'dark' },
            { title: 'Use device theme', value: 'system' },
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
        {data.user.map((group, groupIndex) => (
          <Fragment key={groupIndex}>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              {group.map((item) =>
                item.submenu ? (
                  <DropdownMenuSub key={item.title}>
                    <DropdownMenuSubTrigger>
                      <HugeiconsIcon icon={item.icon} strokeWidth={2} />{' '}
                      {item.title}
                    </DropdownMenuSubTrigger>
                    <DropdownMenuPortal>
                      <DropdownMenuSubContent>
                        {item.submenu.map((subitem) => (
                          <DropdownMenuItem
                            key={subitem.title}
                            onClick={
                              subitem.value
                                ? () => setTheme(subitem.value as Theme)
                                : undefined
                            }
                            className="flex items-center justify-between"
                          >
                            {subitem.title}
                            {subitem.value === theme && (
                              <HugeiconsIcon icon={Tick02Icon} />
                            )}
                          </DropdownMenuItem>
                        ))}
                      </DropdownMenuSubContent>
                    </DropdownMenuPortal>
                  </DropdownMenuSub>
                ) : (
                  <DropdownMenuItem
                    key={item.title}
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
          </Fragment>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
