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
import { useForm } from '@tanstack/react-form';
import * as z from 'zod';

const formSchema = z.object({
  title: z
    .string()
    .min(1, 'Name must be at least 1 character.')
    .max(50, 'Name must be at most 50 characters.'),
});

export function CreateThreadForm() {
  const form = useForm({
    defaultValues: {
      title: '',
    },
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: async ({ value }) => {
      console.log(value);
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
            name="title"
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
                        {field.state.value.length}/50
                      </InputGroupText>
                    </InputGroupAddon>
                    <InputGroupInput
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      aria-invalid={isInvalid}
                      placeholder="Watching movies"
                      autoComplete="off"
                      className="text-lg!"
                    />
                  </InputGroup>
                  {isInvalid && <FieldError errors={field.state.meta.errors} />}
                </Field>
              );
            }}
          />

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
