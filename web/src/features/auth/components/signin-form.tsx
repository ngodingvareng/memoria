import {
  useLogin,
  useGoogleLogin as useGoogleLoginMutation,
} from '@/lib/api/generated/auth/auth';
import { toSession } from '@/features/auth/lib/to-session';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSeparator,
} from '@/components/ui/field';
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from '@/components/ui/input-group';
import { ApiError, getApiErrorMessage } from '@/lib/api-client';
import { setSession } from '@/lib/session';
import { ViewIcon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { useForm } from '@tanstack/react-form';
import { Link, useNavigate } from '@tanstack/react-router';
import { GoogleLogin } from '@react-oauth/google';
import { useState } from 'react';
import * as z from 'zod';

const formSchema = z.object({
  email: z
    .string()
    .min(1, 'Email is required.')
    .email('Enter a valid email address.'),
  password: z.string().min(1, 'Password is required.'),
});

export function SigninForm() {
  const navigate = useNavigate();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const loginMutation = useLogin();
  const googleLoginMutation = useGoogleLoginMutation();

  const form = useForm({
    defaultValues: { email: '', password: '' },
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: async ({ value }) => {
      setSubmitError(null);
      try {
        const data = await loginMutation.mutateAsync({
          data: { email: value.email, password: value.password },
        });
        setSession(toSession(data));
        navigate({ to: data.user?.username ? '/' : '/welcome' });
      } catch (err) {
        setSubmitError(
          err instanceof ApiError && err.status === 401
            ? 'Incorrect email or password.'
            : getApiErrorMessage(err, 'Something went wrong. Please try again.')
        );
      }
    },
  });

  return (
    <form
      id="signin-form"
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
    >
      <FieldGroup>
        <Field className="items-center">
          <GoogleLogin
            theme="outline"
            shape="circle"
            size="large"
            text="signin_with"
            onSuccess={async (credentialResponse) => {
              if (!credentialResponse.credential) return;
              setSubmitError(null);
              try {
                const data = await googleLoginMutation.mutateAsync({
                  data: {
                    id_token: credentialResponse.credential,
                  },
                });
                setSession(toSession(data));
                // A new account created on the fly by Google login has
                // no username yet, exactly like a fresh Register —
                // route it through the same mandatory picker.
                navigate({
                  to: data.user?.username ? '/' : '/welcome',
                });
              } catch (err) {
                setSubmitError(
                  err instanceof ApiError && err.status === 409
                    ? 'An account with this email already exists. Log in with your password instead.'
                    : getApiErrorMessage(
                        err,
                        'Something went wrong. Please try again.'
                      )
                );
              }
            }}
            onError={() =>
              setSubmitError('Google sign-in failed. Please try again.')
            }
          />
        </Field>
        <FieldSeparator className="*:data-[slot=field-separator-content]:bg-card">
          Or continue with
        </FieldSeparator>
        <form.Field name="email">
          {(field) => {
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
        </form.Field>

        <form.Field name="password">
          {(field) => {
            const isInvalid =
              field.state.meta.isTouched && !field.state.meta.isValid;
            return (
              <Field data-invalid={isInvalid}>
                <FieldLabel htmlFor={field.name}>Password</FieldLabel>
                <InputGroup className="h-11!">
                  <InputGroupInput
                    id={field.name}
                    name={field.name}
                    type="password"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    aria-invalid={isInvalid}
                    autoComplete="off"
                    placeholder="••••••••"
                    className="text-lg!"
                  />
                  <InputGroupAddon align="inline-end">
                    <InputGroupButton size="icon-sm">
                      <HugeiconsIcon icon={ViewIcon} />
                    </InputGroupButton>
                  </InputGroupAddon>
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
              <Button type="submit" form="signin-form" disabled={isSubmitting}>
                Sign in
              </Button>
            )}
          </form.Subscribe>
        </Field>

        <Field>
          <FieldDescription>
            <Link to="/forgot-password">Forgot password?</Link>
          </FieldDescription>
          <FieldDescription>
            New to Memoria? <Link to="/signup">Sign up</Link>
          </FieldDescription>
        </Field>
      </FieldGroup>
    </form>
  );
}
