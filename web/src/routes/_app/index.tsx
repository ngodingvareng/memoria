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
import {
  ActivityGraph,
  CaptureList,
  type CaptureCardParam,
} from '@/features/activities';

import { getNoteColorClass } from '@/lib/colors';
import { dummyStories } from '@/lib/dummies';
import { cn } from '@/lib/utils';
import {
  ArrowDown01Icon,
  ArrowRight02Icon,
  Comment02Icon,
  FavouriteIcon,
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

export const dummyActivityStories: CaptureCardParam[] = [
  {
    user: {
      name: 'Budi Santoso',
      username: '@budisans',
      imageSrc: 'https://randomuser.me/api/portraits/men/32.jpg',
      imageAlt: 'Foto profil Budi Santoso',
    },
    activity: {
      name: 'Lari Pagi 5KM',
    },
    color: 'orange', // Oranye
    content: (
      <div>
        <p>
          Memulai hari dengan lari sejauh 5km di sekitar kawasan GBK. Cuaca hari
          ini sangat cerah dan udaranya segar! 🏃‍♂️☀️
        </p>
      </div>
    ),
    tags: ['running', 'health', 'morningvibes'],
    createdAt: new Date('2026-07-26T06:30:00'),
    capturedAt: new Date('2026-07-26T06:30:00'),
    stats: {
      likes: 124,
      comments: 12,
      shares: 3,
    },
  },
  {
    user: {
      name: 'Siti Aminah',
      username: '@sitiamin',
      imageSrc: 'https://randomuser.me/api/portraits/women/44.jpg',
      imageAlt: 'Foto profil Siti Aminah',
    },
    activity: {
      name: 'Eksplorasi Kuliner',
    },
    color: 'yellow', // Kuning
    content: (
      <div>
        <p>
          Mencoba kedai kopi baru yang lagi viral di Jakarta Selatan. Kopinya
          enak dan suasananya sangat <i>cozy</i> untuk nugas atau WFC. ☕🥐
        </p>
        <img
          src="https://images.unsplash.com/photo-1554118811-1e0d58224f24?ixlib=rb-1.2.1&auto=format&fit=crop&w=500&q=60"
          alt="Kopi dan Croissant"
          style={{ width: '100%', borderRadius: '8px', marginTop: '10px' }}
        />
      </div>
    ),
    tags: ['kuliner', 'coffee', 'weekend', 'jaksel'],
    createdAt: new Date('2026-07-25T15:45:00'),
    capturedAt: new Date('2026-07-25T15:45:00'),
    stats: {
      likes: 342,
      comments: 45,
      shares: 18,
    },
  },
  {
    user: {
      name: 'Reza Rahadian',
      username: '@rezadev',
      imageSrc: 'https://randomuser.me/api/portraits/men/75.jpg',
      imageAlt: 'Foto profil Reza Rahadian',
    },
    activity: {
      name: 'Live Coding',
    },
    color: 'blue', // Biru
    content: (
      <div>
        <p>
          Akhirnya berhasil menyelesaikan *bug* yang sudah mengganggu selama 3
          hari. <b>React hooks</b> memang luar biasa tapi kadang bikin pusing!
          💻🚀
        </p>
      </div>
    ),
    tags: ['coding', 'reactjs', 'webdev', 'programming'],
    createdAt: new Date('2026-07-26T20:15:00'),
    capturedAt: new Date('2026-07-26T20:15:00'),
    stats: {
      likes: 89,
      comments: 8,
      shares: 2,
    },
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
              <CaptureList
                captures={[dummyActivityStories[0]]}
                withStoryLayout
              />
            </div>
            <div className="flex flex-col gap-4">
              <h2 className="text-xl font-semibold">Last Stories</h2>

              <CaptureList captures={dummyActivityStories} withStoryLayout />
            </div>
          </div>
        </Wrapper>
      </div>
    </div>
  );
}
