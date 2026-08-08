import { apiFetch } from '@/lib/api-client';
import type { SessionUser } from '@/lib/session';

interface CheckUsernameAvailabilityResponseDto {
  available: boolean;
}

export async function checkUsernameAvailability(
  username: string,
  accessToken: string
): Promise<boolean> {
  const data = await apiFetch<CheckUsernameAvailabilityResponseDto>(
    `/users/username-availability?username=${encodeURIComponent(username)}`,
    { accessToken }
  );
  return data.available;
}

// Mirrors api's dto.UserResponse.
interface UserResponseDto {
  id: string;
  name: string;
  username: string | null;
  email: string;
  email_verified: boolean;
  created_at: string;
}

function toSessionUser(data: UserResponseDto): SessionUser {
  return {
    id: data.id,
    name: data.name,
    username: data.username,
    email: data.email,
    emailVerified: data.email_verified,
    createdAt: data.created_at,
  };
}

export async function setUsername(
  username: string,
  accessToken: string
): Promise<SessionUser> {
  const data = await apiFetch<UserResponseDto>('/users/me/username', {
    method: 'PATCH',
    body: JSON.stringify({ username }),
    accessToken,
  });
  return toSessionUser(data);
}
