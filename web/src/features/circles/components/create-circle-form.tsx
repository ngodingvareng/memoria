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
import { Textarea } from '@/components/ui/textarea';
import { ColorSwatchPicker } from '@/components/color-swatch-picker';
import { useFormSubmitError } from '@/hooks/use-form-submit-error';
import {
  getGetCirclesQueryKey,
  usePostCircles,
} from '@/lib/api/generated/circles/circles';
import { queryClient } from '@/lib/query-client';
import { useForm } from '@tanstack/react-form';
import { useNavigate } from '@tanstack/react-router';
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
  const { error: submitError, guard } = useFormSubmitError();
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
    onSubmit: ({ value }) =>
      guard(async () => {
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
      }),
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

                      <InputGroupText className="text-muted-foreground ml-auto text-xs tabular-nums">
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
                  <ColorSwatchPicker
                    value={field.state.value}
                    onChange={(hex) => field.handleChange(hex)}
                  />
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
