import { Skeleton } from '@/components/ui/skeleton';
import Wrapper from '@/components/wrapper';
import { DeleteAccountDialog, ProfileImageUploader } from '@/features/users';
import { useGetUserByID } from '@/lib/api/generated/users/users';
import { useSession } from '@/lib/session';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_user/user/account')({
  component: RouteComponent,
});

function RouteComponent() {
  const session = useSession();
  const userQuery = useGetUserByID(session?.user.id ?? '', {
    query: { enabled: !!session?.user.id },
  });

  if (userQuery.isPending) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton className="size-20 rounded-full" />
        <Skeleton className="h-6 w-48" />
      </div>
    );
  }

  const name = userQuery.data?.name ?? session?.user.name ?? '';

  return (
    <Wrapper fullWidth>
      <div className="flex max-w-xl flex-col gap-6">
        <ProfileImageUploader
          imagePath={userQuery.data?.image_path}
          name={name}
        />
        <div>
          <p className="text-lg font-medium">{name}</p>
          {session?.user.username && (
            <p className="text-muted-foreground">@{session.user.username}</p>
          )}
        </div>

        <div className="border-destructive/30 flex flex-col gap-2 rounded-2xl border p-4">
          <h2 className="font-heading text-lg font-semibold tracking-tight">
            Danger zone
          </h2>
          <p className="text-muted-foreground text-sm">
            Deleting your account removes your Moments, Threads, and photos.
            This can't be undone.
          </p>
          <DeleteAccountDialog />
        </div>
      </div>
    </Wrapper>
  );
}
