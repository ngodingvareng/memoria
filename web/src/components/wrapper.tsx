import { cn } from '@/lib/utils';

export default function Wrapper({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn(className)}>
      <div className="mx-auto grid w-full max-w-5xl min-w-0 py-4 px-2 pt-2 sm:py-6 lg:py-12 2xl:max-w-5xl">
        {children}
      </div>
    </div>
  );
}
