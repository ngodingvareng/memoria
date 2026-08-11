import { BrandLogo } from './brand-logo';

export function Brand() {
  return (
    <div className="flex items-center gap-0.5">
      <BrandLogo />
      <span className="font-heading text-3xl font-semibold text-olive-700 dark:text-olive-300">
        memoria
      </span>
    </div>
  );
}
