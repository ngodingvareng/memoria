import { Item, ItemContent } from '@/components/ui/item';
import { CaptureList, type CaptureCardParam } from '@/features/activities';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/_story/story/')({
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
  return (
    <div className="flex flex-col gap-4">
      <Item variant="muted">
        <ItemContent>Publish your capture...</ItemContent>
      </Item>

      <CaptureList captures={dummyActivityStories} withStoryLayout />
    </div>
  );
}
