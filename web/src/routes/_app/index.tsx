import ActivityGraph from '@/components/activity-graph';
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
import Wrapper from '@/components/wrapper';
import { getNoteColorClass } from '@/lib/colors';
import { dummyStories } from '@/lib/dummies';
import { cn } from '@/lib/utils';
import {
  ArrowDown01Icon,
  ArrowRight02Icon,
  Comment02Icon,
  FavouriteIcon,
  Globe02Icon,
  LinkForwardIcon,
  MoreVerticalIcon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute, Link } from '@tanstack/react-router';

const items = [
  { label: '2044', value: 'apple' },
  { label: '2045', value: 'banana' },
  { label: '2046', value: 'blueberry' },
  { label: '2047', value: 'grapes' },
  { label: '2048', value: 'pineapple' },
];

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

export const Route = createFileRoute('/_app/')({
  component: Index,
});

function Index() {
  return (
    <div className="flex gap-2">
      <div className="grow">
        <Wrapper>
          <div className="flex flex-col gap-12">
            <div className="flex items-center">
              <div className="flex gap-2 items-center">
                <Avatar size="xl">
                  <AvatarImage
                    src="https://github.com/shadcn.png"
                    alt="@shadcn"
                  />
                  <AvatarFallback>CN</AvatarFallback>
                </Avatar>
                <div>
                  <p className="font-medium text-2xl/5">Rahmat</p>
                  <p className="text-lg/5 text-muted-foreground">@rahmat</p>
                </div>
              </div>
            </div>
            <div className="flex flex-col gap-4">
              <div className="flex items-center">
                <h2 className="text-xl font-semibold">
                  336k activities over the last year
                </h2>

                <div className="grow flex justify-end">
                  <DropdownMenu>
                    <DropdownMenuTrigger
                      render={<Button variant="secondary" />}
                    >
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

              <ActivityGraph />

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
                          <ItemTitle>Kemana aja?</ItemTitle>
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
                          <ItemTitle>Apa si</ItemTitle>
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
                          <ItemTitle>Ntah kenapa sepertinya</ItemTitle>
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
                    <h3 className="text-lg font-semibold">Activities</h3>

                    {/* <Empty>
                  <EmptyHeader>
                    <EmptyDescription>No activities</EmptyDescription>
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
                                  <ItemDescription>
                                    {song.artist}
                                  </ItemDescription>
                                </ItemContent>
                                <ItemContent className="flex-none text-center">
                                  <ItemDescription>
                                    {song.duration}
                                  </ItemDescription>
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
              <div className="flex items-center">
                <h2 className="text-xl font-semibold">Story Today</h2>

                <div className="grow flex justify-end">
                  <Button variant="ghost">
                    More Stories <HugeiconsIcon icon={ArrowRight02Icon} />
                  </Button>
                </div>
              </div>
              {[dummyStories[0]].map((story) => (
                <Item variant="outline">
                  <ItemContent className="flex flex-col gap-4">
                    <div className="flex items-center">
                      <Link
                        to="/$username"
                        params={{ username: story.user.username }}
                        className="flex gap-2 items-center"
                      >
                        <Avatar size="lg">
                          <AvatarImage
                            src={story.user.imageSrc}
                            alt={story.user.imageAlt}
                          />
                          <AvatarFallback>CN</AvatarFallback>
                        </Avatar>
                        <div>
                          <p className="font-medium text-base/3">
                            {story.user.name}
                          </p>
                          <p>{story.user.username}</p>
                        </div>
                      </Link>
                      <div className="grow flex gap-2 justify-end">
                        <div className="flex gap-1 items-center">
                          <p className="font-bold">{story.date}</p>
                          <p className="text-primary/60 text-lg font-semibold">
                            /
                          </p>
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
                      />
                      <Item className="grow">
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
                          <div className="flex gap-1 -ml-3 mt-4">
                            <Button
                              size="sm"
                              variant="ghost"
                              className="[&_svg]:size-5!"
                            >
                              <HugeiconsIcon
                                strokeWidth={2}
                                icon={FavouriteIcon}
                                className="fill-rose-600 text-rose-600"
                              />
                              423 Likes
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              className="[&_svg]:size-5!"
                            >
                              <HugeiconsIcon
                                strokeWidth={2}
                                icon={Comment02Icon}
                              />
                              65 Comments
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              className="[&_svg]:size-5!"
                            >
                              <HugeiconsIcon
                                strokeWidth={2}
                                icon={LinkForwardIcon}
                              />
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
            <div className="flex flex-col gap-4">
              <h2 className="text-xl font-semibold">Last Stories</h2>

              {[dummyStories[0]].map((story) => (
                <Item variant="outline">
                  <ItemContent className="flex flex-col gap-4">
                    <div className="flex items-center gap-4">
                      <Link
                        to="/$username"
                        params={{ username: story.user.username }}
                        className="flex gap-2 items-center"
                      >
                        <Avatar size="lg">
                          <AvatarImage
                            src={story.user.imageSrc}
                            alt={story.user.imageAlt}
                          />
                          <AvatarFallback>CN</AvatarFallback>
                        </Avatar>
                        <div>
                          <p className="font-medium text-base/4">
                            {story.user.name}
                          </p>
                          <p className="text-base/5 text-muted-foreground">
                            {story.user.username}
                          </p>
                        </div>
                      </Link>
                      <p className="text-2xl font-bold text-muted-foreground">
                        /
                      </p>
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
                          <p className="font-bold">{story.date}</p>
                          <p className="text-primary/60 text-lg font-semibold">
                            /
                          </p>
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
                      />
                      <Item className="grow">
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
                          <div className="flex gap-1 -ml-3 mt-4">
                            <Button
                              size="sm"
                              variant="ghost"
                              className="[&_svg]:size-5!"
                            >
                              <HugeiconsIcon
                                strokeWidth={2}
                                icon={FavouriteIcon}
                              />
                              423 Likes
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              className="[&_svg]:size-5!"
                            >
                              <HugeiconsIcon
                                strokeWidth={2}
                                icon={Comment02Icon}
                              />
                              65 Comments
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              className="[&_svg]:size-5!"
                            >
                              <HugeiconsIcon
                                strokeWidth={2}
                                icon={LinkForwardIcon}
                              />
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
        </Wrapper>
      </div>
    </div>
  );
}
