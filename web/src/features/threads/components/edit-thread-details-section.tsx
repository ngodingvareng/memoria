import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { ApiError } from '@/lib/api-client';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse } from '@/lib/api/generated/models';
import {
  getGetThreadsIdQueryKey,
  getGetThreadsQueryKey,
  usePutThreadsId,
} from '@/lib/api/generated/threads/threads';
import { queryClient } from '@/lib/query-client';
import { useForm } from '@tanstack/react-form';
import { useState } from 'react';
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
  const [submitError, setSubmitError] = useState<string | null>(null);
  const updateThread = usePutThreadsId();

  const form = useForm({
    defaultValues: {
      description: thread.description ?? '',
      color_hex: thread.color_hex ?? '',
    },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value }) => {
      setSubmitError(null);
      try {
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
      } catch (err) {
        setSubmitError(
          err instanceof ApiError
            ? (err.fieldErrors?.map((e) => e.message).join(' ') ?? err.message)
            : 'Something went wrong. Please try again.'
        );
      }
    },
  });

  return (
    <div className="flex flex-col gap-4">
      <h2 className="text-xl font-semibold">Description & color</h2>

      <form
        id="edit-thread-details-form"
        onSubmit={(e) => {
          e.preventDefault();
          form.handleSubmit();
        }}
      >
        <FieldGroup>
          <form.Field
            name="description"
            children={(field) => {
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
          />

          <form.Field
            name="color_hex"
            children={(field) => {
              const isInvalid =
                field.state.meta.isTouched && !field.state.meta.isValid;
              return (
                <Field data-invalid={isInvalid}>
                  <FieldLabel htmlFor={field.name}>Color</FieldLabel>
                  <div className="flex items-center gap-2">
                    <input
                      type="color"
                      value={
                        hexColorPattern.test(field.state.value)
                          ? field.state.value
                          : '#94a3b8'
                      }
                      onChange={(e) => field.handleChange(e.target.value)}
                      className="size-9 cursor-pointer rounded-full border border-border bg-transparent p-0"
                      aria-label="Pick a color"
                    />
                    <Input
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      aria-invalid={isInvalid}
                      placeholder="#4F46E5"
                      autoComplete="off"
                      className="max-w-32"
                    />
                  </div>
                  {isInvalid && <FieldError errors={field.state.meta.errors} />}
                </Field>
              );
            }}
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
                  form="edit-thread-details-form"
                  disabled={Boolean(isDefaultValue) || Boolean(isSubmitting)}
                >
                  Save
                </Button>
              )}
            />
          </Field>
        </FieldGroup>
      </form>
    </div>
  );
}
