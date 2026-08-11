'use client';

import { cn } from '@/lib/utils';
import {
  ArrowLeft01Icon,
  ArrowRight01Icon,
  Cancel01Icon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import * as React from 'react';

export interface LightboxImage {
  src: string;
  alt?: string;
  caption?: string;
}

interface LightboxProps {
  images: LightboxImage[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
  index?: number;
  onIndexChange?: (index: number) => void;
  loop?: boolean;
  className?: string;
}

export function Lightbox({
  images,
  open,
  onOpenChange,
  index = 0,
  onIndexChange,
  loop = false,
  className,
}: LightboxProps) {
  const [currentIndex, setCurrentIndex] = React.useState(index);
  const [isLoading, setIsLoading] = React.useState(true);
  const preloadedIndexes = React.useRef<Set<number>>(new Set());

  const lightboxRef = React.useRef<HTMLDivElement>(null);

  const preloadImages = React.useCallback(
    (targetIndex: number) => {
      const imagesToPreload = [targetIndex];
      if (targetIndex > 0 || loop) {
        imagesToPreload.push(
          targetIndex > 0 ? targetIndex - 1 : images.length - 1
        );
      }
      if (targetIndex < images.length - 1 || loop) {
        imagesToPreload.push(
          targetIndex < images.length - 1 ? targetIndex + 1 : 0
        );
      }
      imagesToPreload.forEach((idx) => {
        if (preloadedIndexes.current.has(idx)) return;
        const img = new Image();
        img.src = images[idx].src;
        img.onload = () => {
          preloadedIndexes.current.add(idx);
        };
      });
    },
    [images, loop]
  );

  const goToIndex = React.useCallback(
    (nextIndex: number) => {
      setCurrentIndex(nextIndex);
      setIsLoading(true);
      onIndexChange?.(nextIndex);
      preloadImages(nextIndex);
    },
    [onIndexChange, preloadImages]
  );

  // Reset to the caller-provided index each time the lightbox is opened.
  React.useEffect(() => {
    if (!open) return;
    setCurrentIndex(index);
    setIsLoading(true);
    preloadImages(index);
    document.body.style.overflow = 'hidden';
    return () => {
      document.body.style.overflow = 'auto';
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  const prevImage = React.useCallback(() => {
    if (currentIndex > 0) {
      goToIndex(currentIndex - 1);
    } else if (loop) {
      goToIndex(images.length - 1);
    }
  }, [currentIndex, images.length, loop, goToIndex]);

  const nextImage = React.useCallback(() => {
    if (currentIndex < images.length - 1) {
      goToIndex(currentIndex + 1);
    } else if (loop) {
      goToIndex(0);
    }
  }, [currentIndex, images.length, loop, goToIndex]);

  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!open) return;
      switch (e.key) {
        case 'ArrowLeft':
          prevImage();
          break;
        case 'ArrowRight':
          nextImage();
          break;
        case 'Escape':
          onOpenChange(false);
          break;
        default:
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [open, prevImage, nextImage, onOpenChange]);

  // Focus trap for accessibility
  React.useEffect(() => {
    if (!open || !lightboxRef.current) return;

    const lightbox = lightboxRef.current;
    const focusableElements = lightbox.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );

    const firstElement = focusableElements[0] as HTMLElement | undefined;
    const lastElement = focusableElements[focusableElements.length - 1] as
      HTMLElement | undefined;

    const handleTabKey = (e: KeyboardEvent) => {
      if (e.key !== 'Tab' || !firstElement || !lastElement) return;

      if (e.shiftKey) {
        if (document.activeElement === firstElement) {
          lastElement.focus();
          e.preventDefault();
        }
      } else {
        if (document.activeElement === lastElement) {
          firstElement.focus();
          e.preventDefault();
        }
      }
    };

    lightbox.addEventListener('keydown', handleTabKey);
    firstElement?.focus();

    return () => {
      lightbox.removeEventListener('keydown', handleTabKey);
    };
  }, [open]);

  if (!open || images.length === 0) return null;

  const currentImage = images[currentIndex];

  return (
    <div
      ref={lightboxRef}
      className={cn(
        'fixed inset-0 z-50 flex items-center justify-center bg-black/90',
        className
      )}
      onClick={(e) => {
        if (e.target === e.currentTarget) onOpenChange(false);
      }}
      aria-modal="true"
      role="dialog"
    >
      <button
        className="absolute top-4 right-4 z-10 rounded-full bg-black/50 p-2 text-white transition-opacity hover:bg-black/70"
        onClick={() => onOpenChange(false)}
        aria-label="Close lightbox"
      >
        <HugeiconsIcon icon={Cancel01Icon} className="size-6" />
      </button>

      {(currentIndex > 0 || loop) && images.length > 1 && (
        <button
          className="absolute top-1/2 left-4 -translate-y-1/2 rounded-full bg-black/50 p-2 text-white transition-opacity hover:bg-black/70"
          onClick={prevImage}
          aria-label="Previous image"
        >
          <HugeiconsIcon icon={ArrowLeft01Icon} className="size-8" />
        </button>
      )}

      {(currentIndex < images.length - 1 || loop) && images.length > 1 && (
        <button
          className="absolute top-1/2 right-4 -translate-y-1/2 rounded-full bg-black/50 p-2 text-white transition-opacity hover:bg-black/70"
          onClick={nextImage}
          aria-label="Next image"
        >
          <HugeiconsIcon icon={ArrowRight01Icon} className="size-8" />
        </button>
      )}

      <div className="relative flex max-h-[90vh] max-w-[90vw] flex-col items-center">
        {isLoading && (
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="size-12 animate-spin rounded-full border-4 border-white border-t-transparent" />
          </div>
        )}

        <img
          src={currentImage.src}
          alt={currentImage.alt}
          className={cn(
            'max-h-[80vh] max-w-full object-contain transition-opacity duration-300',
            isLoading ? 'opacity-0' : 'opacity-100'
          )}
          onLoad={() => setIsLoading(false)}
        />

        {currentImage.caption && (
          <div className="mt-4 max-w-full px-4 text-center text-white">
            <p className="text-lg">{currentImage.caption}</p>
          </div>
        )}

        {images.length > 1 && (
          <div className="absolute bottom-4 left-4 rounded-full bg-black/50 px-3 py-1 text-sm text-white">
            {currentIndex + 1} / {images.length}
          </div>
        )}
      </div>
    </div>
  );
}
