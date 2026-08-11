import {
  DropdownMenuItem,
  DropdownMenuPortal,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
} from '@/components/ui/dropdown-menu';
import { Tick02Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';
import type { Theme } from '@/components/theme-provider';

export type MenuGroup = {
  title: string;
  icon: Parameters<typeof HugeiconsIcon>[0]['icon'];
  url?: string;
  submenu?: { title: string; value: string }[];
}[];

interface MenuGroupItemsProps {
  group: MenuGroup;
  theme: string;
  setTheme: (theme: Theme) => void;
}

export function MenuGroupItems({
  group,
  theme,
  setTheme,
}: MenuGroupItemsProps) {
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
          <DropdownMenuItem key={item.title} render={<Link to={item.url!} />}>
            <HugeiconsIcon icon={item.icon} strokeWidth={2} /> {item.title}
          </DropdownMenuItem>
        )
      )}
    </>
  );
}
