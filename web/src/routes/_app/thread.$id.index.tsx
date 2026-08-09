import { ColorPickerDialog } from '@/components/dialogs/color-picker-dialog';
import { DateTimeDialog } from '@/components/dialogs/datetime-dialog';
import { ShareDialog } from '@/components/dialogs/share-dialog';

import Wrapper from '@/components/wrapper';
import {
  MomentInput,
  MomentList,
  type MomentCardParam,
} from '@/features/moments';
import { ThreadHeader, ThreadHero } from '@/features/threads';
import { dummyNotes } from '@/lib/dummies';
import {
  useGetThreadsId,
  useGetThreadsIdImages,
} from '@/lib/api/generated/threads/threads';
import { createFileRoute } from '@tanstack/react-router';
import React from 'react';

export const Route = createFileRoute('/_app/thread/$id/')({
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
  const { id } = Route.useParams();
  const threadQuery = useGetThreadsId(id);
  const imagesQuery = useGetThreadsIdImages(id);
  const [openCostumizationDialog, setOpenCostumizationDialog] =
    React.useState(false);
  const [openShareDialog, setOpenShareDialog] = React.useState(false);
  const [openTimeDialog, setOpenTimeDialog] = React.useState(false);
  const [isReadMode, setIsReadMode] = React.useState(false);

  return (
    <>
      <Wrapper>
        <ThreadHero
          isReadMode={isReadMode}
          imageUrl={imagesQuery.data?.[0]?.url}
          imageAlt={threadQuery.data?.name ?? ''}
          colorHex={threadQuery.data?.color_hex}
        />
        <ThreadHeader
          threadId={id}
          threadName={threadQuery.data?.name ?? ''}
          isReadMode={isReadMode}
          onToggleMode={() => setIsReadMode(!isReadMode)}
          onShare={() => setOpenShareDialog(true)}
        />
      </Wrapper>

      <Wrapper>
        <MomentList moments={dummyThreadStories} />
      </Wrapper>

      {dummyNotes.length > 0 && (
        <Wrapper className="h-dvh flex justify-center text-center items-center p-0">
          <p className="text-muted-foreground text-xl">End of history</p>
        </Wrapper>
      )}

      {!isReadMode && (
        <MomentInput
          onOpenTimeDialog={() => setOpenTimeDialog(true)}
          onPublish={() => console.log('Publishing...')}
        />
      )}

      <DateTimeDialog
        open={openTimeDialog}
        onOpenChange={setOpenTimeDialog}
        onSave={() => setOpenTimeDialog(false)}
      />

      <ColorPickerDialog
        open={openCostumizationDialog}
        onOpenChange={setOpenCostumizationDialog}
        onSave={() => setOpenCostumizationDialog(false)}
      />

      <ShareDialog open={openShareDialog} onOpenChange={setOpenShareDialog} />
    </>
  );
}
