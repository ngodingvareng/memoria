import { ColorPickerDialog } from '@/components/dialogs/color-picker-dialog';
import { DateTimeDialog } from '@/components/dialogs/datetime-dialog';
import { ShareDialog } from '@/components/dialogs/share-dialog';

import Wrapper from '@/components/wrapper';
import {
  CaptureInput,
  CaptureList,
  ActivityHeader,
  ActivityHero,
  type CaptureCardParam,
} from '@/features/activities';
import { dummyNotes } from '@/lib/dummies';
import { createFileRoute } from '@tanstack/react-router';
import React from 'react';

export const Route = createFileRoute('/_app/activity/$id/')({
  component: RouteComponent,
});

export const dummyActivityStories: CaptureCardParam[] = [
  {
    user: {
      name: 'Ahmad Faisal',
      username: '@ahmad_f',
      imageSrc: 'https://randomuser.me/api/portraits/men/45.jpg',
      imageAlt: 'Foto profil Ahmad Faisal',
    },
    activity: {
      name: 'Mendaki Gunung Rinjani',
    },
    color: 'emerald', // Tailwind color
    content: (
      <div className="flex flex-col gap-3">
        <p>
          Perjalanan panjang dan melelahkan terbayar lunas saat melihat matahari
          terbit dari puncak Rinjani. Sungguh pengalaman magis yang tak
          terlupakan! 🏔️✨
        </p>
        <img
          src="https://images.unsplash.com/photo-1596404835697-3d964fdd9d49?ixlib=rb-1.2.1&auto=format&fit=crop&w=600&q=80"
          alt="Puncak Gunung Rinjani"
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
      imageAlt: 'Foto profil Nadia Putri',
    },
    activity: {
      name: 'Nonton Konser Musik',
    },
    color: 'purple', // Tailwind color
    content: (
      <div>
        <p>
          Pecah banget malam ini! Berhasil dapet posisi paling depan dan nyanyi
          bareng dari awal sampai akhir. Suara udah habis tapi{' '}
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
      imageAlt: 'Foto profil Chef Dimas',
    },
    activity: {
      name: 'Eksperimen Resep Baru',
    },
    color: 'orange', // Tailwind color
    content: (
      <div className="space-y-2">
        <p>
          Hari ini mencoba membuat <b>Beef Wellington</b> untuk pertama kalinya.{' '}
          <i>Pastry</i>-nya renyah dan dagingnya matang sempurna medium rare!
          🥩🔥 Ada yang mau resepnya?
        </p>
        <img
          src="https://images.unsplash.com/photo-1600891964092-4316c288032e?ixlib=rb-1.2.1&auto=format&fit=crop&w=600&q=80"
          alt="Steak Daging"
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
      imageAlt: 'Foto profil Kevin',
    },
    activity: {
      name: 'Turnamen Valorant',
    },
    color: 'rose', // Tailwind color
    content: (
      <div>
        <p>
          Momen <i>Clutch 1v5</i> yang mengamankan tiket tim kami ke babak Final
          besok! Tangan masih gemeteran tapi GGWP untuk semua tim yang
          bertanding hari ini. 🎮🏆
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
      imageAlt: 'Foto profil Maya Sari',
    },
    activity: {
      name: 'Selesai Redesign UI/UX',
    },
    color: 'cyan', // Tailwind color
    content: (
      <div>
        <p>
          Akhirnya proyek <i>redesign</i> aplikasi <i>fintech</i> selesai dan
          sudah di-<i>handover</i> ke tim developer. Sangat puas dengan
          pemilihan palet warna dan interaksinya. Waktunya istirahat! 💻🎨
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
      imageAlt: 'Foto profil Sinta Wijaya',
    },
    activity: {
      name: 'Sesi Yoga Sore',
    },
    color: 'teal', // Tailwind color
    content: (
      <div className="flex flex-col gap-2">
        <p>
          Menutup akhir pekan dengan sesi Vinyasa Yoga. Sangat membantu untuk
          menenangkan pikiran sebelum menghadapi hari Senin besok. Namaste 🙏🧘‍♀️
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
  const [openCostumizationDialog, setOpenCostumizationDialog] =
    React.useState(false);
  const [openShareDialog, setOpenShareDialog] = React.useState(false);
  const [openTimeDialog, setOpenTimeDialog] = React.useState(false);
  const [isReadMode, setIsReadMode] = React.useState(false);

  const noteCount = dummyNotes.length;

  return (
    <>
      <Wrapper>
        <ActivityHero
          isReadMode={isReadMode}
          imageUrl="https://images.unsplash.com/photo-1604076850742-4c7221f3101b?q=80&w=1887&auto=format&fit=crop&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D"
          imageAlt="Photo by mymind on Unsplash"
        />
        <ActivityHeader
          isReadMode={isReadMode}
          noteCount={noteCount}
          onToggleMode={() => setIsReadMode(!isReadMode)}
          onShare={() => setOpenShareDialog(true)}
        />
      </Wrapper>

      <Wrapper>
        <CaptureList captures={dummyActivityStories} />
      </Wrapper>

      {dummyNotes.length > 0 && (
        <Wrapper className="h-dvh flex justify-center text-center items-center p-0">
          <p className="text-muted-foreground text-xl">End of history</p>
        </Wrapper>
      )}

      {!isReadMode && (
        <CaptureInput
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
