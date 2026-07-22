"use client";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import React from "react";

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
          <div className="bg-zinc-200 size-4 rounded-full cursor-pointer hover:ring-2 ring-primary ring-offset-2 transition-all" />
          <div className="bg-red-200 size-4 rounded-full cursor-pointer hover:ring-2 ring-primary ring-offset-2 transition-all" />
          <div className="bg-green-200 size-4 rounded-full cursor-pointer hover:ring-2 ring-primary ring-offset-2 transition-all" />
          <div className="bg-blue-200 size-4 rounded-full cursor-pointer hover:ring-2 ring-primary ring-offset-2 transition-all" />
          <div className="bg-yellow-200 size-4 rounded-full cursor-pointer hover:ring-2 ring-primary ring-offset-2 transition-all" />
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
