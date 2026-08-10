import { Cancel01Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { useEffect, useMemo } from 'react';

interface ImagePreviewListProps {
  images: File[];
  onRemove: (index: number) => void;
}

export function ImagePreviewList({ images, onRemove }: ImagePreviewListProps) {
  const previews = useMemo(
    () => images.map((file) => URL.createObjectURL(file)),
    [images]
  );

  useEffect(() => {
    return () => previews.forEach((url) => URL.revokeObjectURL(url));
  }, [previews]);

  if (images.length === 0) return null;

  return (
    <div className="flex flex-wrap gap-2">
      {previews.map((url, index) => (
        <div
          key={url}
          className="group relative size-20 overflow-hidden rounded-lg"
        >
          <img src={url} alt="" className="size-full object-cover" />
          <button
            type="button"
            onClick={() => onRemove(index)}
            className="absolute top-1 right-1 flex size-5 items-center justify-center rounded-full bg-black/60 text-white opacity-0 transition-opacity group-hover:opacity-100"
          >
            <HugeiconsIcon
              icon={Cancel01Icon}
              className="size-3"
              strokeWidth={2}
            />
            <span className="sr-only">Remove image</span>
          </button>
        </div>
      ))}
    </div>
  );
}
