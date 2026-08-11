import { Alert, AlertDescription } from '@/components/ui/alert';
import { Empty, EmptyDescription, EmptyHeader } from '@/components/ui/empty';
import { Lightbox } from '@/components/ui/lightbox';
import { InfiniteScrollSentinel } from '@/components/infinite-scroll-sentinel';
import Wrapper from '@/components/wrapper';
import {
  AlbumImageList,
  AlbumTimeline,
  formatAlbumImageCaption,
  groupAlbumImagesByDate,
  useAlbumImages,
} from '@/features/album';
import { createFileRoute } from '@tanstack/react-router';
import React from 'react';

export const Route = createFileRoute('/_app/album/')({
  component: RouteComponent,
});

function RouteComponent() {
  const albumQuery = useAlbumImages();
  const images =
    albumQuery.data?.pages.flatMap((page) => page.images ?? []) ?? [];
  const buckets = groupAlbumImagesByDate(images);

  const [isLightboxOpen, setIsLightboxOpen] = React.useState(false);
  const [lightboxIndex, setLightboxIndex] = React.useState(0);

  return (
    <Wrapper>
      <div className="flex flex-col gap-8">
        <h1 className="text-2xl font-semibold">Album</h1>

        {albumQuery.isError && (
          <Alert variant="destructive">
            <AlertDescription>
              Couldn't load your Album. Please try again.
            </AlertDescription>
          </Alert>
        )}

        {!albumQuery.isPending &&
          !albumQuery.isError &&
          images.length === 0 && (
            <Empty>
              <EmptyHeader>
                <EmptyDescription>No photos yet.</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}

        {buckets.length > 0 && (
          <AlbumTimeline
            albums={buckets.map((bucket) => ({
              date: bucket.date,
              content: (
                <AlbumImageList
                  images={bucket.images}
                  onImageClick={(image) => {
                    const index = images.findIndex((i) => i.id === image.id);
                    if (index === -1) return;
                    setLightboxIndex(index);
                    setIsLightboxOpen(true);
                  }}
                />
              ),
            }))}
          />
        )}

        <InfiniteScrollSentinel
          enabled={
            Boolean(albumQuery.hasNextPage) && !albumQuery.isFetchingNextPage
          }
          isLoading={albumQuery.isFetchingNextPage}
          onIntersect={() => albumQuery.fetchNextPage()}
        />
      </div>

      {images.length > 0 && (
        <Lightbox
          images={images.map((image) => ({
            src: image.url ?? '',
            alt: image.image_alt,
            caption: formatAlbumImageCaption(image),
          }))}
          open={isLightboxOpen}
          onOpenChange={setIsLightboxOpen}
          index={lightboxIndex}
          onIndexChange={setLightboxIndex}
          loop
        />
      )}
    </Wrapper>
  );
}
