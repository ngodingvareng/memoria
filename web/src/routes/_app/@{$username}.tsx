import { Alert, AlertDescription } from '@/components/ui/alert';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import Wrapper from '@/components/wrapper';
import { ApiError } from '@/lib/api-client';
import {
  useGetUserByUsername,
  useMarkUserKnown,
} from '@/lib/api/generated/users/users';
import { useSession } from '@/lib/session';
import { createFileRoute } from '@tanstack/react-router';
import { useState } from 'react';

export const Route = createFileRoute('/_app/@{$username}')({
  component: RouteComponent,
});

function RouteComponent() {
  const { username } = Route.useParams();
  const session = useSession();
  const [error, setError] = useState<string | null>(null);
  const [known, setKnown] = useState(false);

  const profileQuery = useGetUserByUsername(username);
  const markKnown = useMarkUserKnown();

  const isOwnProfile = session?.user.username === username;

  const handleMarkKnown = async () => {
    setError(null);
    try {
      await markKnown.mutateAsync({ username });
      setKnown(true);
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Something went wrong. Please try again.'
      );
    }
  };

  if (profileQuery.isPending) {
    return (
      <Wrapper>
        <div className="flex gap-4">
          <Skeleton className="size-20 rounded-full" />
          <div className="flex flex-col gap-2">
            <Skeleton className="h-6 w-40" />
            <Skeleton className="h-5 w-28" />
          </div>
        </div>
      </Wrapper>
    );
  }

  if (profileQuery.isError || !profileQuery.data) {
    return (
      <Wrapper>
        <Alert variant="destructive">
          <AlertDescription>Couldn't find this user.</AlertDescription>
        </Alert>
      </Wrapper>
    );
  }

  const user = profileQuery.data;

  return (
    <Wrapper>
      <div className="flex gap-4">
        <Avatar className="size-20">
          <AvatarImage src={user.image_path ?? undefined} alt={user.name} />
          <AvatarFallback>{user.name?.charAt(0).toUpperCase()}</AvatarFallback>
        </Avatar>
        <div className="flex flex-col gap-1">
          <div>
            <p className="text-2xl/5 font-medium">{user.name}</p>
            {user.username && (
              <p className="text-muted-foreground text-lg/5">
                @{user.username}
              </p>
            )}
          </div>

          {user.bio && <p className="max-w-prose">{user.bio}</p>}

          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {!isOwnProfile && (
            <div>
              <Button
                onClick={handleMarkKnown}
                disabled={known || markKnown.isPending}
              >
                {known ? 'You know this person' : 'I know this person'}
              </Button>
            </div>
          )}
        </div>
      </div>
    </Wrapper>
  );
}
