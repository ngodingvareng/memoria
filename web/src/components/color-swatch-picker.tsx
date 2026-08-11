import { cn } from '@/lib/utils';
import { MOMENT_COLOR_PRESETS } from '@/lib/color-presets';

interface ColorSwatchPickerProps {
  value: string;
  onChange: (hex: string) => void;
  className?: string;
}

export function ColorSwatchPicker({
  value,
  onChange,
  className,
}: ColorSwatchPickerProps) {
  return (
    <div className={cn('flex flex-wrap gap-2', className)}>
      {MOMENT_COLOR_PRESETS.map((preset) => (
        <button
          key={preset.hex}
          type="button"
          aria-label={preset.name}
          onClick={() => onChange(value === preset.hex ? '' : preset.hex)}
          style={{ backgroundColor: preset.hex }}
          className={cn(
            'ring-offset-background size-7 cursor-pointer rounded-full ring-offset-1 transition-all',
            value === preset.hex
              ? 'ring-primary ring-2'
              : 'hover:ring-primary/50 hover:ring-2'
          )}
        />
      ))}
    </div>
  );
}
