import { useSyncExternalStore } from 'react';

// Mirrors api's dto.UserResponse.
export interface SessionUser {
  id: string;
  name: string;
  username: string | null;
  email: string;
  emailVerified: boolean;
  createdAt: string;
}

export interface Session {
  // Held in memory only, never localStorage — matches api's own comment
  // on dto.LoginResponse.AccessToken. Lost on page reload by design; the
  // refresh cookie is what survives a reload, not this.
  accessToken: string;
  user: SessionUser;
}

let session: Session | null = null;
const listeners = new Set<() => void>();

export function getSession(): Session | null {
  return session;
}

export function setSession(next: Session | null): void {
  session = next;
  listeners.forEach((listener) => listener());
}

function subscribe(callback: () => void): () => void {
  listeners.add(callback);
  return () => listeners.delete(callback);
}

// useSyncExternalStore keeps this store usable both inside React render
// (here) and outside it — route beforeLoad guards call getSession()
// directly, since they run outside any component.
export function useSession(): Session | null {
  return useSyncExternalStore(subscribe, getSession);
}
