import { apiFetch } from '@/lib/api-client';
import { setSession } from '@/lib/session';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoLoginResponse } from '@/lib/api/generated/models';
import { toSession } from '@/features/auth/lib/to-session';

// Exchanges the httpOnly refresh cookie for a fresh access token. Used
// both at app bootstrap (src/main.tsx, to survive a page reload without
// forcing a re-login) and by the API mutator's 401 retry.
//
// Deliberately calls apiFetch directly instead of the generated
// `refreshToken()` fetcher — that fetcher goes through apiMutator, and
// apiMutator calls this function on a 401, so routing through it here
// would risk apiMutator recursively retrying its own refresh call.
//
// Never throws: a missing/expired refresh cookie just means "not logged
// in", which is a normal outcome here, not an error.
export async function restoreSession(): Promise<boolean> {
  try {
    const data =
      await apiFetch<GithubComNgodingvarengMemoriaInternalDeliveryRestDtoLoginResponse>(
        '/auth/refresh',
        { method: 'POST' }
      );
    setSession(toSession(data));
    return true;
  } catch {
    setSession(null);
    return false;
  }
}
