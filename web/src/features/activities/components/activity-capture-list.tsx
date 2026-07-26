import type { Note } from '@/types/note';
import { ActivityCaptureCard } from './activity-capture-card';

interface ActivityCaptureListProps {
  notes: Note[];
  onSetColor: (id: number) => void;
  onShare: (id: number) => void;
  onEdit: (id: number) => void;
  onFavorite: (id: number) => void;
}

export function ActivityCaptureList({
  notes,
  onSetColor,
  onShare,
  onEdit,
  onFavorite,
}: ActivityCaptureListProps) {
  return (
    <div className="flex flex-col gap-6 w-full bottom-0">
      {notes.map((note) => (
        <ActivityCaptureCard
          key={note.id}
          note={note}
          onSetColor={onSetColor}
          onShare={onShare}
          onEdit={onEdit}
          onFavorite={onFavorite}
        />
      ))}
    </div>
  );
}
