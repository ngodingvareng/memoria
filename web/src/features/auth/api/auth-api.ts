import { apiFetch } from '@/lib/api-client';
import type { Session } from '@/lib/session';

export interface RegisterAccountInput {
  name: string;
  email: string;
  password: string;
}

// Mirrors api's dto.LoginResponse — returned by /auth/register too,
// since registering also starts a session (see auth_usecase.go's
// Register).
interface LoginResponseDto {
  user: {
    id: string;
    name: string;
    username: string | null;
    email: string;
    email_verified: boolean;
    created_at: string;
  };
  access_token: string;
  expires_in: number;
}

function toSession(data: LoginResponseDto): Session {
  return {
    accessToken: data.access_token,
    user: {
      id: data.user.id,
      name: data.user.name,
      username: data.user.username,
      email: data.user.email,
      emailVerified: data.user.email_verified,
      createdAt: data.user.created_at,
    },
  };
}

export async function registerAccount(input: RegisterAccountInput): Promise<Session> {
  const data = await apiFetch<LoginResponseDto>('/auth/register', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  return toSession(data);
}
