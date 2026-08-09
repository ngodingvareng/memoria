import { cn } from '@/lib/utils';

// Neutral texture tinted per-thread via the color overlay below — used
// whenever a thread has no cover image of its own uploaded yet.
const DEFAULT_COVER_IMAGE_URL =
  'https://images.unsplash.com/photo-1604076850742-4c7221f3101b?q=80&w=1887&auto=format&fit=crop&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D';

interface ThreadHeroProps {
  isReadMode: boolean;
  imageUrl?: string;
  imageAlt: string;
  colorHex?: string;
  className?: string;
}

export function ThreadHero({
  isReadMode,
  imageUrl,
  imageAlt,
  colorHex,
  className,
}: ThreadHeroProps) {
  if (isReadMode) return null;

  return (
    <div
      className={cn(
        'aspect-7/1 rounded-4xl relative overflow-hidden',
        className
      )}
    >
      <div
        className="bg-primary absolute inset-0 z-3 aspect-7/1 opacity-50 mix-blend-color"
        style={colorHex ? { backgroundColor: colorHex } : undefined}
      />
      <img
        width={400}
        height={400}
        src={imageUrl || DEFAULT_COVER_IMAGE_URL}
        alt={imageAlt}
        title={imageAlt}
        className="relative z-2 aspect-7/1 w-full object-cover brightness-60"
      />
    </div>
  );
}
