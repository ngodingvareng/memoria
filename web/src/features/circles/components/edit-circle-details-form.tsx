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
import { useFormSubmitError } from '@/hooks/use-form-submit-error';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoCircleResponse } from '@/lib/api/generated/models';
import {
  getGetCirclesIdQueryKey,
  getGetCirclesQueryKey,
  usePutCirclesId,
} from '@/lib/api/generated/circles/circles';
import { queryClient } from '@/lib/query-client';
import { useForm } from '@tanstack/react-form';
import * as z from 'zod';

interface EditCircleDetailsFormProps {
  circleId: string;
  circle: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoCircleResponse;
}

const hexColorPattern = /^#[0-9a-fA-F]{6}$/;

const formSchema = z.object({
  name: z
    .string()
    .min(1, 'Name must be at least 1 character.')
    .max(255, 'Name must be at most 255 characters.'),
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

export function EditCircleDetailsForm({
  circleId,
  circle,
}: EditCircleDetailsFormProps) {
  const { error: submitError, guard } = useFormSubmitError();
  const updateCircle = usePutCirclesId();

  const form = useForm({
    defaultValues: {
      name: circle.name ?? '',
      description: circle.description ?? '',
      color_hex: circle.color_hex ?? '',
    },
    validators: { onSubmit: formSchema },
    onSubmit: ({ value }) =>
      guard(async () => {
        await updateCircle.mutateAsync({
          id: circleId,
          data: {
            name: value.name,
            description: value.description || undefined,
            color_hex: value.color_hex || undefined,
          },
        });
        await Promise.all([
          queryClient.invalidateQueries({
            queryKey: getGetCirclesIdQueryKey(circleId),
          }),
          queryClient.invalidateQueries({
            queryKey: getGetCirclesQueryKey(),
          }),
        ]);
      }),
  });

  return (
    <form
      id="edit-circle-details-form"
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
    >
      <FieldGroup>
        <form.Field
          name="name"
          children={(field) => {
            const isInvalid =
              field.state.meta.isTouched && !field.state.meta.isValid;
            return (
              <Field data-invalid={isInvalid}>
                <FieldLabel htmlFor={field.name}>Name</FieldLabel>
                <Input
                  id={field.name}
                  name={field.name}
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  aria-invalid={isInvalid}
                  autoComplete="off"
                />
                {isInvalid && <FieldError errors={field.state.meta.errors} />}
              </Field>
            );
          }}
        />

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
                  placeholder="What's this circle about?"
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
                    className="border-border size-9 cursor-pointer rounded-full border bg-transparent p-0"
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
                form="edit-circle-details-form"
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
