export interface ApiFieldError {
  field: string;
  tag: string;
  message: string;
}

// Mirrors api's dto.WebResponse[T] — data is present on success,
// errors is either a validator field-error array (400) or a plain
// message string (every other AppError), see
// api/internal/delivery/rest/middleware/error_handler.go.
interface WebResponse<T> {
  code: number;
  message: string;
  data?: T;
  errors?: ApiFieldError[] | string;
}

export class ApiError extends Error {
  status: number;
  fieldErrors?: ApiFieldError[];

  constructor(status: number, message: string, fieldErrors?: ApiFieldError[]) {
    super(message);
    this.status = status;
    this.fieldErrors = fieldErrors;
  }
}

// Joins validator field errors into one string when present, otherwise
// falls back to the error's own message; non-ApiErrors (network failures,
// unexpected throws) get the caller-supplied fallback instead.
export function getApiErrorMessage(err: unknown, fallback: string): string {
  if (!(err instanceof ApiError)) {
    return fallback;
  }
  return err.fieldErrors?.map((e) => e.message).join(' ') ?? err.message;
}

// apiFetch talks to api/'s REST API. credentials: 'include' is required
// on every call so the httpOnly refresh cookie (scoped to /auth) rides
// along; accessToken is attached as a Bearer header when provided.
export async function apiFetch<T>(
  path: string,
  options: RequestInit & { accessToken?: string } = {}
): Promise<T> {
  const { accessToken, headers, ...rest } = options;

  // FormData bodies (file uploads) need the browser to set their own
  // multipart Content-Type with the boundary — forcing application/json
  // here would send the wrong header and break the server's form parser.
  const isFormData = rest.body instanceof FormData;

  const res = await fetch(`${import.meta.env.VITE_API_URL}${path}`, {
    ...rest,
    credentials: 'include',
    headers: {
      ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
      ...headers,
    },
  });

  const body = (await res.json().catch(() => null)) as WebResponse<T> | null;

  if (!res.ok) {
    const fieldErrors = Array.isArray(body?.errors) ? body.errors : undefined;
    throw new ApiError(
      res.status,
      body?.message ?? 'Request failed',
      fieldErrors
    );
  }

  return body?.data as T;
}
