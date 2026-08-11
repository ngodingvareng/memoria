import { cn } from '@/lib/utils';

export default function Wrapper({
  children,
  className,
  fullWidth = false,
}: {
  children: React.ReactNode;
  className?: string;
  fullWidth?: boolean;
}) {
  return (
    <div className={cn(className)}>
      <div
        className={cn(
          'mx-auto grid w-full min-w-0 px-2 py-4 pt-2 sm:py-6 lg:py-12 sm:px-4',
          fullWidth ? 'max-w-none' : 'max-w-5xl 2xl:max-w-5xl'
        )}
      >
        {children}
      </div>
    </div>
  );
}
