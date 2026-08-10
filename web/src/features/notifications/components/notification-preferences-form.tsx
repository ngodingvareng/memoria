import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldSeparator,
  FieldTitle,
} from '@/components/ui/field';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { Switch } from '@/components/ui/switch';
import { useFormSubmitError } from '@/hooks/use-form-submit-error';
import {
  getGetNotificationsPreferencesQueryKey,
  useGetNotificationsPreferences,
  usePutNotificationsPreferences,
} from '@/lib/api/generated/notifications/notifications';
import { queryClient } from '@/lib/query-client';
import { useForm } from '@tanstack/react-form';

// HH:MM (native <input type="time">'s format) <-> HH:MM:SS (the API's
// wall-clock format, see NotificationPreferencesResponse.quiet_hours_*).
function toApiTime(value: string): string | undefined {
  return value ? `${value}:00` : undefined;
}
function fromApiTime(value?: string): string {
  return value ? value.slice(0, 5) : '';
}

const digestHourOptions = Array.from({ length: 24 }, (_, hour) => hour);

export function NotificationPreferencesForm() {
  const preferencesQuery = useGetNotificationsPreferences();
  const { error: submitError, guard } = useFormSubmitError();
  const updatePreferences = usePutNotificationsPreferences();

  const preferences = preferencesQuery.data;

  const form = useForm({
    defaultValues: {
      commitment_upcoming_enabled:
        preferences?.commitment_upcoming_enabled ?? true,
      moment_due_enabled: preferences?.moment_due_enabled ?? true,
      moment_missed_enabled: preferences?.moment_missed_enabled ?? false,
      echo_ready_enabled: preferences?.echo_ready_enabled ?? true,
      recap_generated_enabled: preferences?.recap_generated_enabled ?? true,
      circle_invite_received_enabled:
        preferences?.circle_invite_received_enabled ?? true,
      circle_join_request_received_enabled:
        preferences?.circle_join_request_received_enabled ?? true,
      mentioned_in_moment_enabled:
        preferences?.mentioned_in_moment_enabled ?? true,
      added_to_circle_enabled: preferences?.added_to_circle_enabled ?? true,
      circle_join_request_approved_enabled:
        preferences?.circle_join_request_approved_enabled ?? true,
      response_digest_enabled: preferences?.response_digest_enabled ?? true,
      response_digest_hour: preferences?.response_digest_hour ?? 20,
      quiet_hours_start: fromApiTime(preferences?.quiet_hours_start),
      quiet_hours_end: fromApiTime(preferences?.quiet_hours_end),
    },
    onSubmit: ({ value }) =>
      guard(async () => {
        await updatePreferences.mutateAsync({
          data: {
            ...value,
            quiet_hours_start: toApiTime(value.quiet_hours_start),
            quiet_hours_end: toApiTime(value.quiet_hours_end),
          },
        });
        await queryClient.invalidateQueries({
          queryKey: getGetNotificationsPreferencesQueryKey(),
        });
      }),
  });

  if (preferencesQuery.isPending) {
    return (
      <div className="flex flex-col gap-3">
        <Skeleton className="h-10 w-full rounded-2xl" />
        <Skeleton className="h-10 w-full rounded-2xl" />
        <Skeleton className="h-10 w-full rounded-2xl" />
      </div>
    );
  }

  return (
    <form
      id="notification-preferences-form"
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
    >
      <FieldGroup>
        <ToggleField
          form={form}
          name="mentioned_in_moment_enabled"
          label="Mentions"
          description="Someone names you in a moment."
        />
        <ToggleField
          form={form}
          name="circle_invite_received_enabled"
          label="Circle invites"
          description="Someone invites you to a circle."
        />
        <ToggleField
          form={form}
          name="circle_join_request_received_enabled"
          label="Join requests"
          description="Someone asks to join a circle you can invite into."
        />

        <FieldSeparator />

        <ToggleField
          form={form}
          name="added_to_circle_enabled"
          label="Added to a circle"
          description="You're added to a circle, changing who can see your moments."
        />
        <ToggleField
          form={form}
          name="circle_join_request_approved_enabled"
          label="Join request approved"
          description="Your request to join a circle is approved."
        />

        <FieldSeparator />

        <ToggleField
          form={form}
          name="echo_ready_enabled"
          label="Echoes"
          description="A flashback is ready to look back on."
        />
        <ToggleField
          form={form}
          name="recap_generated_enabled"
          label="Recaps"
          description="A period recap has been generated."
        />

        <FieldSeparator />

        <ToggleField
          form={form}
          name="commitment_upcoming_enabled"
          label="Commitment upcoming"
          description="A reminder before a commitment is due."
        />
        <ToggleField
          form={form}
          name="moment_due_enabled"
          label="Moment due"
          description="A commitment's moment has come due."
        />
        <ToggleField
          form={form}
          name="moment_missed_enabled"
          label="Moment missed"
          description="A commitment's moment was missed. Off by default."
        />

        <FieldSeparator />

        <ToggleField
          form={form}
          name="response_digest_enabled"
          label="Response digest"
          description="A once-daily summary of comments and reactions, never individual pings."
        />

        <form.Field
          name="response_digest_hour"
          children={(field) => (
            <Field orientation="horizontal">
              <FieldContent>
                <FieldTitle>Digest hour</FieldTitle>
                <FieldDescription>
                  What time of day the response digest is delivered.
                </FieldDescription>
              </FieldContent>
              <Select
                value={String(field.state.value)}
                onValueChange={(value) => field.handleChange(Number(value))}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {digestHourOptions.map((hour) => (
                    <SelectItem key={hour} value={String(hour)}>
                      {String(hour).padStart(2, '0')}:00
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
          )}
        />

        <FieldSeparator />

        <Field>
          <FieldLabel>Quiet hours</FieldLabel>
          <FieldDescription>
            Nothing is delivered during this window; anything due is held until
            it ends.
          </FieldDescription>
          <div className="flex items-center gap-2">
            <form.Field
              name="quiet_hours_start"
              children={(field) => (
                <input
                  type="time"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  className="bg-input/50 focus-visible:border-ring focus-visible:ring-ring/30 h-9 rounded-3xl border border-transparent px-3 text-sm outline-none focus-visible:ring-3"
                />
              )}
            />
            <span className="text-muted-foreground text-sm">to</span>
            <form.Field
              name="quiet_hours_end"
              children={(field) => (
                <input
                  type="time"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  className="bg-input/50 focus-visible:border-ring focus-visible:ring-ring/30 h-9 rounded-3xl border border-transparent px-3 text-sm outline-none focus-visible:ring-3"
                />
              )}
            />
          </div>
        </Field>

        {submitError && (
          <Alert variant="destructive">
            <AlertDescription>{submitError}</AlertDescription>
          </Alert>
        )}

        <Field orientation="horizontal" className="justify-end">
          <form.Subscribe
            selector={(state) => [state.isDefaultValue, state.isSubmitting]}
            children={([isDefaultValue, isSubmitting]) => (
              <Button
                type="submit"
                form="notification-preferences-form"
                disabled={Boolean(isDefaultValue) || Boolean(isSubmitting)}
              >
                Save
              </Button>
            )}
          />
        </Field>
      </FieldGroup>
    </form>
  );
}

interface ToggleFieldProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: any;
  name: string;
  label: string;
  description: string;
}

// Every preference below is a plain boolean toggle with the same
// label+description+Switch layout — a small local helper avoids
// repeating that layout eleven times without over-abstracting the
// per-field TanStack Form binding itself.
function ToggleField({ form, name, label, description }: ToggleFieldProps) {
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
