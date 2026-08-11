import { ColorSwatchPicker } from '@/components/color-swatch-picker';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field';
import { Textarea } from '@/components/ui/textarea';
import { useFormSubmitError } from '@/hooks/use-form-submit-error';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse } from '@/lib/api/generated/models';
import {
  getGetThreadsIdQueryKey,
  getGetThreadsQueryKey,
  usePutThreadsId,
} from '@/lib/api/generated/threads/threads';
import { queryClient } from '@/lib/query-client';
import { useForm } from '@tanstack/react-form';
import * as z from 'zod';

interface EditThreadDetailsSectionProps {
  threadId: string;
  thread: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse;
}

const hexColorPattern = /^#[0-9a-fA-F]{6}$/;

const formSchema = z.object({
  description: z
    .string()
    .max(2000, 'Description must be at most 2000 characters.'),
  color_hex: z
    .string()
    .refine(
      (value) => value === '' || hexColorPattern.test(value),
      'Enter a valid hex color, e.g. #4F46E5.'
    ),
});

export function EditThreadDetailsSection({
  threadId,
  thread,
}: EditThreadDetailsSectionProps) {
  const { error: submitError, guard } = useFormSubmitError();
  const updateThread = usePutThreadsId();

  const form = useForm({
    defaultValues: {
      description: thread.description ?? '',
      color_hex: thread.color_hex ?? '',
    },
    validators: { onSubmit: formSchema },
    onSubmit: ({ value }) =>
      guard(async () => {
        await updateThread.mutateAsync({
          id: threadId,
          data: {
            name: thread.name ?? '',
            description: value.description,
            color_hex: value.color_hex || undefined,
          },
        });
        await Promise.all([
          queryClient.invalidateQueries({
            queryKey: getGetThreadsIdQueryKey(threadId),
          }),
          queryClient.invalidateQueries({
            queryKey: getGetThreadsQueryKey(),
          }),
        ]);
      }),
  });

  return (
    <div className="flex flex-col gap-4">
      {/* <h2 className="text-xl font-semibold">Description & color</h2> */}

      <form
        id="edit-thread-details-form"
        onSubmit={(e) => {
          e.preventDefault();
          form.handleSubmit();
        }}
      >
        <FieldGroup>
          <form.Field name="description">
            {(field) => {
              const isInvalid =
                field.state.meta.isTouched && !field.state.meta.isValid;
              return (
                <Field data-invalid={isInvalid}>
                  <FieldLabel htmlFor={field.name}>Description</FieldLabel>
                  <Textarea
                    id={field.name}
                    name={field.name}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    aria-invalid={isInvalid}
                    placeholder="What's this thread about?"
                    className="min-h-24"
                  />
                  {isInvalid && <FieldError errors={field.state.meta.errors} />}
                </Field>
              );
            }}
          </form.Field>

          <form.Field name="color_hex">
            {(field) => {
              const isInvalid =
                field.state.meta.isTouched && !field.state.meta.isValid;
              return (
                <Field data-invalid={isInvalid}>
                  <FieldLabel htmlFor={field.name}>Color</FieldLabel>
                  <ColorSwatchPicker
                    value={field.state.value}
                    onChange={(hex) => field.handleChange(hex)}
                  />
                  {isInvalid && <FieldError errors={field.state.meta.errors} />}
                </Field>
              );
            }}
          </form.Field>

          {submitError && (
            <Alert variant="destructive">
              <AlertDescription>{submitError}</AlertDescription>
            </Alert>
          )}

          <Field orientation="horizontal" className="justify-end">
            <form.Subscribe
              selector={(state) => [state.isDefaultValue, state.isSubmitting]}
            >
              {([isDefaultValue, isSubmitting]) => (
                <Button
                  type="submit"
                  form="edit-thread-details-form"
                  disabled={Boolean(isDefaultValue) || Boolean(isSubmitting)}
                >
                  Save
                </Button>
              )}
            </form.Subscribe>
          </Field>
        </FieldGroup>
      </form>
    </div>
  );
}
