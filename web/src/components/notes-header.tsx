'use client';

import { Button } from '@/components/ui/button';
import { Edit04Icon, EyeIcon, Search01Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import React from 'react';
import { ButtonGroup } from './ui/button-group';
import { Link } from '@tanstack/react-router';

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
        <ButtonGroup>
          <Button variant="outline">
            <HugeiconsIcon icon={Search01Icon} />
          </Button>
          <Button variant="outline" onClick={onToggleMode}>
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
          <Button
            variant="outline"
            render={
              <Link to="/activity/$id/info" params={{ id: '1' }}>
                Info
              </Link>
            }
          />
        </ButtonGroup>
      </div>
      <div className="grow">
        <h1 className="text-3xl font-semibold font-heading">
          Adalah Pokoknya ({noteCount} items)
        </h1>
      </div>
    </div>
  );
};
