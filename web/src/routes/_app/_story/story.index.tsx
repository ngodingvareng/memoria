import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Item, ItemContent } from '@/components/ui/item';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { getNoteColorClass } from '@/lib/colors';
import { dummyStories } from '@/lib/dummies';
import { cn } from '@/lib/utils';
import {
  Comment02Icon,
  Edit04Icon,
  FavouriteIcon,
  MoreVerticalIcon,
  PaintBoardIcon,
  Share01Icon,
  StarIcon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute, Link } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_story/story/')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="flex flex-col gap-4">
      {dummyStories.map((story) => (
        <Item variant="outline">
          <ItemContent className="flex flex-col gap-4">
            <div className="flex items-center">
              <Link
                to="/$username"
                params={{ username: story.user.username }}
                className="flex gap-2 items-center"
              >
                <Avatar>
                  <AvatarImage
                    src={story.user.imageSrc}
                    alt={story.user.imageAlt}
                  />
                  <AvatarFallback>CN</AvatarFallback>
                </Avatar>
                <div>
                  <p className="font-medium text-base/3">{story.user.name}</p>
                  <p>{story.user.username}</p>
                </div>
              </Link>
              <div className="grow flex gap-2 justify-end">
                <div className="flex gap-1 items-center">
                  <p className="font-bold">{story.date}</p>
                  <p className="text-primary/60 text-lg font-semibold">/</p>
                  <p className="font-medium">{story.time}</p>
                </div>
                <DropdownMenu>
                  <DropdownMenuTrigger
                    render={
                      <Button size="icon" variant="ghost">
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
            </div>

            <div className="flex group select-text hover:cursor-default selection:bg-primary selection:text-primary-foreground gap-4">
              <Item
                className={cn(
                  'max-w-2xs flex flex-col w-full overflow-hidden',
                  story.color
                    ? getNoteColorClass(story.color)
                    : getNoteColorClass('zinc')
                )}
              >
                <ItemContent className="flex flex-col items-end w-full relative z-10">
                  <div className="absolute -bottom-1 left-0 group-hover:opacity-100 opacity-0 transition-opacity">
                    <Tooltip>
                      <TooltipTrigger
                        render={
                          <Button
                            variant="ghost"
                            size="icon-lg"
                            // onClick={() => onFavorite(story.id)}
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
                            // onClick={() => onSetColor(story.id)}
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
                            // onClick={() => onEdit(story.id)}
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
                            // onClick={() => onShare(story.id)}
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
                </ItemContent>
              </Item>
              <Item className="grow hover:bg-muted/50">
                <ItemContent className="typeset max-w-none">
                  {story.content}
                  <div className="flex flex-wrap gap-x-2">
                    {story.tags
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
              </Item>
            </div>
            <div className="flex gap-1">
              <Button size="sm" variant="ghost">
                <HugeiconsIcon strokeWidth={2} icon={FavouriteIcon} />
                423 Likes
              </Button>
              <Button size="sm" variant="ghost">
                <HugeiconsIcon strokeWidth={2} icon={Comment02Icon} />
                65 Comments
              </Button>
              <Button size="sm" variant="ghost">
                <HugeiconsIcon strokeWidth={2} icon={Share01Icon} />
                64k Share
              </Button>
            </div>
          </ItemContent>
        </Item>
      ))}
    </div>
  );
}
