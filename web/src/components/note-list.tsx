'use client';

import React from 'react';
import { NoteCard } from './note-card';
import type { Note } from '@/types/note';

interface NoteListProps {
  notes: Note[];
  onSetColor: (id: number) => void;
  onShare: (id: number) => void;
  onEdit: (id: number) => void;
  onFavorite: (id: number) => void;
}

export const NoteList: React.FC<NoteListProps> = ({
  notes,
  onSetColor,
  onShare,
  onEdit,
  onFavorite,
}) => {
  return (
    <div className="flex flex-col gap-6 w-full bottom-0">
      {notes.map((note) => (
        <NoteCard
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
};
