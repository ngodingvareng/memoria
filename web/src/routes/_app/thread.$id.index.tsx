import { DateTimeDialog } from '@/components/dialogs/datetime-dialog';
import { ShareDialog } from '@/components/dialogs/share-dialog';
import { Alert, AlertDescription } from '@/components/ui/alert';
import Wrapper from '@/components/wrapper';
import {
  MomentInput,
  MomentList,
  toDatetimeLocalValue,
  toRFC3339WithOffset,
  type MomentCardParam,
  type MomentDraft,
} from '@/features/moments';
import { ThreadHeader, ThreadHero } from '@/features/threads';
import { ApiError } from '@/lib/api-client';
import {
  getGetThreadsIdMomentsQueryKey,
  useDeleteMomentsId,
  useGetThreadsIdMoments,
  usePostMoments,
  usePostMomentsIdImages,
  usePutMomentsId,
} from '@/lib/api/generated/moments/moments';
import {
  useGetThreadsId,
  useGetThreadsIdImages,
} from '@/lib/api/generated/threads/threads';
import { queryClient } from '@/lib/query-client';
import { useSession } from '@/lib/session';
import { createFileRoute } from '@tanstack/react-router';
import React from 'react';

export const Route = createFileRoute('/_app/thread/$id/')({
  component: RouteComponent,
});

function RouteComponent() {
  const { id } = Route.useParams();
  const session = useSession();
  const threadQuery = useGetThreadsId(id);
  const imagesQuery = useGetThreadsIdImages(id);
  const momentsQuery = useGetThreadsIdMoments(id);
  const createMoment = usePostMoments();
  const uploadMomentImage = usePostMomentsIdImages();
  const updateMoment = usePutMomentsId();
  const deleteMoment = useDeleteMomentsId();

  const [openShareDialog, setOpenShareDialog] = React.useState(false);
  const [openTimeDialog, setOpenTimeDialog] = React.useState(false);
  const [isReadMode, setIsReadMode] = React.useState(false);
  const [publishError, setPublishError] = React.useState<string | null>(null);

  const invalidateMoments = () =>
    queryClient.invalidateQueries({
      queryKey: getGetThreadsIdMomentsQueryKey(id),
    });

  const moments: MomentCardParam[] = (momentsQuery.data?.moments ?? []).map(
    (moment) => ({
      id: moment.id,
      user: {
        name: session?.user.name ?? '',
        username: session?.user.username ?? '',
        imageSrc: '',
        imageAlt: session?.user.name ?? '',
      },
      thread: { name: threadQuery.data?.name ?? '' },
      colorHex: moment.color_hex,
      content: moment.note ?? '',
      createdAt: new Date(moment.created_at ?? moment.occurred_at ?? 0),
      capturedAt: new Date(moment.occurred_at ?? 0),
      isOwnedByCurrentUser: true,
    })
  );

  const handlePublish = async (draft: MomentDraft) => {
    setPublishError(null);
    try {
      const { occurredAt, offsetMinutes } = toRFC3339WithOffset(
        toDatetimeLocalValue(new Date())
      );
      const moment = await createMoment.mutateAsync({
        data: {
          thread_id: id,
          occurred_at: occurredAt,
          occurred_utc_offset_minutes: offsetMinutes,
          note: draft.note || undefined,
          color_hex: draft.colorHex || undefined,
        },
      });

      if (draft.images.length > 0 && moment.id) {
        await Promise.all(
          draft.images.map((image) =>
            uploadMomentImage.mutateAsync({ id: moment.id!, data: { image } })
          )
        );
      }

      await invalidateMoments();
    } catch (err) {
      setPublishError(
        err instanceof ApiError
          ? err.message
          : 'Failed to publish this moment. Please try again.'
      );
    }
  };

  const findMoment = (momentId: string) =>
    momentsQuery.data?.moments?.find((moment) => moment.id === momentId);

  const handleEditNote = async (momentId: string, note: string) => {
    const existing = findMoment(momentId);
    if (!existing?.occurred_at) return;
    await updateMoment.mutateAsync({
      id: momentId,
      data: {
        thread_id: id,
        occurred_at: existing.occurred_at,
        occurred_utc_offset_minutes: existing.occurred_utc_offset_minutes,
        note,
        color_hex: existing.color_hex,
        place_name: existing.place_name,
        latitude: existing.latitude,
        longitude: existing.longitude,
      },
    });
    await invalidateMoments();
  };

  const handleEditColor = async (momentId: string, colorHex: string) => {
    const existing = findMoment(momentId);
    if (!existing?.occurred_at) return;
    await updateMoment.mutateAsync({
      id: momentId,
      data: {
        thread_id: id,
        occurred_at: existing.occurred_at,
        occurred_utc_offset_minutes: existing.occurred_utc_offset_minutes,
        note: existing.note,
        color_hex: colorHex || undefined,
        place_name: existing.place_name,
        latitude: existing.latitude,
        longitude: existing.longitude,
      },
    });
    await invalidateMoments();
  };

  const handleDelete = async (momentId: string) => {
    await deleteMoment.mutateAsync({ id: momentId });
    await invalidateMoments();
  };

  return (
    <>
      <Wrapper>
        <ThreadHero
          isReadMode={isReadMode}
          imageUrl={imagesQuery.data?.[0]?.url}
          imageAlt={threadQuery.data?.name ?? ''}
          colorHex={threadQuery.data?.color_hex}
        />
        <ThreadHeader
          threadId={id}
          threadName={threadQuery.data?.name ?? ''}
          circleId={threadQuery.data?.circle_id}
          isReadMode={isReadMode}
          onToggleMode={() => setIsReadMode(!isReadMode)}
          onShare={() => setOpenShareDialog(true)}
        />
      </Wrapper>

      <Wrapper>
        <MomentList
          moments={moments}
          isEditMode={!isReadMode}
          onEditNote={handleEditNote}
          onEditColor={handleEditColor}
          onDelete={handleDelete}
        />
      </Wrapper>

      {moments.length > 0 && (
        <Wrapper className="flex justify-center text-center items-center p-0">
          <p className="text-muted-foreground text-xl">End of history</p>
        </Wrapper>
      )}

      {moments.length == 0 && (
        <Wrapper className="flex justify-center text-center items-center p-0">
          <p className="text-muted-foreground text-xl">
            Start your story by adding the first moment
          </p>
        </Wrapper>
      )}

      {!isReadMode && (
        <div className="flex flex-col gap-2">
          {publishError && (
            <Wrapper className="pb-0">
              <Alert variant="destructive" className="mx-auto max-w-5xl">
                <AlertDescription>{publishError}</AlertDescription>
              </Alert>
            </Wrapper>
          )}
          <MomentInput
            onOpenTimeDialog={() => setOpenTimeDialog(true)}
            onPublish={handlePublish}
          />
        </div>
      )}

      <DateTimeDialog
        open={openTimeDialog}
        onOpenChange={setOpenTimeDialog}
        onSave={() => setOpenTimeDialog(false)}
      />

      <ShareDialog open={openShareDialog} onOpenChange={setOpenShareDialog} />
    </>
  );
}
