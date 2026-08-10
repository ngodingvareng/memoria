import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import React from 'react';

interface ColorPickerDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: () => void;
}

export const ColorPickerDialog: React.FC<ColorPickerDialogProps> = ({
  open,
  onOpenChange,
  onSave,
}) => {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Set color</DialogTitle>
        </DialogHeader>
        <div className="flex gap-2">
          <div className="ring-primary size-4 cursor-pointer rounded-full bg-zinc-200 ring-offset-2 transition-all hover:ring-2" />
          <div className="ring-primary size-4 cursor-pointer rounded-full bg-red-200 ring-offset-2 transition-all hover:ring-2" />
          <div className="ring-primary size-4 cursor-pointer rounded-full bg-green-200 ring-offset-2 transition-all hover:ring-2" />
          <div className="ring-primary size-4 cursor-pointer rounded-full bg-blue-200 ring-offset-2 transition-all hover:ring-2" />
          <div className="ring-primary size-4 cursor-pointer rounded-full bg-yellow-200 ring-offset-2 transition-all hover:ring-2" />
        </div>
        <DialogFooter>
          <DialogClose render={<Button type="button" variant="ghost" />}>
            Cancel
          </DialogClose>
          <Button type="button" onClick={onSave}>
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
