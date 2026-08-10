import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Field, FieldLabel } from '@/components/ui/field';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import React from 'react';

interface DateTimeDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: () => void;
}

export const DateTimeDialog: React.FC<DateTimeDialogProps> = ({
  open,
  onOpenChange,
  onSave,
}) => {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Set Datetime</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col gap-6">
          <div className="flex gap-4">
            <Field>
              <FieldLabel>Date</FieldLabel>
              <ButtonGroup>
                <Select defaultValue="09">
                  <Tooltip>
                    <TooltipTrigger render={<SelectTrigger id="day" />}>
                      <SelectValue />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Day</p>
                    </TooltipContent>
                  </Tooltip>
                  <SelectContent>
                    {Array.from({ length: 31 }, (_, i) => (
                      <SelectItem
                        key={i + 1}
                        value={(i + 1).toString().padStart(2, '0')}
                      >
                        {(i + 1).toString().padStart(2, '0')}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Select defaultValue="09">
                  <Tooltip>
                    <TooltipTrigger render={<SelectTrigger id="month" />}>
                      <SelectValue />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Month</p>
                    </TooltipContent>
                  </Tooltip>
                  <SelectContent>
                    {Array.from({ length: 12 }, (_, i) => (
                      <SelectItem
                        key={i + 1}
                        value={(i + 1).toString().padStart(2, '0')}
                      >
                        {(i + 1).toString().padStart(2, '0')}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Select defaultValue="2029">
                  <Tooltip>
                    <TooltipTrigger render={<SelectTrigger id="year" />}>
                      <SelectValue />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Year</p>
                    </TooltipContent>
                  </Tooltip>
                  <SelectContent>
                    <SelectItem value="2024">2024</SelectItem>
                    <SelectItem value="2025">2025</SelectItem>
                    <SelectItem value="2026">2026</SelectItem>
                    <SelectItem value="2027">2027</SelectItem>
                    <SelectItem value="2028">2028</SelectItem>
                    <SelectItem value="2029">2029</SelectItem>
                  </SelectContent>
                </Select>
              </ButtonGroup>
            </Field>
            <Field>
              <FieldLabel>Time</FieldLabel>
              <ButtonGroup>
                <Select defaultValue="23">
                  <Tooltip>
                    <TooltipTrigger render={<SelectTrigger id="hour" />}>
                      <SelectValue />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Hour</p>
                    </TooltipContent>
                  </Tooltip>
                  <SelectContent>
                    {Array.from({ length: 24 }, (_, i) => (
                      <SelectItem key={i} value={i.toString().padStart(2, '0')}>
                        {i.toString().padStart(2, '0')}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Select defaultValue="00">
                  <Tooltip>
                    <TooltipTrigger render={<SelectTrigger id="minute" />}>
                      <SelectValue />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Minute</p>
                    </TooltipContent>
                  </Tooltip>
                  <SelectContent>
                    {Array.from({ length: 60 }, (_, i) => (
                      <SelectItem key={i} value={i.toString().padStart(2, '0')}>
                        {i.toString().padStart(2, '0')}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </ButtonGroup>
            </Field>
          </div>
          <div className="flex flex-wrap gap-1">
            <Button variant="outline" size="xs">
              Now
            </Button>
            <Button variant="outline" size="xs">
              Yesterday
            </Button>
            <Button variant="outline" size="xs">
              2 days before
            </Button>
          </div>
        </div>
        <DialogFooter>
          <DialogClose render={<Button type="button" variant="ghost" />}>
            Cancel
          </DialogClose>
          <Button type="button" onClick={onSave}>
            Set
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
