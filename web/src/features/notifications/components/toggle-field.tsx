import {
  Field,
  FieldContent,
  FieldDescription,
  FieldTitle,
} from '@/components/ui/field';
import { Switch } from '@/components/ui/switch';

interface ToggleFieldProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: any;
  name: string;
  label: string;
  description: string;
}

// Every preference in NotificationPreferencesForm is a plain boolean
// toggle with the same label+description+Switch layout — this helper
// avoids repeating that layout eleven times without over-abstracting the
// per-field TanStack Form binding itself.
export function ToggleField({
  form,
  name,
  label,
  description,
}: ToggleFieldProps) {
  return (
    <form.Field
      name={name}
      children={(field: {
        state: { value: boolean };
        handleChange: (value: boolean) => void;
        name: string;
      }) => (
        <Field orientation="horizontal">
          <FieldContent>
            <FieldTitle>{label}</FieldTitle>
            <FieldDescription>{description}</FieldDescription>
          </FieldContent>
          <Switch
            checked={field.state.value}
            onCheckedChange={(checked) => field.handleChange(checked)}
          />
        </Field>
      )}
    />
  );
}
