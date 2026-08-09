import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import Wrapper from '@/components/wrapper';
import { MomentList, type MomentCardParam } from '@/features/moments';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/$username')({
  component: RouteComponent,
});

export const dummyThreadStories: MomentCardParam[] = [
  {
    user: {
      name: 'Ahmad Faisal',
      username: '@ahmad_f',
      imageSrc: 'https://randomuser.me/api/portraits/men/45.jpg',
      imageAlt: 'Profile photo of Ahmad Faisal',
    },
    thread: {
      name: 'Hiking Mount Rinjani',
    },
    color: 'emerald', // Tailwind color
    content: (
      <div className="flex flex-col gap-3">
        <p>
          The long, exhausting trek paid off in full watching the sunrise from
          Rinjani's summit. Truly an unforgettable, magical experience! 🏔️✨
        </p>
        <img
          src="https://images.unsplash.com/photo-1596404835697-3d964fdd9d49?ixlib=rb-1.2.1&auto=format&fit=crop&w=600&q=80"
          alt="Mount Rinjani summit"
          className="w-full rounded-xl object-cover h-64 shadow-sm"
        />
      </div>
    ),
    tags: ['hiking', 'rinjani', 'nature', 'adventure'],
    createdAt: new Date('2026-07-24T05:15:00'),
    capturedAt: new Date('2026-07-24T05:15:00'),
    stats: {
      likes: 1250,
      comments: 89,
      shares: 45,
    },
  },
  {
    user: {
      name: 'Nadia Putri',
      username: '@nadiasings',
      imageSrc: 'https://randomuser.me/api/portraits/women/68.jpg',
      imageAlt: 'Profile photo of Nadia Putri',
    },
    thread: {
      name: 'Concert Night',
    },
    color: 'purple', // Tailwind color
    content: (
      <div>
        <p>
          Tonight was insane! Got a spot right at the front and sang along from
          start to finish. My voice is completely gone but{' '}
          <i>totally worth it!</i> 🎤🎶
        </p>
      </div>
    ),
    tags: ['concert', 'music', 'live', 'jakarta'],
    createdAt: new Date('2026-07-25T22:30:00'),
    capturedAt: new Date('2026-07-25T22:30:00'),
    stats: {
      likes: 843,
      comments: 112,
      shares: 20,
    },
  },
  {
    user: {
      name: 'Chef Dimas',
      username: '@dimasmasak',
      imageSrc: 'https://randomuser.me/api/portraits/men/22.jpg',
      imageAlt: 'Profile photo of Chef Dimas',
    },
    thread: {
      name: 'New Recipe Experiment',
    },
    color: 'orange', // Tailwind color
    content: (
      <div className="space-y-2">
        <p>
          Tried making <b>Beef Wellington</b> for the first time today. The{' '}
          <i>pastry</i> came out crisp and the meat was a perfect medium rare!
          🥩🔥 Anyone want the recipe?
        </p>
        <img
          src="https://images.unsplash.com/photo-1600891964092-4316c288032e?ixlib=rb-1.2.1&auto=format&fit=crop&w=600&q=80"
          alt="Steak"
          className="w-full rounded-lg object-cover h-48"
        />
      </div>
    ),
    tags: ['cooking', 'chef', 'beefwellington', 'foodie'],
    createdAt: new Date('2026-07-26T12:00:00'),
    capturedAt: new Date('2026-07-26T12:00:00'),
    stats: {
      likes: 567,
      comments: 230,
      shares: 88,
    },
  },
  {
    user: {
      name: 'Kevin Pratama',
      username: '@kevinzoro',
      imageSrc: 'https://randomuser.me/api/portraits/men/85.jpg',
      imageAlt: 'Profile photo of Kevin',
    },
    thread: {
      name: 'Valorant Tournament',
    },
    color: 'rose', // Tailwind color
    content: (
      <div>
        <p>
          A <i>Clutch 1v5</i> moment that secured our team's spot in tomorrow's
          Final! Hands still shaking but GGWP to every team that played today.
          🎮🏆
        </p>
      </div>
    ),
    tags: ['valorant', 'esports', 'gaming', 'clutch'],
    createdAt: new Date('2026-07-26T19:45:00'),
    capturedAt: new Date('2026-07-26T19:45:00'),
    stats: {
      likes: 2100,
      comments: 340,
      shares: 156,
    },
  },
  {
    user: {
      name: 'Maya Sari',
      username: '@mayadesign',
      imageSrc: 'https://randomuser.me/api/portraits/women/12.jpg',
      imageAlt: 'Profile photo of Maya Sari',
    },
    thread: {
      name: 'UI/UX Redesign Wrapped',
    },
    color: 'cyan', // Tailwind color
    content: (
      <div>
        <p>
          Finally finished the fintech app <i>redesign</i> project and handed it
          over to the dev team. Really happy with how the color palette and
          interactions turned out. Time to rest! 💻🎨
        </p>
      </div>
    ),
    tags: ['uiux', 'design', 'figma', 'freelance'],
    createdAt: new Date('2026-07-26T16:20:00'),
    capturedAt: new Date('2026-07-26T16:20:00'),
    stats: {
      likes: 312,
      comments: 45,
      shares: 12,
    },
  },
  {
    user: {
      name: 'Sinta Wijaya',
      username: '@sintayoga',
      imageSrc: 'https://randomuser.me/api/portraits/women/33.jpg',
      imageAlt: 'Profile photo of Sinta Wijaya',
    },
    thread: {
      name: 'Evening Yoga Session',
    },
    color: 'teal', // Tailwind color
    content: (
      <div className="flex flex-col gap-2">
        <p>
          Closing out the weekend with a Vinyasa Yoga session. Really helped
          clear my mind before facing Monday tomorrow. Namaste 🙏🧘‍♀️
        </p>
      </div>
    ),
    tags: ['yoga', 'mindfulness', 'health', 'weekend'],
    createdAt: new Date('2026-07-26T17:30:00'),
    capturedAt: new Date('2026-07-26T17:30:00'),
    stats: {
      likes: 420,
      comments: 18,
      shares: 5,
    },
  },
];

function RouteComponent() {
  return (
    <Wrapper>
      <div className="flex flex-col gap-12">
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
                <span className="text-muted-foreground">threads</span>
              </p>
            </div>

            <div>
              <p>None knows me</p>
            </div>

            <div>
              <Button>Follow</Button>
            </div>
          </div>
        </div>

        <div className="flex flex-col gap-2">
          <MomentList moments={dummyThreadStories} withStoryLayout />
        </div>
      </div>
    </Wrapper>
  );
}
