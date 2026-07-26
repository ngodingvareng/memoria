interface ActivityHeroProps {
  isReadMode: boolean;
  imageUrl: string;
  imageAlt: string;
}

export function ActivityHero({
  isReadMode,
  imageUrl,
  imageAlt,
}: ActivityHeroProps) {
  if (isReadMode) return null;

  return (
    <div className="h-40 rounded-4xl relative overflow-hidden">
      <div className="bg-primary absolute inset-0 z-30 aspect-video opacity-50 mix-blend-color" />
      <img
        width={400}
        height={400}
        src={imageUrl}
        alt={imageAlt}
        title={imageAlt}
        className="relative z-20 aspect-video w-full object-cover brightness-60 grayscale"
      />
    </div>
  );
}
