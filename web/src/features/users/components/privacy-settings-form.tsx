import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldGroup,
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
  GithubComNgodingvarengMemoriaInternalDeliveryRestDtoUpdatePrivacySettingsRequestCircleInvitePolicy as CircleInvitePolicy,
  GithubComNgodingvarengMemoriaInternalDeliveryRestDtoUpdatePrivacySettingsRequestMentionPolicy as MentionPolicy,
} from '@/lib/api/generated/models';
import {
  getGetUsersMeQueryKey,
  useGetUsersMe,
  usePutUsersMePrivacy,
} from '@/lib/api/generated/users/users';
import { queryClient } from '@/lib/query-client';
import { useForm } from '@tanstack/react-form';

const policyOptions: { value: 'anyone' | 'known' | 'nobody'; label: string }[] =
  [
    { value: 'anyone', label: 'Anyone' },
    { value: 'known', label: 'People I know' },
    { value: 'nobody', label: 'No one' },
  ];

export function PrivacySettingsForm() {
  const meQuery = useGetUsersMe();
  const { error: submitError, guard } = useFormSubmitError();
  const updatePrivacy = usePutUsersMePrivacy();

  const me = meQuery.data;

  const form = useForm({
    defaultValues: {
      mention_policy: (me?.mention_policy ?? 'anyone') as MentionPolicy,
      circle_invite_policy: (me?.circle_invite_policy ??
        'anyone') as CircleInvitePolicy,
      discoverable_by_username: me?.discoverable_by_username ?? true,
      strip_photo_metadata: me?.strip_photo_metadata ?? false,
    },
    onSubmit: ({ value }) =>
      guard(async () => {
        await updatePrivacy.mutateAsync({ data: value });
        await queryClient.invalidateQueries({
          queryKey: getGetUsersMeQueryKey(),
        });
      }),
  });

  if (meQuery.isPending) {
    return (
      <div className="flex flex-col gap-3">
        <Skeleton className="h-10 w-full rounded-2xl" />
        <Skeleton className="h-10 w-full rounded-2xl" />
      </div>
    );
  }

  return (
    <form
      id="privacy-settings-form"
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
    >
      <FieldGroup>
        <form.Field
          name="mention_policy"
          children={(field) => (
            <Field orientation="horizontal">
              <FieldContent>
                <FieldTitle>Who can mention me</FieldTitle>
                <FieldDescription>
                  Controls who can name you in a moment.
                </FieldDescription>
              </FieldContent>
              <Select
                value={field.state.value}
                onValueChange={(value) =>
                  field.handleChange(value as MentionPolicy)
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {policyOptions.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
          )}
        />

        <form.Field
          name="circle_invite_policy"
          children={(field) => (
            <Field orientation="horizontal">
              <FieldContent>
                <FieldTitle>Who can invite me to a circle</FieldTitle>
                <FieldDescription>
                  Anyone else opens a pending invite you can accept or decline
                  instead.
                </FieldDescription>
              </FieldContent>
              <Select
                value={field.state.value}
                onValueChange={(value) =>
                  field.handleChange(value as CircleInvitePolicy)
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {policyOptions.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
          )}
        />

        <FieldSeparator />

        <form.Field
          name="discoverable_by_username"
          children={(field) => (
            <Field orientation="horizontal">
              <FieldContent>
                <FieldTitle>Discoverable by username</FieldTitle>
                <FieldDescription>
                  Lets others find you by searching your @username.
                </FieldDescription>
              </FieldContent>
              <Switch
                checked={field.state.value}
                onCheckedChange={(checked) => field.handleChange(checked)}
              />
            </Field>
          )}
        />

        <form.Field
          name="strip_photo_metadata"
          children={(field) => (
            <Field orientation="horizontal">
              <FieldContent>
                <FieldTitle>Strip photo metadata</FieldTitle>
                <FieldDescription>
                  Removes EXIF/GPS data from photos you upload.
                </FieldDescription>
              </FieldContent>
              <Switch
                checked={field.state.value}
                onCheckedChange={(checked) => field.handleChange(checked)}
              />
            </Field>
          )}
        />

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
                form="privacy-settings-form"
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
