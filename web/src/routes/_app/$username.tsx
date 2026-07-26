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
import Wrapper from '@/components/wrapper';
import { getNoteColorClass } from '@/lib/colors';
import { dummyStories } from '@/lib/dummies';
import { cn } from '@/lib/utils';
import {
  Comment02Icon,
  FavouriteIcon,
  MoreVerticalIcon,
  Share01Icon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute, Link } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/$username')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <Wrapper>
      <div className="relative">
        <div className="h-40 rounded-4xl  overflow-hidden">
          <div className="bg-primary absolute inset-0 z-30 aspect-video opacity-50 mix-blend-color" />
          <img
            width={400}
            height={400}
            src="https://images.unsplash.com/photo-1610280777472-54133d004c8c?q=80&w=640&auto=format&fit=crop"
            className="relative z-20 aspect-video w-full object-cover brightness-60 grayscale"
          />
        </div>
      </div>
      <div className="flex flex-col gap-6 pt-6">
        <div className="flex gap-4">
          <Avatar className="size-20">
            <AvatarImage src="https://github.com/shadcn.png" alt="@shadcn" />
            <AvatarFallback>CN</AvatarFallback>
          </Avatar>
          <div className="flex flex-col gap-1">
            <div>
              <p className="font-medium text-2xl/5">zaraa</p>
              <p className="text-lg/5 text-muted-foreground">@safaza_</p>
            </div>

            <div className="flex gap-4">
              <p>
                <span className="font-semibold">1553</span>{' '}
                <span className="text-muted-foreground">followers</span>
              </p>
              <p>
                <span className="font-semibold">954</span>{' '}
                <span className="text-muted-foreground">following</span>
              </p>
              <p>
                <span className="font-semibold">6.7M</span>{' '}
                <span className="text-muted-foreground">activities</span>
              </p>
            </div>

            <div>
              <Button>Follow</Button>
            </div>
          </div>
        </div>

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
                ></Item>
                <Item className="grow hover:bg-muted/50">
                  <ItemContent className="typeset max-w-none">
                    {story.content}
                    <div className="flex flex-wrap gap-x-2">
                      {story.tags
                        .map((tag) => `#${tag}`)
                        .map((tag) => (
                          <p>
                            <a
                              key={tag}
                              className="hover:underline font-medium"
                            >
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
    </Wrapper>
  );
}
