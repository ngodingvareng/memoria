import { cn } from '@/lib/utils';
import { MOMENT_COLOR_PRESETS } from '../lib/color-presets';

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
    <div className={cn('flex gap-2 flex-wrap', className)}>
      {MOMENT_COLOR_PRESETS.map((preset) => (
        <button
          key={preset.hex}
          type="button"
          aria-label={preset.name}
          onClick={() => onChange(value === preset.hex ? '' : preset.hex)}
          style={{ backgroundColor: preset.hex }}
          className={cn(
            'size-7 rounded-full cursor-pointer ring-offset-2 ring-offset-background transition-all',
            value === preset.hex
              ? 'ring-2 ring-primary'
              : 'hover:ring-2 hover:ring-primary/50'
          )}
        />
      ))}
    </div>
  );
}
