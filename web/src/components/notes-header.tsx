'use client';

import { Button } from '@/components/ui/button';
import {
  CarouselHorizontalIcon,
  Edit04Icon,
  EyeIcon,
  MoreHorizontalSquare01Icon,
  Share01Icon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import React from 'react';

interface NotesHeaderProps {
  isReadMode: boolean;
  onToggleMode: () => void;
  onShare: () => void;
  noteCount: number;
}

export const NotesHeader: React.FC<NotesHeaderProps> = ({
  isReadMode,
  onToggleMode,
  onShare,
  noteCount,
}) => {
  return (
    <div className="flex flex-col gap-2 mt-4">
      <div className="flex items-center justify-end gap-1">
        <Button variant="ghost" disabled>
          <HugeiconsIcon icon={CarouselHorizontalIcon} /> Slideshow (Coming
          soon)
        </Button>
        <Button variant="ghost" onClick={onToggleMode}>
          {isReadMode ? (
            <>
              <HugeiconsIcon icon={EyeIcon} /> Read mode
            </>
          ) : (
            <>
              <HugeiconsIcon icon={Edit04Icon} /> Edit mode
            </>
          )}
        </Button>
        <Button variant="default" onClick={onShare}>
          Share <HugeiconsIcon icon={Share01Icon} />
        </Button>
        <Button variant="secondary" size="icon">
          <HugeiconsIcon icon={MoreHorizontalSquare01Icon} />
        </Button>
      </div>
      <div className="grow">
        <h1 className="text-3xl font-semibold font-heading">
          Adalah Pokoknya ({noteCount} items)
        </h1>
      </div>
    </div>
  );
};
