import { Alert, AlertDescription } from '@/components/ui/alert';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { ApiError } from '@/lib/api-client';
import {
  getGetUserByIDQueryKey,
  useUploadProfileImage,
} from '@/lib/api/generated/users/users';
import { queryClient } from '@/lib/query-client';
import { useSession } from '@/lib/session';
import { Camera01Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { useRef, useState } from 'react';

interface ProfileImageUploaderProps {
  imagePath?: string;
  name: string;
}

// Same hidden-input-behind-a-camera-icon-button pattern as
// EditThreadHero, but a single atomic upload call — the backend
// replaces the previous photo itself, so there's no separate
// delete-then-upload from here.
export function ProfileImageUploader({
  imagePath,
  name,
}: ProfileImageUploaderProps) {
  const session = useSession();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [error, setError] = useState<string | null>(null);
  const uploadImage = useUploadProfileImage();

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) return;

    setError(null);
    try {
      await uploadImage.mutateAsync({ data: { image: file } });
      if (session?.user.id) {
        await queryClient.invalidateQueries({
          queryKey: getGetUserByIDQueryKey(session.user.id),
        });
      }
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Failed to update your photo. Please try again.'
      );
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <div className="relative inline-block w-fit">
        <Avatar size="xl">
          <AvatarImage src={imagePath} alt={name} />
          <AvatarFallback>{name.charAt(0).toUpperCase()}</AvatarFallback>
        </Avatar>
        <Button
          type="button"
          variant="secondary"
          size="icon-sm"
          className="absolute -right-1 -bottom-1"
          disabled={uploadImage.isPending}
          onClick={() => fileInputRef.current?.click()}
        >
          <HugeiconsIcon icon={Camera01Icon} strokeWidth={2} />
          <span className="sr-only">Change profile photo</span>
        </Button>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={handleFileChange}
        />
      </div>
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
    </div>
  );
}
