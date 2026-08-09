import {
  getGetCirclesIdThreadsQueryKey,
  getGetThreadsQueryKey,
  usePostThreads,
} from '@/lib/api/generated/threads/threads';
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
import { ApiError } from '@/lib/api-client';
import { queryClient } from '@/lib/query-client';
import { useForm } from '@tanstack/react-form';
import { useNavigate } from '@tanstack/react-router';
import { useState } from 'react';
import * as z from 'zod';

const formSchema = z.object({
  name: z
    .string()
    .min(1, 'Name must be at least 1 character.')
    .max(255, 'Name must be at most 255 characters.'),
});

interface CreateThreadFormProps {
  circleId?: string;
}

export function CreateThreadForm({ circleId }: CreateThreadFormProps = {}) {
  const navigate = useNavigate();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const createThreadMutation = usePostThreads();

  const form = useForm({
    defaultValues: {
      name: '',
    },
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: async ({ value }) => {
      setSubmitError(null);
      try {
        const thread = await createThreadMutation.mutateAsync({
          data: { name: value.name, circle_id: circleId },
        });
        await queryClient.invalidateQueries({
          queryKey: circleId
            ? getGetCirclesIdThreadsQueryKey(circleId)
            : getGetThreadsQueryKey(),
        });
        navigate({ to: '/thread/$id', params: { id: thread.id! } });
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
        id="create-thread-form"
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
                  form="create-thread-form"
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
