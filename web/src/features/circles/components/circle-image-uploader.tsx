import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { CircleProfileImage } from './circle-profile-image';
import { ApiError } from '@/lib/api-client';
import {
  getGetCirclesIdQueryKey,
  usePostCirclesIdImage,
} from '@/lib/api/generated/circles/circles';
import { queryClient } from '@/lib/query-client';
import { Camera01Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { useRef, useState } from 'react';

interface CircleImageUploaderProps {
  circleId: string;
  imagePath?: string | null;
  name?: string | null;
}

// Admin-only in practice — the upload endpoint itself is the real gate
// (POST /circles/{id}/image, admin-only WHERE clause); this is only
// ever rendered on the settings page a non-admin can't reach anyway.
export function CircleImageUploader({
  circleId,
  imagePath,
  name,
}: CircleImageUploaderProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [error, setError] = useState<string | null>(null);
  const uploadImage = usePostCirclesIdImage();

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) return;

    setError(null);
    try {
      await uploadImage.mutateAsync({ id: circleId, data: { image: file } });
      await queryClient.invalidateQueries({
        queryKey: getGetCirclesIdQueryKey(circleId),
      });
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : "Failed to update this circle's photo. Please try again."
      );
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <div className="relative inline-block w-fit">
        <CircleProfileImage
          circle={{ name, image_path: imagePath }}
          size="xl"
        />
        <Button
          type="button"
          variant="secondary"
          size="icon-sm"
          className="absolute -right-1 -bottom-1"
          disabled={uploadImage.isPending}
          onClick={() => fileInputRef.current?.click()}
        >
          <HugeiconsIcon icon={Camera01Icon} strokeWidth={2} />
          <span className="sr-only">Change circle photo</span>
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
