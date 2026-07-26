import { ColorPickerDialog } from '@/components/dialogs/color-picker-dialog';
import { DateTimeDialog } from '@/components/dialogs/datetime-dialog';
import { ShareDialog } from '@/components/dialogs/share-dialog';

import Wrapper from '@/components/wrapper';
import {
  ActivityCaptureInput,
  ActivityCaptureList,
  ActivityHeader,
  ActivityHero,
} from '@/features/activities';
import { dummyNotes } from '@/lib/dummies';
import { createFileRoute } from '@tanstack/react-router';
import React from 'react';

export const Route = createFileRoute('/_app/activity/$id/')({
  component: RouteComponent,
});

function RouteComponent() {
  const [openCostumizationDialog, setOpenCostumizationDialog] =
    React.useState(false);
  const [openShareDialog, setOpenShareDialog] = React.useState(false);
  const [openTimeDialog, setOpenTimeDialog] = React.useState(false);
  const [isReadMode, setIsReadMode] = React.useState(false);

  const noteCount = dummyNotes.length;

  return (
    <>
      <Wrapper>
        <ActivityHero
          isReadMode={isReadMode}
          imageUrl="https://images.unsplash.com/photo-1604076850742-4c7221f3101b?q=80&w=1887&auto=format&fit=crop&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D"
          imageAlt="Photo by mymind on Unsplash"
        />
        <ActivityHeader
          isReadMode={isReadMode}
          noteCount={noteCount}
          onToggleMode={() => setIsReadMode(!isReadMode)}
          onShare={() => setOpenShareDialog(true)}
        />
      </Wrapper>

      <Wrapper>
        <ActivityCaptureList
          notes={dummyNotes}
          onSetColor={() => setOpenCostumizationDialog(true)}
          onShare={() => setOpenShareDialog(true)}
          onEdit={(id) => console.log('Editing note', id)}
          onFavorite={(id) => console.log('Favoriting note', id)}
        />
      </Wrapper>

      {dummyNotes.length > 0 && (
        <Wrapper className="h-dvh flex justify-center text-center items-center p-0">
          <p className="text-muted-foreground text-xl">End of history</p>
        </Wrapper>
      )}

      {!isReadMode && (
        <ActivityCaptureInput
          onOpenTimeDialog={() => setOpenTimeDialog(true)}
          onPublish={() => console.log('Publishing...')}
        />
      )}

      <DateTimeDialog
        open={openTimeDialog}
        onOpenChange={setOpenTimeDialog}
        onSave={() => setOpenTimeDialog(false)}
      />

      <ColorPickerDialog
        open={openCostumizationDialog}
        onOpenChange={setOpenCostumizationDialog}
        onSave={() => setOpenCostumizationDialog(false)}
      />

      <ShareDialog open={openShareDialog} onOpenChange={setOpenShareDialog} />
    </>
  );
}
