import { Fragment } from 'react';
import {
  HelpCircleIcon,
  LanguageSquareIcon,
  Logout01Icon,
  Settings02Icon,
  Sun01Icon,
  UserCircleIcon,
} from '@hugeicons/core-free-icons';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { HugeiconsIcon } from '@hugeicons/react';
import { useNavigate } from '@tanstack/react-router';
import { useTheme } from '@/components/theme-provider';
import { useSession, setSession } from '@/lib/session';
import { useLogout } from '@/lib/api/generated/auth/auth';
import { useGetUserByID } from '@/lib/api/generated/users/users';
import { MenuGroupItems, type MenuGroup } from './menu-group-items';

export function UserHeaderMenu() {
  const { theme, setTheme } = useTheme();
  const session = useSession();
  const navigate = useNavigate();
  const logoutMutation = useLogout();
  const userQuery = useGetUserByID(session?.user.id ?? '', {
    query: { enabled: !!session?.user.id },
  });

  const handleSignOut = async () => {
    try {
      await logoutMutation.mutateAsync();
    } finally {
      // The refresh cookie is cleared server-side either way — drop the
      // local session and send the user back to /signin regardless of
      // whether the request itself succeeded.
      setSession(null);
      navigate({ to: '/signin' });
    }
  };

  const groupsBeforeSignOut: MenuGroup[] = [
    [
      {
        title: 'Profile',
        icon: UserCircleIcon,
        url: session?.user.username ? `/@${session.user.username}` : '/user',
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
  ];

  const groupsAfterSignOut: MenuGroup[] = [
    [
      {
        title: 'Help',
        icon: HelpCircleIcon,
        url: '/help',
      },
    ],
  ];

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Avatar size="lg">
            <AvatarImage
              src={userQuery.data?.image_path ?? undefined}
              alt={session?.user.name}
            />
            <AvatarFallback>
              {session?.user.name.slice(0, 2).toUpperCase()}
            </AvatarFallback>
          </Avatar>
        }
      />
      <DropdownMenuContent className="w-72">
        <DropdownMenuGroup>
          <div className="flex gap-3 items-center px-3 py-1">
            <Avatar>
              <AvatarImage
                src={userQuery.data?.image_path ?? undefined}
                alt={session?.user.name}
              />
              <AvatarFallback>
                {session?.user.name.slice(0, 2).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <div>
              <p className="font-medium text-base/3">{session?.user.name}</p>
              <p>{session?.user.username ? `@${session.user.username}` : ''}</p>
            </div>
          </div>
        </DropdownMenuGroup>
        {groupsBeforeSignOut.map((group, groupIndex) => (
          <Fragment key={groupIndex}>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <MenuGroupItems group={group} theme={theme} setTheme={setTheme} />
            </DropdownMenuGroup>
          </Fragment>
        ))}
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuItem onClick={handleSignOut}>
            <HugeiconsIcon icon={Logout01Icon} strokeWidth={2} /> Sign Out
          </DropdownMenuItem>
        </DropdownMenuGroup>
        {groupsAfterSignOut.map((group, groupIndex) => (
          <Fragment key={groupIndex}>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <MenuGroupItems group={group} theme={theme} setTheme={setTheme} />
            </DropdownMenuGroup>
          </Fragment>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
