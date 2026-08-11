import { cn } from '@/lib/utils';

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
        'relative aspect-7/1 overflow-hidden rounded-xl sm:rounded-4xl',
        className
      )}
    >
      <div
        className="bg-primary absolute inset-0 z-3 aspect-7/1 opacity-50 mix-blend-color"
        style={colorHex ? { backgroundColor: colorHex } : undefined}
      />
      {imageUrl ? (
        <img
          width={400}
          height={400}
          src={imageUrl}
          alt={imageAlt}
          title={imageAlt}
          className="bg-muted relative z-2 aspect-7/1 w-full object-cover brightness-60"
        />
      ) : (
        <div
          className='className="relative brightness-60" z-2 aspect-7/1 w-full object-cover'
          style={colorHex ? { backgroundColor: colorHex } : undefined}
        />
      )}
    </div>
  );
}
