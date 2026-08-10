import { getApiErrorMessage } from '@/lib/api-client';
import * as React from 'react';

// Wraps a form's onSubmit body with the error/message boilerplate every
// create/edit form repeats: clear the previous error, run the submit
// logic, and surface ApiError (or a fallback) as a string on failure.
export function useFormSubmitError(
  fallback = 'Something went wrong. Please try again.'
) {
  const [error, setError] = React.useState<string | null>(null);

  const guard = async (submit: () => Promise<void>) => {
    setError(null);
    try {
      await submit();
    } catch (err) {
      setError(getApiErrorMessage(err, fallback));
    }
  };

  const clear = () => setError(null);

  return { error, guard, clear };
}
