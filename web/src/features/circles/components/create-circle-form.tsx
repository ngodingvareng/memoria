import {
  getGetCirclesQueryKey,
  usePostCircles,
} from '@/lib/api/generated/circles/circles';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field';
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
  InputGroupText,
} from '@/components/ui/input-group';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { ApiError } from '@/lib/api-client';
import { queryClient } from '@/lib/query-client';
import { useForm } from '@tanstack/react-form';
import { useNavigate } from '@tanstack/react-router';
import { useState } from 'react';
import * as z from 'zod';

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

export function CreateCircleForm() {
  const navigate = useNavigate();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const createCircleMutation = usePostCircles();

  const form = useForm({
    defaultValues: {
      name: '',
      description: '',
      color_hex: '',
    },
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: async ({ value }) => {
      setSubmitError(null);
      try {
        const circle = await createCircleMutation.mutateAsync({
          data: {
            name: value.name,
            description: value.description || undefined,
            color_hex: value.color_hex || undefined,
          },
        });
        await queryClient.invalidateQueries({
          queryKey: getGetCirclesQueryKey(),
        });
        navigate({ to: '/c/$id', params: { id: circle.id! } });
      } catch (err) {
        if (err instanceof ApiError) {
          setSubmitError(
            err.fieldErrors?.map((e) => e.message).join(' ') ?? err.message
          );
        } else {
          setSubmitError('Something went wrong. Please try again.');
        }
      }
    },
  });
  return (
    <>
      <form
        id="create-circle-form"
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
                  <FieldLabel htmlFor={field.name} hidden>
                    Name
                  </FieldLabel>
                  <InputGroup>
                    <InputGroupAddon align="block-start">
                      <InputGroupText>Name</InputGroupText>

                      <InputGroupText className="ml-auto tabular-nums text-xs text-muted-foreground">
                        {field.state.value.length}/255
                      </InputGroupText>
                    </InputGroupAddon>
                    <InputGroupInput
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      aria-invalid={isInvalid}
                      placeholder="NgodingVareng"
                      autoComplete="off"
                      className="text-lg!"
                    />
                  </InputGroup>
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
                  form="create-circle-form"
                  disabled={isDefaultValue || isSubmitting}
                >
                  Create
                </Button>
              )}
            />
          </Field>
        </FieldGroup>
      </form>
    </>
  );
}
