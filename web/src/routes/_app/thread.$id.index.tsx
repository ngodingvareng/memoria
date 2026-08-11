import { DateTimeDialog } from '@/components/dialogs/datetime-dialog';
import { ShareDialog } from '@/components/dialogs/share-dialog';
import { Alert, AlertDescription } from '@/components/ui/alert';
import Wrapper from '@/components/wrapper';
import {
  showMentionShareOffers,
  syncMomentMentionsFromText,
} from '@/features/mentions';
import {
  MomentFeedList,
  MomentInput,
  toDatetimeLocalValue,
  toRFC3339WithOffset,
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
import { createFileRoute } from '@tanstack/react-router';
import React from 'react';

export const Route = createFileRoute('/_app/thread/$id/')({
  component: RouteComponent,
});

function RouteComponent() {
  const { id } = Route.useParams();
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
  const [captureError, setCaptureError] = React.useState<string | null>(null);

  const invalidateMoments = () =>
    queryClient.invalidateQueries({
      queryKey: getGetThreadsIdMomentsQueryKey(id),
    });

  const moments = momentsQuery.data?.moments ?? [];
  // Only a Circle-owned thread's own page shows the creator/thread-name
  // header — a personal thread is always "mine", so it'd be redundant.
  const showHeader = !!threadQuery.data?.circle_id;

  const handleCapture = async (draft: MomentDraft) => {
    setCaptureError(null);
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
            uploadMomentImage.mutateAsync({
              id: moment.id!,
              data: { image },
            })
          )
        );
      }

      if (moment.id && draft.note) {
        const { shareOffers } = await syncMomentMentionsFromText(
          moment.id,
          draft.note
        );
        if (shareOffers.length > 0) {
          showMentionShareOffers(moment.id, shareOffers);
        }
      }

      await invalidateMoments();
    } catch (err) {
      setCaptureError(
        err instanceof ApiError
          ? err.message
          : 'Failed to capture this moment. Please try again.'
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
    const { shareOffers } = await syncMomentMentionsFromText(momentId, note);
    if (shareOffers.length > 0) {
      showMentionShareOffers(momentId, shareOffers);
    }
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
      <div className="flex flex-col">
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
          <MomentFeedList
            moments={moments}
            showHeader={showHeader}
            onEditNote={handleEditNote}
            onEditColor={handleEditColor}
            onDelete={handleDelete}
          />
        </Wrapper>

        {moments.length > 0 && (
          <Wrapper className="flex items-center justify-center p-0 text-center">
            <p className="text-muted-foreground text-xl">End of history</p>
          </Wrapper>
        )}

        {moments.length == 0 && (
          <Wrapper className="flex items-center justify-center p-0 text-center">
            <p className="text-muted-foreground text-xl">
              Start your story by adding the first moment
            </p>
          </Wrapper>
        )}

        {!isReadMode && (
          <>
            {captureError && (
              <Wrapper className="pb-0">
                <Alert variant="destructive" className="mx-auto max-w-5xl">
                  <AlertDescription>{captureError}</AlertDescription>
                </Alert>
              </Wrapper>
            )}
            <MomentInput
              onOpenTimeDialog={() => setOpenTimeDialog(true)}
              onCapture={handleCapture}
            />
          </>
        )}
      </div>

      <DateTimeDialog
        open={openTimeDialog}
        onOpenChange={setOpenTimeDialog}
        onSave={() => setOpenTimeDialog(false)}
      />

      <ShareDialog open={openShareDialog} onOpenChange={setOpenShareDialog} />
    </>
  );
}
