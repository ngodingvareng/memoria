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

// apiFetch talks to api/'s REST API. credentials: 'include' is required
// on every call so the httpOnly refresh cookie (scoped to /auth) rides
// along; accessToken is attached as a Bearer header when provided.
export async function apiFetch<T>(
  path: string,
  options: RequestInit & { accessToken?: string } = {}
): Promise<T> {
  const { accessToken, headers, ...rest } = options;

  const res = await fetch(`${import.meta.env.VITE_API_URL}${path}`, {
    ...rest,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
      ...headers,
    },
  });

  const body = (await res.json().catch(() => null)) as WebResponse<T> | null;

  if (!res.ok) {
    const fieldErrors = Array.isArray(body?.errors) ? body.errors : undefined;
    throw new ApiError(res.status, body?.message ?? 'Request failed', fieldErrors);
  }

  return body?.data as T;
}
