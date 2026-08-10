import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field';
import { InputGroup, InputGroupInput } from '@/components/ui/input-group';
import { useFormSubmitError } from '@/hooks/use-form-submit-error';
import { useForgotPassword } from '@/lib/api/generated/auth/auth';
import { useForm } from '@tanstack/react-form';
import { Link } from '@tanstack/react-router';
import { useState } from 'react';
import * as z from 'zod';

const formSchema = z.object({
  email: z
    .string()
    .min(1, 'Email is required.')
    .email('Enter a valid email address.'),
});

export function ForgotPasswordForm() {
  const { error: submitError, guard } = useFormSubmitError();
  const [sent, setSent] = useState(false);
  const forgotPassword = useForgotPassword();

  const form = useForm({
    defaultValues: { email: '' },
    validators: { onSubmit: formSchema },
    onSubmit: ({ value }) =>
      guard(async () => {
        await forgotPassword.mutateAsync({ data: { email: value.email } });
        // The backend responds identically whether or not the email is
        // registered — this screen must too, or it becomes an oracle
        // for probing which emails have accounts.
        setSent(true);
      }),
  });

  if (sent) {
    return (
      <Alert>
        <AlertDescription>
          If that email is registered, a reset link is on its way. Check your
          inbox.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <form
      id="forgot-password-form"
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
    >
      <FieldGroup>
        <form.Field
          name="email"
          children={(field) => {
            const isInvalid =
              field.state.meta.isTouched && !field.state.meta.isValid;
            return (
              <Field data-invalid={isInvalid}>
                <FieldLabel htmlFor={field.name}>Email</FieldLabel>
                <InputGroup className="h-11!">
                  <InputGroupInput
                    id={field.name}
                    name={field.name}
                    type="email"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    aria-invalid={isInvalid}
                    autoComplete="off"
                    placeholder="memoria@example.com"
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

        <Field>
          <form.Subscribe
            selector={(state) => state.isSubmitting}
            children={(isSubmitting) => (
              <Button
                type="submit"
                form="forgot-password-form"
                disabled={isSubmitting}
              >
                Send reset link
              </Button>
            )}
          />
        </Field>

        <Field>
          <FieldDescription>
            <Link to="/signin">Back to sign in</Link>
          </FieldDescription>
        </Field>
      </FieldGroup>
    </form>
  );
}
