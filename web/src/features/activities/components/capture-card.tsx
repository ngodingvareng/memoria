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
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { getNoteColorClass } from '@/lib/colors';
import { cn } from '@/lib/utils';
import {
  ArrowRight01Icon,
  Comment02Icon,
  Edit04Icon,
  FavouriteIcon,
  Globe02Icon,
  LinkForwardIcon,
  MoreVerticalIcon,
  PaintBoardIcon,
  Share01Icon,
  SmilePlusIcon,
  StarIcon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';
import { formatDistanceToNow } from 'date-fns';

export interface CaptureCardParam {
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
  capturedAt: Date;
  stats: {
    likes: number;
    comments: number;
    shares: number;
  };
  isPublished?: boolean;
  isOwnedByCurrentUser?: boolean;
}

export function CaptureCard({
  user,
  activity,
  color,
  content,
  tags,
  createdAt,
  capturedAt,
  stats,
  isPublished = false,
  isOwnedByCurrentUser = false,
}: CaptureCardParam) {
  return (
    <Item size="xs">
      <ItemContent className="flex flex-col gap-4">
        <div className="flex group select-text hover:cursor-default selection:bg-primary selection:text-primary-foreground gap-4">
          <Item
            className={cn(
              'max-w-2xs flex flex-none flex-col rounded-xl! w-full overflow-hidden',
              color ? getNoteColorClass(color) : getNoteColorClass('zinc')
            )}
          >
            <ItemContent className="flex flex-col items-end w-full relative z-10">
              <div className="items-end text-right flex flex-col gap-1 grow">
                <div className="flex flex-col">
                  <p className="font-bold text-lg">
                    {capturedAt.toLocaleDateString('id-ID', {
                      day: 'numeric',
                      month: 'long',
                      year: 'numeric',
                    })}
                  </p>
                  <p className="font-medium text-xl/4">
                    {capturedAt.toLocaleTimeString('id-ID', {
                      hour: '2-digit',
                      minute: '2-digit',
                    })}
                  </p>
                </div>
              </div>
              {isOwnedByCurrentUser && (
                <div className="absolute -bottom-1 -left-2 group-hover:opacity-100 opacity-0 transition-opacity">
                  <Tooltip>
                    <TooltipTrigger
                      render={
                        <Button
                          variant="ghost"
                          size="icon-lg"
                          // onClick={() => onFavorite(note.id)}
                        />
                      }
                    >
                      <HugeiconsIcon strokeWidth={2} icon={StarIcon} />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Add to favorites</p>
                    </TooltipContent>
                  </Tooltip>
                  <Tooltip>
                    <TooltipTrigger
                      render={
                        <Button
                          variant="ghost"
                          size="icon"
                          // onClick={() => onSetColor(note.id)}
                        />
                      }
                    >
                      <HugeiconsIcon strokeWidth={2} icon={PaintBoardIcon} />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Set color</p>
                    </TooltipContent>
                  </Tooltip>
                  <Tooltip>
                    <TooltipTrigger
                      render={
                        <Button
                          variant="ghost"
                          size="icon"
                          // onClick={() => onEdit(note.id)}
                        />
                      }
                    >
                      <HugeiconsIcon strokeWidth={2} icon={Edit04Icon} />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Edit content</p>
                    </TooltipContent>
                  </Tooltip>
                  <Tooltip>
                    <TooltipTrigger
                      render={
                        <Button
                          variant="ghost"
                          size="icon"
                          // onClick={() => onShare(note.id)}
                        />
                      }
                    >
                      <HugeiconsIcon strokeWidth={2} icon={Share01Icon} />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Share</p>
                    </TooltipContent>
                  </Tooltip>
                </div>
              )}
            </ItemContent>
          </Item>
          <Item className="grow pt-1 pb-0">
            {isPublished && (
              <ItemHeader>
                <Link
                  to="/$username"
                  params={{ username: user.username }}
                  className="flex gap-2 items-center"
                >
                  <Avatar>
                    <AvatarImage src={user.imageSrc} alt={user.imageAlt} />
                    <AvatarFallback>CN</AvatarFallback>
                  </Avatar>
                  <div>
                    <p className="text-base">{user.username}</p>
                  </div>
                </Link>
                <HugeiconsIcon
                  icon={ArrowRight01Icon}
                  className="size-4 font-bold text-muted-foreground"
                />
                <div className="font-medium">
                  <p>{activity.name}</p>
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
                    <p className="font-medium text-muted-foreground">
                      {formatDistanceToNow(createdAt, {
                        addSuffix: true,
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
            )}
            <ItemContent className="typeset max-w-none">
              {content}
              <div className="flex flex-wrap gap-x-2">
                {tags
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
            {isPublished && (
              <ItemFooter className="flex gap-1 -ml-3 justify-start">
                <Button size="icon" variant="ghost" className="[&_svg]:size-5!">
                  <HugeiconsIcon strokeWidth={2} icon={SmilePlusIcon} />
                </Button>
                <Button size="sm" variant="ghost" className="[&_svg]:size-5!">
                  <HugeiconsIcon strokeWidth={2} icon={Comment02Icon} />
                  {stats.comments}
                </Button>
              </ItemFooter>
            )}
          </Item>
        </div>
      </ItemContent>
    </Item>
  );
}
