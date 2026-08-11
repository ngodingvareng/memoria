import { cn } from '@/lib/utils';

interface WrapperProps {
  children: React.ReactNode;
  className?: string;
  fullWidth?: boolean;
}

export default function Wrapper({
  children,
  className,
  fullWidth = false,
}: WrapperProps) {
  return (
    <div className={cn(className)}>
      <div
        className={cn(
          'mx-auto grid w-full min-w-0 px-3 py-4 pt-6 sm:px-4 sm:py-6 sm:pt-2 lg:py-12',
          fullWidth ? 'max-w-none' : 'max-w-5xl 2xl:max-w-5xl'
        )}
      >
        {children}
      </div>
    </div>
  );
}
