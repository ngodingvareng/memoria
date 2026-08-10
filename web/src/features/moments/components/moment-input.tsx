import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { InputGroupTextarea } from '@/components/ui/input-group';
import { Item, ItemContent } from '@/components/ui/item';
import {
  AddMentionButton,
  MentionAutocompletePopover,
  useInlineMentionAutocomplete,
} from '@/features/mentions';
import { ArrowUp02Icon, PlusSignIcon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { useRef, useState } from 'react';
import { ColorSwatchPicker } from './color-swatch-picker';
import { ImagePreviewList } from './image-preview-list';

export interface MomentDraft {
  note: string;
  colorHex: string;
  images: File[];
}

interface MomentInputProps {
  onOpenTimeDialog: () => void;
  onCapture: (draft: MomentDraft) => void | Promise<void>;
}

export function MomentInput({ onOpenTimeDialog, onCapture }: MomentInputProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [note, setNote] = useState('');
  const [colorHex, setColorHex] = useState('');
  const [images, setImages] = useState<File[]>([]);
  const [isCapturing, setIsCapturing] = useState(false);
  const mentionAutocomplete = useInlineMentionAutocomplete({
    onChange: setNote,
    textareaRef,
  });

  const handleCapture = async () => {
    if (!note.trim() && images.length === 0) return;
    setIsCapturing(true);
    try {
      await onCapture({ note, colorHex, images });
      setNote('');
      setColorHex('');
      setImages([]);
    } finally {
      setIsCapturing(false);
    }
  };

  return (
    <div className="sticky bottom-0 pb-6 z-30 bg-linear-to-t from-background pt-20 from-60% to-transparent w-full">
      <Item
        variant="outline"
        className="mx-auto shadow-sm bg-card max-w-5xl rounded-4xl"
      >
        <ItemContent className="flex flex-col min-h-20 max-h-[calc(100vh-10rem)]">
          <div className="grow overflow-y-auto flex flex-col gap-2">
            <InputGroupTextarea
              ref={textareaRef}
              value={note}
              onChange={mentionAutocomplete.handleChange}
              onKeyDown={mentionAutocomplete.handleKeyDown}
              onClick={mentionAutocomplete.handleSelectionChange}
              onKeyUp={mentionAutocomplete.handleSelectionChange}
              placeholder="Woylah cikk, ketik sini..."
              maxLength={10000}
              className="text-foreground! text-base! min-h-16"
            />
            <MentionAutocompletePopover
              anchorRef={textareaRef}
              open={mentionAutocomplete.isOpen}
              onOpenChange={(open) =>
                !open && mentionAutocomplete.closeSuggestions()
              }
              users={mentionAutocomplete.suggestions}
              activeIndex={mentionAutocomplete.activeIndex}
              isLoading={mentionAutocomplete.isLoading}
              onSelect={mentionAutocomplete.handleSelect}
            />
            <ImagePreviewList
              images={images}
              onRemove={(index) =>
                setImages((prev) => prev.filter((_, i) => i !== index))
              }
            />
          </div>
          <div className="flex justify-between gap-1">
            <div className="flex gap-1">
              <Button
                variant="secondary"
                size="sm"
                onClick={() => fileInputRef.current?.click()}
              >
                <HugeiconsIcon strokeWidth={2.5} icon={PlusSignIcon} />
                Add photos
              </Button>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                multiple
                className="hidden"
                onChange={(e) => {
                  const files = Array.from(e.target.files ?? []);
                  e.target.value = '';
                  if (files.length > 0) {
                    setImages((prev) => [...prev, ...files]);
                  }
                }}
              />

              <DropdownMenu>
                <DropdownMenuTrigger
                  render={<Button variant="secondary" size="sm" />}
                >
                  <span
                    className="size-4 rounded-full border border-foreground"
                    style={{ backgroundColor: colorHex || undefined }}
                  />
                  <span>Color</span>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <ColorSwatchPicker
                    value={colorHex}
                    onChange={setColorHex}
                    className="p-2"
                  />
                </DropdownMenuContent>
              </DropdownMenu>

              <AddMentionButton onInsert={mentionAutocomplete.insertAtCursor} />
            </div>

            <div className="flex items-center gap-2">
              <ButtonGroup>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={onOpenTimeDialog}
                >
                  10/02/2033 10:00
                </Button>
                <Button
                  size="icon-sm"
                  onClick={handleCapture}
                  disabled={isCapturing}
                >
                  <HugeiconsIcon strokeWidth={2.5} icon={ArrowUp02Icon} />
                </Button>
              </ButtonGroup>
            </div>
          </div>
        </ItemContent>
      </Item>
    </div>
  );
}
