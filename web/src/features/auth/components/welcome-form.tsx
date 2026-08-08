import {
  checkUsernameAvailability,
  setUsername,
} from '@/features/auth/api/user-api';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field';
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from '@/components/ui/input-group';
import { ApiError } from '@/lib/api-client';
import { getSession, setSession } from '@/lib/session';
import { useForm } from '@tanstack/react-form';
import { useNavigate } from '@tanstack/react-router';
import { useState } from 'react';
import * as z from 'zod';
import { HugeiconsIcon } from '@hugeicons/react';
import { AtIcon } from '@hugeicons/core-free-icons';

// Mirrors users.chk_users_username_format exactly.
const usernamePattern = /^[a-z0-9_.]{3,30}$/;

const usernameFieldSchema = z
  .string()
  .min(3, 'Username must be at least 3 characters.')
  .max(30, 'Username must be at most 30 characters.')
  .regex(
    usernamePattern,
    'Only lowercase letters, numbers, underscore, and dot are allowed.'
  );

const formSchema = z.object({ username: usernameFieldSchema });

export function WelcomeForm() {
  const navigate = useNavigate();
  const [submitError, setSubmitError] = useState<string | null>(null);

  const form = useForm({
    defaultValues: { username: '' },
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: async ({ value }) => {
      const session = getSession();
      if (!session) return;

      setSubmitError(null);
      try {
        const user = await setUsername(value.username, session.accessToken);
        setSession({ ...session, user });
        navigate({ to: '/' });
      } catch (err) {
        if (err instanceof ApiError) {
          setSubmitError(
            err.status === 409
              ? 'That username was just taken — try another.'
              : (err.fieldErrors?.map((e) => e.message).join(' ') ??
                  err.message)
          );
        } else {
          setSubmitError('Something went wrong. Please try again.');
        }
      }
    },
  });

  return (
    <form
      id="welcome-form"
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
    >
      <FieldGroup>
        <form.Field
          name="username"
          validators={{
            onChange: usernameFieldSchema,
            onChangeAsyncDebounceMs: 400,
            onChangeAsync: async ({ value }) => {
              // Malformed input is already flagged by the sync validator
              // above — no point spending a network round trip on it.
              if (!usernamePattern.test(value)) return undefined;

              const session = getSession();
              if (!session) return undefined;

              const available = await checkUsernameAvailability(
                value,
                session.accessToken
              );
              return available
                ? undefined
                : { message: 'That username is already taken.' };
            },
          }}
          children={(field) => {
            const isInvalid =
              field.state.meta.isTouched && !field.state.meta.isValid;
            const isChecking = field.state.meta.isValidating;
            const isAvailable =
              field.state.meta.isTouched &&
              !isInvalid &&
              !isChecking &&
              field.state.value.length > 0;

            return (
              <Field data-invalid={isInvalid}>
                <FieldLabel htmlFor={field.name}>Username</FieldLabel>
                <InputGroup className="h-11!">
                  <InputGroupAddon>
                    <HugeiconsIcon icon={AtIcon} />
                  </InputGroupAddon>
                  <InputGroupInput
                    id={field.name}
                    name={field.name}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) =>
                      field.handleChange(e.target.value.toLowerCase())
                    }
                    aria-invalid={isInvalid}
                    autoComplete="off"
                    placeholder="budisantoso"
                    className="text-lg!"
                  />
                </InputGroup>
                {isInvalid ? (
                  <FieldError errors={field.state.meta.errors} />
                ) : (
                  <FieldDescription>
                    {isChecking && 'Checking availability…'}
                    {isAvailable && 'This username is available.'}
                  </FieldDescription>
                )}
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
            selector={(state) => [
              state.isSubmitting,
              state.isValidating,
              state.isFieldsValid,
              state.values.username,
            ]}
            children={([
              isSubmitting,
              isValidating,
              isFieldsValid,
              username,
            ]) => (
              <Button
                type="submit"
                form="welcome-form"
                disabled={
                  Boolean(isSubmitting) ||
                  Boolean(isValidating) ||
                  !isFieldsValid ||
                  !username
                }
              >
                Use this username
              </Button>
            )}
          />
        </Field>
      </FieldGroup>
    </form>
  );
}
