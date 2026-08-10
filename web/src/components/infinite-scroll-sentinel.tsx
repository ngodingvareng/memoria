import { Loading03Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { useEffect, useRef } from 'react';

interface InfiniteScrollSentinelProps {
  onIntersect: () => void;
  enabled: boolean;
  // Shows a centered spinner while the next page is in flight — same
  // idea as YouTube's feed: nothing pops in out of empty space, a
  // loading state fills that space until it does.
  isLoading?: boolean;
}

export function InfiniteScrollSentinel({
  onIntersect,
  enabled,
  isLoading = false,
}: InfiniteScrollSentinelProps) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!enabled) return;
    const el = ref.current;
    if (!el) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) onIntersect();
      },
      // Triggers well before the sentinel is actually visible, so the
      // next page has time to arrive before the user scrolls into the
      // gap it would otherwise leave.
      { rootMargin: '800px' }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [enabled, onIntersect]);

  return (
    <div ref={ref} className="flex justify-center py-8">
      {isLoading && (
        <HugeiconsIcon
          icon={Loading03Icon}
          strokeWidth={2}
          className="text-muted-foreground size-6 animate-spin"
        />
      )}
    </div>
  );
}
