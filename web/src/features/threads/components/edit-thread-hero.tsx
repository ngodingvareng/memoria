import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { ThreadHero } from './thread-hero';
import { ApiError } from '@/lib/api-client';
import {
  getGetThreadsIdImagesQueryKey,
  useDeleteThreadsIdImagesImageId,
  usePostThreadsIdImages,
} from '@/lib/api/generated/threads/threads';
import { queryClient } from '@/lib/query-client';
import { Camera01Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { useRef, useState } from 'react';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadImageResponse } from '@/lib/api/generated/models';

interface EditThreadHeroProps {
  threadId: string;
  threadName: string;
  heroImage?: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadImageResponse;
  colorHex?: string;
}

export function EditThreadHero({
  threadId,
  threadName,
  heroImage,
  colorHex,
}: EditThreadHeroProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [error, setError] = useState<string | null>(null);
  const uploadImage = usePostThreadsIdImages();
  const deleteImage = useDeleteThreadsIdImagesImageId();
  const isPending = uploadImage.isPending || deleteImage.isPending;

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) return;

    setError(null);
    try {
      if (heroImage?.id) {
        await deleteImage.mutateAsync({ id: threadId, imageId: heroImage.id });
      }
      await uploadImage.mutateAsync({ id: threadId, data: { image: file } });
      await queryClient.invalidateQueries({
        queryKey: getGetThreadsIdImagesQueryKey(threadId),
      });
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Failed to update the cover image. Please try again.'
      );
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <div className="relative">
        <ThreadHero
          isReadMode={false}
          imageUrl={heroImage?.url}
          imageAlt={threadName}
          colorHex={colorHex}
        />

        <Button
          type="button"
          variant="secondary"
          size="icon-sm"
          className="absolute top-3 right-3 z-40"
          disabled={isPending}
          onClick={() => fileInputRef.current?.click()}
        >
          <HugeiconsIcon icon={Camera01Icon} strokeWidth={2} />
          <span className="sr-only">Change cover image</span>
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
