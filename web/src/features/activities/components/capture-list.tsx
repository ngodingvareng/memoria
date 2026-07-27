import type { Note } from '@/types/note';
import { CaptureCard } from './capture-card';

interface CaptureListProps {
  notes: Note[];
  onSetColor: (id: number) => void;
  onShare: (id: number) => void;
  onEdit: (id: number) => void;
  onFavorite: (id: number) => void;
}

export function CaptureList({
  notes,
  onSetColor,
  onShare,
  onEdit,
  onFavorite,
}: CaptureListProps) {
  return (
    <div className="flex flex-col gap-6 w-full bottom-0">
      {notes.map((note) => (
        <CaptureCard
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
