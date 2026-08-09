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
import { Link, useNavigate } from '@tanstack/react-router';
import { useTheme, type Theme } from '@/components/theme-provider';
import { useSession, setSession } from '@/lib/session';
import { useLogout } from '@/lib/api/generated/auth/auth';

type MenuGroup = {
  title: string;
  icon: Parameters<typeof HugeiconsIcon>[0]['icon'];
  url?: string;
  submenu?: { title: string; value: string }[];
}[];

function MenuGroupItems({
  group,
  theme,
  setTheme,
}: {
  group: MenuGroup;
  theme: string;
  setTheme: (theme: Theme) => void;
}) {
  return (
    <>
      {group.map((item) =>
        item.submenu ? (
          <DropdownMenuSub key={item.title}>
            <DropdownMenuSubTrigger>
              <HugeiconsIcon icon={item.icon} strokeWidth={2} /> {item.title}
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
              <Link to={item.url!}>
                <HugeiconsIcon icon={item.icon} strokeWidth={2} /> {item.title}
              </Link>
            }
          />
        )
      )}
    </>
  );
}

export function UserHeaderMenu() {
  const { theme, setTheme } = useTheme();
  const session = useSession();
  const navigate = useNavigate();
  const logoutMutation = useLogout();

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
        url: session?.user.username ? `/${session.user.username}` : '/user',
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
              src="https://github.com/shadcn.png"
              alt={session?.user.name}
              className="grayscale"
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
                src="https://github.com/shadcn.png"
                alt={session?.user.name}
                className="grayscale"
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
