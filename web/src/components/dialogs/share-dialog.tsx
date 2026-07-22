'use client';

import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Facebook02Icon,
  Link04Icon,
  NewTwitterIcon,
  WhatsappIcon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import React from 'react';

interface ShareDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export const ShareDialog: React.FC<ShareDialogProps> = ({
  open,
  onOpenChange,
}) => {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Share to</DialogTitle>
        </DialogHeader>
        <div className="flex gap-2">
          <Button variant="secondary" size="icon-lg">
            <HugeiconsIcon strokeWidth={2} icon={WhatsappIcon} />
          </Button>
          <Button variant="secondary" size="icon-lg">
            <HugeiconsIcon strokeWidth={2} icon={Link04Icon} />
          </Button>
          <Button variant="secondary" size="icon-lg">
            <HugeiconsIcon strokeWidth={2} icon={Facebook02Icon} />
          </Button>
          <Button variant="secondary" size="icon-lg">
            <HugeiconsIcon strokeWidth={2} icon={NewTwitterIcon} />
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
};
