import { Button } from '@/components/ui/button';
import { CircleNameLabel } from '@/features/circles';
import {
  audienceContextKey,
  type AudienceContext,
} from '@/features/moments/lib/use-moment-audience';

interface AudiencePickerProps {
  contexts: AudienceContext[];
  value: AudienceContext | undefined;
  onChange: (context: AudienceContext) => void;
}

// Shown only when a viewer has more than one valid Response context for
// a Moment (FEATURES.md, Response) — e.g. a Circle member who is also
// mentioned — letting them choose which audience to post/react into.
export function AudiencePicker({
  contexts,
  value,
  onChange,
}: AudiencePickerProps) {
  return (
    <div className="flex flex-wrap gap-1">
      {contexts.map((context) => {
        const key = audienceContextKey(context);
        const isSelected = value ? audienceContextKey(value) === key : false;
        return (
          <Button
            key={key}
            type="button"
            size="sm"
            variant={isSelected ? 'default' : 'outline'}
            onClick={() => onChange(context)}
          >
            {context.type === 'mention' ? (
              'Mentioned'
            ) : (
              <CircleNameLabel circleId={context.circleId} />
            )}
          </Button>
        );
      })}
    </div>
  );
}
