import { Button } from '@/components/ui/button';
import { Item, ItemContent } from '@/components/ui/item';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { getNoteColorClass } from '@/lib/colors';
import { cn } from '@/lib/utils';
import type { Note } from '@/types/note';
import {
  Edit04Icon,
  PaintBoardIcon,
  Share01Icon,
  StarIcon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';

interface CaptureCardProps {
  note: Note;
  onSetColor: (id: number) => void;
  onShare: (id: number) => void;
  onEdit: (id: number) => void;
  onFavorite: (id: number) => void;
}

export function CaptureCard({
  note,
  onSetColor,
  onShare,
  onEdit,
  onFavorite,
}: CaptureCardProps) {
  return (
    <div className="flex group select-text hover:cursor-default selection:bg-primary selection:text-primary-foreground gap-4">
      <Item
        className={cn(
          'max-w-2xs flex flex-col w-full overflow-hidden',
          note.color ? getNoteColorClass(note.color) : getNoteColorClass('zinc')
        )}
      >
        <ItemContent className="flex flex-col items-end w-full relative z-10">
          <div className="items-end text-right flex flex-col gap-1 grow">
            <div className="flex flex-col">
              <p className="font-bold text-lg">{note.date}</p>
              <p className="font-medium text-xl">{note.time}</p>
            </div>
          </div>
          <div className="absolute -bottom-1 left-0 group-hover:opacity-100 opacity-0 transition-opacity">
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button
                    variant="ghost"
                    size="icon-lg"
                    onClick={() => onFavorite(note.id)}
                  />
                }
              >
                <HugeiconsIcon strokeWidth={2} icon={StarIcon} />
              </TooltipTrigger>
              <TooltipContent>
                <p>Add to favorites</p>
              </TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => onSetColor(note.id)}
                  />
                }
              >
                <HugeiconsIcon strokeWidth={2} icon={PaintBoardIcon} />
              </TooltipTrigger>
              <TooltipContent>
                <p>Set color</p>
              </TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => onEdit(note.id)}
                  />
                }
              >
                <HugeiconsIcon strokeWidth={2} icon={Edit04Icon} />
              </TooltipTrigger>
              <TooltipContent>
                <p>Edit content</p>
              </TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => onShare(note.id)}
                  />
                }
              >
                <HugeiconsIcon strokeWidth={2} icon={Share01Icon} />
              </TooltipTrigger>
              <TooltipContent>
                <p>Share</p>
              </TooltipContent>
            </Tooltip>
          </div>
        </ItemContent>
      </Item>
      <Item className="grow hover:bg-muted/50">
        <ItemContent className="typeset max-w-none">
          {note.content}
          <div className="flex flex-wrap gap-x-2">
            {note.tags
              .map((tag) => `#${tag}`)
              .map((tag) => (
                <p>
                  <a key={tag} className="hover:underline font-medium">
                    {tag}
                  </a>
                </p>
              ))}
          </div>
        </ItemContent>
      </Item>
    </div>
  );
}
