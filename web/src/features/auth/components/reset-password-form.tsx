import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field';
import { InputGroup, InputGroupInput } from '@/components/ui/input-group';
import { ApiError, getApiErrorMessage } from '@/lib/api-client';
import { useResetPassword } from '@/lib/api/generated/auth/auth';
import { useForm } from '@tanstack/react-form';
import { useNavigate } from '@tanstack/react-router';
import { useState } from 'react';
import * as z from 'zod';

interface ResetPasswordFormProps {
  email: string;
  token: string;
}

const formSchema = z.object({
  newPassword: z.string().min(8, 'Password must be at least 8 characters.'),
});

export function ResetPasswordForm({ email, token }: ResetPasswordFormProps) {
  const navigate = useNavigate();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const resetPassword = useResetPassword();

  const form = useForm({
    defaultValues: { newPassword: '' },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value }) => {
      setSubmitError(null);
      try {
        await resetPassword.mutateAsync({
          data: { email, token, new_password: value.newPassword },
        });
        navigate({ to: '/signin' });
      } catch (err) {
        setSubmitError(
          err instanceof ApiError && err.status === 401
            ? 'This reset link is invalid or has expired. Request a new one.'
            : getApiErrorMessage(err, 'Something went wrong. Please try again.')
        );
      }
    },
  });

  return (
    <form
      id="reset-password-form"
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
    >
      <FieldGroup>
        <form.Field name="newPassword">
          {(field) => {
            const isInvalid =
              field.state.meta.isTouched && !field.state.meta.isValid;
            return (
              <Field data-invalid={isInvalid}>
                <FieldLabel htmlFor={field.name}>New password</FieldLabel>
                <InputGroup className="h-11!">
                  <InputGroupInput
                    id={field.name}
                    name={field.name}
                    type="password"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    aria-invalid={isInvalid}
                    autoComplete="new-password"
                    placeholder="••••••••"
                    className="text-lg!"
                  />
                </InputGroup>
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

        <Field>
          <form.Subscribe selector={(state) => state.isSubmitting}>
            {(isSubmitting) => (
              <Button
                type="submit"
                form="reset-password-form"
                disabled={isSubmitting}
              >
                Reset password
              </Button>
            )}
          </form.Subscribe>
        </Field>
      </FieldGroup>
    </form>
  );
}
