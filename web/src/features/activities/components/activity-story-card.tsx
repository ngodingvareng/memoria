import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Item,
  ItemContent,
  ItemFooter,
  ItemHeader,
} from '@/components/ui/item';
import { getNoteColorClass } from '@/lib/colors';
import { cn } from '@/lib/utils';
import {
  ArrowRight01Icon,
  Comment02Icon,
  FavouriteIcon,
  Globe02Icon,
  LinkForwardIcon,
  MoreVerticalIcon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';

export interface ActivityStoryParam {
  capture: {
    user: {
      name: string;
      username: string;
      imageSrc: string;
      imageAlt: string;
    };
    activity: {
      name: string;
    };
    color: string;
    content: React.ReactNode;
    tags: string[];
    createdAt: Date;
    stats: {
      likes: number;
      comments: number;
      shares: number;
    };
  };
}

export function ActivityStoryCard({ capture }: ActivityStoryParam) {
  return (
    <Item variant="outline">
      <ItemContent className="flex flex-col gap-4">
        <div className="flex group select-text hover:cursor-default selection:bg-primary selection:text-primary-foreground gap-4">
          <Item
            className={cn(
              'max-w-2xs flex flex-col rounded-xl! w-full overflow-hidden',
              capture.color
                ? getNoteColorClass(capture.color)
                : getNoteColorClass('zinc')
            )}
          />
          <Item className="grow">
            <ItemHeader>
              <Link
                to="/$username"
                params={{ username: capture.user.username }}
                className="flex gap-2 items-center"
              >
                <Avatar>
                  <AvatarImage
                    src={capture.user.imageSrc}
                    alt={capture.user.imageAlt}
                  />
                  <AvatarFallback>CN</AvatarFallback>
                </Avatar>
                <div>
                  <p className="text-base">{capture.user.username}</p>
                </div>
              </Link>
              <HugeiconsIcon
                icon={ArrowRight01Icon}
                className="size-4 font-bold text-muted-foreground"
              />
              <div className="font-medium">
                <p>{capture.activity.name}</p>
              </div>
              <div className="grow items-center flex gap-2 justify-end">
                <div>
                  <Badge className="bg-sky-50 text-sky-700 dark:bg-sky-950 dark:text-sky-300">
                    <HugeiconsIcon
                      data-icon="inline-start"
                      icon={Globe02Icon}
                    />
                    Public
                  </Badge>
                </div>
                <div className="flex gap-1 items-center">
                  <p className="font-bold">
                    {capture.createdAt.toLocaleDateString('id-ID', {
                      day: 'numeric',
                      month: 'long',
                      year: 'numeric',
                    })}
                  </p>
                  <p className="text-primary/60 text-lg font-semibold">/</p>
                  <p className="font-medium">
                    {capture.createdAt.toLocaleTimeString('id-ID', {
                      hour: '2-digit',
                      minute: '2-digit',
                    })}
                  </p>
                </div>
                <DropdownMenu>
                  <DropdownMenuTrigger
                    render={
                      <Button size="icon" variant="ghost" className="-mr-3">
                        <HugeiconsIcon
                          strokeWidth={2}
                          icon={MoreVerticalIcon}
                        />
                      </Button>
                    }
                  />
                  <DropdownMenuContent className="w-40" align="start">
                    <DropdownMenuGroup>
                      <DropdownMenuItem>Profile</DropdownMenuItem>
                      <DropdownMenuItem>Billing</DropdownMenuItem>
                      <DropdownMenuItem>Settings</DropdownMenuItem>
                    </DropdownMenuGroup>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </ItemHeader>
            <ItemContent className="typeset max-w-none">
              {capture.content}
              <div className="flex flex-wrap gap-x-2">
                {capture.tags
                  .map((tag) => `#${tag}`)
                  .map((tag) => (
                    <p>
                      <a key={tag} className="hover:underline font-medium">
                        {tag}
                      </a>
                    </p>
                  ))}
              </div>
            </ItemContent>
            <ItemFooter className="flex gap-1 -ml-3 justify-start">
              <Button size="sm" variant="ghost" className="[&_svg]:size-5!">
                <HugeiconsIcon strokeWidth={2} icon={FavouriteIcon} />
                {capture.stats.likes} Likes
              </Button>
              <Button size="sm" variant="ghost" className="[&_svg]:size-5!">
                <HugeiconsIcon strokeWidth={2} icon={Comment02Icon} />
                {capture.stats.comments} Comments
              </Button>
              <Button size="sm" variant="ghost" className="[&_svg]:size-5!">
                <HugeiconsIcon strokeWidth={2} icon={LinkForwardIcon} />
                {capture.stats.shares} Share
              </Button>
            </ItemFooter>
          </Item>
        </div>
      </ItemContent>
    </Item>
  );
}
