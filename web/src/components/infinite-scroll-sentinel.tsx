import { useEffect, useRef } from 'react';

interface InfiniteScrollSentinelProps {
  onIntersect: () => void;
  enabled: boolean;
}

// A near-invisible marker at the end of a list — once it scrolls into
// view, onIntersect fires to load the next page.
export function InfiniteScrollSentinel({
  onIntersect,
  enabled,
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
      { rootMargin: '400px' }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [enabled, onIntersect]);

  return <div ref={ref} className="h-1" />;
}
