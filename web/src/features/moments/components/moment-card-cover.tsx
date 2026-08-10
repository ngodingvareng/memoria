import { hexToRgba } from '@/lib/colors';
import { cn } from '@/lib/utils';
import React from 'react';

interface MomentCardCoverProps {
  colorHex?: string;
  capturedAt: Date;
  coverImage?: string;
  onOpen: () => void;
}

export function MomentCardCover({
  colorHex,
  capturedAt,
  coverImage,
  onOpen,
}: MomentCardCoverProps) {
  // Fades the cover image in once it's actually decoded instead of
  // popping in mid-layout — most noticeable scrolling fast through a
  // freshly-loaded page of cards, where several images finish loading
  // at once.
  const [isImageLoaded, setIsImageLoaded] = React.useState(false);
  React.useEffect(() => {
    setIsImageLoaded(false);
  }, [coverImage]);

  const buttonRef = React.useRef<HTMLButtonElement>(null);
  const labelRef = React.useRef<HTMLDivElement>(null);
  const [isCompact, setIsCompact] = React.useState(true);

  React.useLayoutEffect(() => {
    const button = buttonRef.current;
    const label = labelRef.current;
    if (!button || !label) return;

    const update = () => {
      setIsCompact(button.offsetHeight <= label.offsetHeight + 10);
    };
    update();

    const observer = new ResizeObserver(update);
    observer.observe(button);
    observer.observe(label);
    return () => observer.disconnect();
  }, []);

  return (
    <button
      ref={buttonRef}
      type="button"
      disabled={!coverImage}
      onClick={onOpen}
      style={{
        backgroundColor: colorHex ? hexToRgba(colorHex, 1) : undefined,
      }}
      className={cn(
        'max-w-2xs flex flex-none flex-col rounded-xl w-full overflow-hidden relative isolate text-left outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50',
        !colorHex && 'bg-gray-500',
        coverImage && 'cursor-pointer'
      )}
    >
      <div
        ref={labelRef}
        className="items-end text-right text-white flex flex-col gap-1 flex-none relative z-10 px-4 py-3.5"
      >
        <div className="flex flex-col">
          <p className="font-bold text-lg">
            {capturedAt.toLocaleDateString('id-ID', {
              day: 'numeric',
              month: 'long',
              year: 'numeric',
            })}
          </p>
          <p className="font-medium text-xl/4">
            {capturedAt.toLocaleTimeString('id-ID', {
              hour: '2-digit',
              minute: '2-digit',
            })}
          </p>
        </div>
      </div>
      {coverImage && (
        <div
          className={cn(
            'absolute inset-x-0 bottom-0 z-0',
            isCompact ? 'top-0' : 'h-2/3'
          )}
        >
          <div
            className="absolute inset-0 z-10"
            style={
              isCompact
                ? {
                    backgroundColor: colorHex
                      ? hexToRgba(colorHex, 0.8)
                      : 'color-mix(in srgb, #6b7280 40%, transparent)',
                  }
                : {
                    backgroundImage: colorHex
                      ? `linear-gradient(to bottom, ${hexToRgba(colorHex, 1)}, ${hexToRgba(colorHex, 0.4)})`
                      : `linear-gradient(to bottom, #6b7280, color-mix(in srgb, #6b7280 40%, transparent))`,
                  }
            }
          />
          <img
            src={coverImage}
            alt=""
            loading="lazy"
            decoding="async"
            onLoad={() => setIsImageLoaded(true)}
            className={cn(
              'absolute inset-0 size-full object-cover transition-opacity duration-300',
              isImageLoaded ? 'opacity-100' : 'opacity-0'
            )}
          />
        </div>
      )}
    </button>
  );
}
