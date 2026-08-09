import { apiFetch, ApiError } from '@/lib/api-client';
import { getSession, setSession } from '@/lib/session';
import { restoreSession } from '@/features/auth/api/session-bootstrap';

// Dedupes concurrent 401s into a single /auth/refresh call instead of one
// per failed request.
let refreshInFlight: Promise<boolean> | null = null;

function refreshOnce(): Promise<boolean> {
  if (!refreshInFlight) {
    refreshInFlight = restoreSession().finally(() => {
      refreshInFlight = null;
    });
  }
  return refreshInFlight;
}

// The mutator Orval's generated hooks call under the hood. Delegates the
// actual request to apiFetch (WebResponse envelope unwrap, ApiError, the
// refresh cookie via credentials: 'include' all already live there) and
// adds what generated hooks need on top: automatic Bearer token injection
// from the session store, and a refresh-and-retry-once on 401.
export async function apiMutator<T>(
  url: string,
  options: RequestInit = {}
): Promise<T> {
  const session = getSession();

  try {
    return await apiFetch<T>(url, { ...options, accessToken: session?.accessToken });
  } catch (err) {
    if (!(err instanceof ApiError) || err.status !== 401 || !session) {
      throw err;
    }

    const refreshed = await refreshOnce();
    if (!refreshed) {
      setSession(null);
      throw err;
    }

    return apiFetch<T>(url, { ...options, accessToken: getSession()?.accessToken });
  }
}
