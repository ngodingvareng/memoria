import { BrandLogo } from './brand-logo';

export function Brand() {
  return (
    <div className="flex items-center gap-0.5">
      <BrandLogo />
      <span className="font-heading hidden text-3xl font-semibold text-olive-700 sm:block dark:text-olive-300">
        memoria
      </span>
    </div>
  );
}
