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
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemMedia,
  ItemSeparator,
  ItemTitle,
} from '@/components/ui/item';
import { MomentHeatmap } from '@/features/moments';
import { getNoteColorClass } from '@/lib/colors';
import { dummyStories } from '@/lib/dummies';
import { cn } from '@/lib/utils';
import {
  ArrowDown01Icon,
  Comment02Icon,
  FavouriteIcon,
  Globe02Icon,
  LinkForwardIcon,
  MoreVerticalIcon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute, Link } from '@tanstack/react-router';

const music = [
  {
    title: 'Midnight City Lights',
    artist: 'Neon Dreams',
    album: 'Electric Nights',
    duration: '3:45',
  },
  {
    title: 'Coffee Shop Conversations',
    artist: 'The Morning Brew',
    album: 'Urban Stories',
    duration: '4:05',
  },
  {
    title: 'Digital Rain',
    artist: 'Cyber Symphony',
    album: 'Binary Beats',
    duration: '3:30',
  },
];
export const Route = createFileRoute('/_app/_circle/c/$id/')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="flex flex-col gap-12">
      <div className="flex flex-col gap-4">
        <div className="flex items-center">
          <h2 className="text-xl font-semibold">
            336k threads over the last year
          </h2>

          <div className="grow flex justify-end">
            <DropdownMenu>
              <DropdownMenuTrigger render={<Button variant="secondary" />}>
                <span className="font-semibold text-muted-foreground">
                  Year
                </span>{' '}
                2077
                <HugeiconsIcon icon={ArrowDown01Icon} strokeWidth={2} />
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuGroup>
                  <DropdownMenuItem>2077</DropdownMenuItem>
                  <DropdownMenuItem>2078</DropdownMenuItem>
                  <DropdownMenuItem>2079</DropdownMenuItem>
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        <MomentHeatmap />

        <div className="flex gap-4">
          <Item variant="outline" className="items-start">
            <ItemContent className="flex flex-col gap-4">
              <h3 className="text-lg font-semibold">
                Confirmations{' '}
                <Badge variant="secondary" className="text-base">
                  19
                </Badge>
              </h3>

              {/* <Empty>
                    <EmptyHeader>
                      <EmptyDescription>No confirmations</EmptyDescription>
                    </EmptyHeader>
                  </Empty> */}

              <ItemGroup>
                <Item size="xs">
                  <ItemContent>
                    <ItemTitle>Where'd you go?</ItemTitle>
                    <ItemDescription className="flex gap-1 items-center">
                      <span>10:30 until 11:00</span>
                      <Badge className="bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300">
                        Awaiting
                      </Badge>
                    </ItemDescription>
                  </ItemContent>
                  <ItemActions>
                    <Button size="sm">Confirm</Button>
                  </ItemActions>
                </Item>
                <ItemSeparator />
                <Item size="xs">
                  <ItemContent>
                    <ItemTitle>What is this</ItemTitle>
                    <ItemDescription className="flex gap-1 items-center">
                      <span>23:00</span>
                      <Badge className="bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300">
                        Overdue
                      </Badge>
                    </ItemDescription>
                  </ItemContent>
                  <ItemActions>
                    <Button size="sm">Confirm</Button>
                  </ItemActions>
                </Item>
                <ItemSeparator />
                <Item size="xs">
                  <ItemContent>
                    <ItemTitle>Not sure why but it seems</ItemTitle>
                    <ItemDescription className="flex gap-1 items-center">
                      <span>23:00</span>
                      <Badge variant="outline">Upcoming</Badge>
                    </ItemDescription>
                  </ItemContent>
                </Item>
              </ItemGroup>
            </ItemContent>
          </Item>

          <Item variant="outline" className="items-start">
            <ItemContent className="flex flex-col gap-4">
              <h3 className="text-lg font-semibold">Threads</h3>

              {/* <Empty>
                    <EmptyHeader>
                      <EmptyDescription>No threads</EmptyDescription>
                    </EmptyHeader>
                  </Empty> */}

              <ItemGroup>
                {music.map((song) => (
                  <>
                    <Item
                      size="sm"
                      key={song.title}
                      role="listitem"
                      render={
                        <a href="#">
                          <ItemMedia variant="image">
                            <img
                              src={`https://avatar.vercel.sh/${song.title}`}
                              alt={song.title}
                              width={32}
                              height={32}
                              className="object-cover grayscale"
                            />
                          </ItemMedia>
                          <ItemContent>
                            <ItemTitle className="line-clamp-1">
                              {song.title} -{' '}
                              <span className="text-muted-foreground">
                                {song.album}
                              </span>
                            </ItemTitle>
                            <ItemDescription>{song.artist}</ItemDescription>
                          </ItemContent>
                          <ItemContent className="flex-none text-center">
                            <ItemDescription>{song.duration}</ItemDescription>
                          </ItemContent>
                        </a>
                      }
                    />
                    {music[music.length - 1].title != song.title && (
                      <ItemSeparator />
                    )}
                  </>
                ))}
              </ItemGroup>
            </ItemContent>
          </Item>
        </div>
      </div>
      <div className="flex flex-col gap-4">
        <h2 className="text-xl font-semibold">Last Stories</h2>

        {[dummyStories[0]].map((moment) => (
          <Item variant="outline">
            <ItemContent className="flex flex-col gap-4">
              <div className="flex items-center gap-4">
                <Link
                  to="/$username"
                  params={{ username: moment.user.username }}
                  className="flex gap-2 items-center"
                >
                  <Avatar size="lg">
                    <AvatarImage
                      src={moment.user.imageSrc}
                      alt={moment.user.imageAlt}
                    />
                    <AvatarFallback>CN</AvatarFallback>
                  </Avatar>
                  <div>
                    <p className="font-medium text-base/4">
                      {moment.user.name}
                    </p>
                    <p className="text-base/5 text-muted-foreground">
                      {moment.user.username}
                    </p>
                  </div>
                </Link>
                <p className="text-2xl font-bold text-muted-foreground">/</p>
                <div className="font-semibold text-lg">
                  <p>He is not here anymore...</p>
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
                    <p className="font-bold">{moment.date}</p>
                    <p className="text-primary/60 text-lg font-semibold">/</p>
                    <p className="font-medium">{moment.time}</p>
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
                    moment.color
                      ? getNoteColorClass(moment.color)
                      : getNoteColorClass('zinc')
                  )}
                />
                <Item className="grow">
                  <ItemContent className="typeset max-w-none">
                    {moment.content}
                    <div className="flex flex-wrap gap-x-2">
                      {moment.tags
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
                    <div className="flex gap-1 -ml-3 mt-4">
                      <Button
                        size="sm"
                        variant="ghost"
                        className="[&_svg]:size-5!"
                      >
                        <HugeiconsIcon strokeWidth={2} icon={FavouriteIcon} />
                        423 Likes
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        className="[&_svg]:size-5!"
                      >
                        <HugeiconsIcon strokeWidth={2} icon={Comment02Icon} />
                        65 Comments
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        className="[&_svg]:size-5!"
                      >
                        <HugeiconsIcon strokeWidth={2} icon={LinkForwardIcon} />
                        64k Share
                      </Button>
                    </div>
                  </ItemContent>
                </Item>
              </div>
            </ItemContent>
          </Item>
        ))}
      </div>
    </div>
  );
}
