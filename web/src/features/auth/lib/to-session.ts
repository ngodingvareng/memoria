import type {
  GithubComNgodingvarengMemoriaInternalDeliveryRestDtoLoginResponse,
  GithubComNgodingvarengMemoriaInternalDeliveryRestDtoUserResponse,
} from '@/lib/api/generated/models';
import type { Session, SessionUser } from '@/lib/session';

// Every field on the generated response types is optional (swag doesn't
// mark response DTOs as `required` the way it does request DTOs, since
// they carry no validate tags) — these are always populated in practice
// on a successful response, so asserting non-null here is safe.
export function toSessionUser(
  data: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoUserResponse
): SessionUser {
  return {
    id: data.id!,
    name: data.name!,
    username: data.username ?? null,
    email: data.email!,
    emailVerified: Boolean(data.email_verified),
    createdAt: data.created_at!,
  };
}

export function toSession(
  data: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoLoginResponse
): Session {
  return {
    accessToken: data.access_token!,
    user: toSessionUser(data.user!),
  };
}
