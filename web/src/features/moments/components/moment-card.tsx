import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuPortal,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Item,
  ItemContent,
  ItemFooter,
  ItemHeader,
} from '@/components/ui/item';
import { Textarea } from '@/components/ui/textarea';
import { renderTextWithMentions } from '@/features/mentions';
import { ReactionPicker, ReactionSummary } from '@/features/threads';
import { hexToRgba } from '@/lib/colors';
import { cn } from '@/lib/utils';
import {
  ArrowRight01Icon,
  Comment02Icon,
  Delete02Icon,
  Edit04Icon,
  Globe02Icon,
  MoreVerticalIcon,
  PaintBoardIcon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';
import { formatDistanceToNow } from 'date-fns';
import React from 'react';
import { ColorSwatchPicker } from './color-swatch-picker';
import { MomentImagesDialog } from './moment-images-dialog';

export interface MomentCardParam {
  id?: string;
  user: {
    name: string;
    username: string;
    imageSrc: string;
    imageAlt: string;
  };
  thread: {
    name: string;
  };
  colorHex?: string;
  content: string;
  images?: string[];
  createdAt: Date;
  capturedAt: Date;
  isPublished?: boolean;
  isOwnedByCurrentUser?: boolean;
}

interface MomentCardProps extends MomentCardParam {
  isEditMode?: boolean;
  onEditNote?: (note: string) => void | Promise<void>;
  onEditColor?: (colorHex: string) => void | Promise<void>;
  onDelete?: () => void | Promise<void>;
}

export function MomentCard({
  id,
  user,
  thread,
  colorHex,
  content,
  images = [],
  createdAt,
  capturedAt,
  isPublished = false,
  isOwnedByCurrentUser = false,
  isEditMode = false,
  onEditNote,
  onEditColor,
  onDelete,
}: MomentCardProps) {
  const [isImagesDialogOpen, setIsImagesDialogOpen] = React.useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = React.useState(false);
  const [isEditingNote, setIsEditingNote] = React.useState(false);
  const [noteDraft, setNoteDraft] = React.useState(content);
  const [isSaving, setIsSaving] = React.useState(false);
  const [actionError, setActionError] = React.useState<string | null>(null);
  const coverImage = images[images.length - 1];

  const canManage = isOwnedByCurrentUser && isEditMode;

  const startEditingNote = () => {
    setActionError(null);
    setNoteDraft(content);
    setIsEditingNote(true);
  };

  const handleSaveNote = async () => {
    setActionError(null);
    setIsSaving(true);
    try {
      await onEditNote?.(noteDraft);
      setIsEditingNote(false);
    } catch {
      setActionError('Failed to save. Please try again.');
    } finally {
      setIsSaving(false);
    }
  };

  const handleColorChange = async (hex: string) => {
    setActionError(null);
    try {
      await onEditColor?.(hex);
    } catch {
      setActionError('Failed to update color. Please try again.');
    }
  };

  const handleDelete = async () => {
    setActionError(null);
    setIsSaving(true);
    try {
      await onDelete?.();
      setIsDeleteDialogOpen(false);
    } catch {
      setActionError('Failed to delete this moment. Please try again.');
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <Item size="xs">
      <ItemContent className="flex flex-col gap-4">
        <div className="flex group relative select-text hover:cursor-default selection:bg-primary selection:text-primary-foreground gap-4">
          {canManage && (
            <div className="absolute -top-1 -right-6 z-20">
              <DropdownMenu>
                <DropdownMenuTrigger
                  render={<Button variant="ghost" size="icon" />}
                >
                  <HugeiconsIcon strokeWidth={2} icon={MoreVerticalIcon} />
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuGroup>
                    <DropdownMenuItem onClick={startEditingNote}>
                      <HugeiconsIcon strokeWidth={2} icon={Edit04Icon} />
                      Edit content
                    </DropdownMenuItem>
                    <DropdownMenuSub>
                      <DropdownMenuSubTrigger>
                        <HugeiconsIcon strokeWidth={2} icon={PaintBoardIcon} />
                        Set color
                      </DropdownMenuSubTrigger>
                      <DropdownMenuPortal>
                        <DropdownMenuSubContent>
                          <ColorSwatchPicker
                            value={colorHex ?? ''}
                            onChange={handleColorChange}
                            className="p-2"
                          />
                        </DropdownMenuSubContent>
                      </DropdownMenuPortal>
                    </DropdownMenuSub>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      variant="destructive"
                      onClick={() => setIsDeleteDialogOpen(true)}
                    >
                      <HugeiconsIcon strokeWidth={2} icon={Delete02Icon} />
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuGroup>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          )}
          <button
            type="button"
            disabled={!coverImage}
            onClick={() => setIsImagesDialogOpen(true)}
            style={{
              backgroundColor: colorHex ? hexToRgba(colorHex, 0.2) : undefined,
            }}
            className={cn(
              'max-w-2xs flex flex-none flex-col rounded-xl w-full overflow-hidden relative isolate text-left outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50',
              !colorHex && 'bg-muted',
              coverImage && 'cursor-pointer'
            )}
          >
            <div className="items-end text-right flex flex-col gap-1 grow relative z-10 px-4 py-3.5">
              <div className="flex flex-col">
                <p className="font-bold text-lg">
                  {capturedAt.toLocaleDateString('id-ID', {
                    day: 'numeric',
                    month: 'long',
                    year: 'numeric',
                  })}
                </p>
                <p className="font-medium text-xl/4">
                  {capturedAt.toLocaleTimeString('id-ID', {
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </p>
              </div>
            </div>
            {coverImage && (
              <div className="absolute inset-x-0 bottom-0 h-2/3 z-0">
                <div
                  className="absolute inset-x-0 top-0 h-1/3 z-10"
                  style={{
                    backgroundImage: `linear-gradient(to bottom, ${colorHex ?? 'var(--muted)'}, transparent)`,
                  }}
                />
                <img
                  src={coverImage}
                  alt=""
                  className="size-full object-cover"
                />
              </div>
            )}
          </button>
          <Item className="grow pt-1 pb-0">
            {isPublished && (
              <ItemHeader>
                <Link
                  to="/@{$username}"
                  params={{ username: user.username }}
                  className="flex gap-2 items-center"
                >
                  <Avatar>
                    <AvatarImage src={user.imageSrc} alt={user.imageAlt} />
                    <AvatarFallback>CN</AvatarFallback>
                  </Avatar>
                  <div>
                    <p className="text-base">{user.username}</p>
                  </div>
                </Link>
                <HugeiconsIcon
                  icon={ArrowRight01Icon}
                  className="size-4 font-bold text-muted-foreground"
                />
                <div className="font-medium">
                  <p>{thread.name}</p>
                </div>
                <div className="grow items-center flex gap-2 justify-end">
                  <div>
                    <Badge className="bg-sky-50 text-sky-700 dark:bg-sky-950 dark:text-sky-300">
                      <HugeiconsIcon
                        data-icon="inline-start"
                        icon={Globe02Icon}
                      />
                      Public
                    </Badge>
                  </div>
                  <div className="flex gap-1 items-center">
                    <p className="font-medium text-muted-foreground">
                      {formatDistanceToNow(createdAt, {
                        addSuffix: true,
                      })}
                    </p>
                  </div>
                  <DropdownMenu>
                    <DropdownMenuTrigger
                      render={
                        <Button size="icon" variant="ghost" className="-mr-3">
                          <HugeiconsIcon
                            strokeWidth={2}
                            icon={MoreVerticalIcon}
                          />
                        </Button>
                      }
                    />
                    <DropdownMenuContent className="w-40" align="start">
                      <DropdownMenuGroup>
                        <DropdownMenuItem>Profile</DropdownMenuItem>
                        <DropdownMenuItem>Billing</DropdownMenuItem>
                        <DropdownMenuItem>Settings</DropdownMenuItem>
                      </DropdownMenuGroup>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </ItemHeader>
            )}
            <ItemContent className="max-w-none gap-2">
              {isEditingNote ? (
                <div className="flex flex-col gap-2">
                  <Textarea
                    value={noteDraft}
                    onChange={(e) => setNoteDraft(e.target.value)}
                    maxLength={10000}
                    autoFocus
                    className="text-base min-h-24"
                    disabled={isSaving}
                  />
                  <div className="flex gap-2 justify-end">
                    <Button
                      size="sm"
                      variant="ghost"
                      disabled={isSaving}
                      onClick={() => setIsEditingNote(false)}
                    >
                      Cancel
                    </Button>
                    <Button
                      size="sm"
                      disabled={isSaving}
                      onClick={handleSaveNote}
                    >
                      Save
                    </Button>
                  </div>
                </div>
              ) : (
                <p className="whitespace-pre-wrap text-base">
                  {renderTextWithMentions(content)}
                </p>
              )}
              {actionError && (
                <Alert variant="destructive">
                  <AlertDescription>{actionError}</AlertDescription>
                </Alert>
              )}
            </ItemContent>
            {isPublished && id && (
              <ItemFooter className="flex gap-2 -ml-3 justify-start items-center">
                <ReactionPicker momentId={id} />
                <ReactionSummary momentId={id} />
                <Button
                  size="sm"
                  variant="ghost"
                  className="[&_svg]:size-5!"
                  render={<Link to="/moment/$id" params={{ id }} />}
                >
                  <HugeiconsIcon strokeWidth={2} icon={Comment02Icon} />
                </Button>
              </ItemFooter>
            )}
          </Item>
        </div>
      </ItemContent>
      {images.length > 0 && (
        <MomentImagesDialog
          open={isImagesDialogOpen}
          onOpenChange={setIsImagesDialogOpen}
          images={images}
        />
      )}
      <AlertDialog
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete this moment?</AlertDialogTitle>
            <AlertDialogDescription>
              This can&apos;t be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          {actionError && (
            <Alert variant="destructive">
              <AlertDescription>{actionError}</AlertDescription>
            </Alert>
          )}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isSaving}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={handleDelete}
              disabled={isSaving}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Item>
  );
}
