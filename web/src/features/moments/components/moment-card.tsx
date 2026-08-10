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
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
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
import {
  AddMentionButton,
  MentionAutocompletePopover,
  renderTextWithMentions,
  useInlineMentionAutocomplete,
} from '@/features/mentions';
import {
  CommentAuthorsAvatarGroup,
  ReactionPicker,
  ReactionSummary,
} from '@/features/threads';
import {
  getGetMentionsQueryKey,
  useGetMomentsIdMentions,
  usePostMomentsIdMentionsLeave,
} from '@/lib/api/generated/mentions/mentions';
import {
  getGetMomentsQueryKey,
  useGetMomentsIdImages,
} from '@/lib/api/generated/moments/moments';
import { hexToRgba } from '@/lib/colors';
import { queryClient } from '@/lib/query-client';
import { cn } from '@/lib/utils';
import {
  ArrowRight01Icon,
  Comment02Icon,
  Delete02Icon,
  Edit04Icon,
  Logout01Icon,
  MoreVerticalIcon,
  PaintBoardIcon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link, useNavigate, useRouterState } from '@tanstack/react-router';
import { formatDistanceToNow } from 'date-fns';
import React from 'react';
import { ColorSwatchPicker } from './color-swatch-picker';
import { MomentImagesDialog } from './moment-images-dialog';
import { useMomentAudience } from '../lib/use-moment-audience';

interface MomentCardProps {
  id?: string;
  user: {
    name: string;
    username: string;
    imageSrc: string;
    imageAlt: string;
  };
  thread: {
    id?: string;
    name: string;
  };
  colorHex?: string;
  content: string;
  createdAt: Date;
  capturedAt: Date;
  isOwnedByCurrentUser?: boolean;
  // The creator/thread-name row — only shown on Home, a Circle
  // overview, and a Circle-owned thread's own page (callers decide,
  // per page context — this component doesn't infer it).
  showHeader?: boolean;
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
  createdAt,
  capturedAt,
  isOwnedByCurrentUser = false,
  showHeader = false,
  onEditNote,
  onEditColor,
  onDelete,
}: MomentCardProps) {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });
  const navigate = useNavigate();
  const [isImagesDialogOpen, setIsImagesDialogOpen] = React.useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = React.useState(false);
  const [isEditingNote, setIsEditingNote] = React.useState(false);
  const [noteDraft, setNoteDraft] = React.useState(content);
  const [isSaving, setIsSaving] = React.useState(false);
  const [isLeavingMention, setIsLeavingMention] = React.useState(false);
  const [actionError, setActionError] = React.useState<string | null>(null);
  const noteTextareaRef = React.useRef<HTMLTextAreaElement>(null);
  const mentionAutocomplete = useInlineMentionAutocomplete({
    onChange: setNoteDraft,
    textareaRef: noteTextareaRef,
  });
  const imagesQuery = useGetMomentsIdImages(id ?? '', {
    query: { enabled: !!id },
  });
  const images = (imagesQuery.data ?? [])
    .map((image) => image.url)
    .filter((url): url is string => !!url);
  const coverImage = images[images.length - 1];
  // Fades the cover image in once it's actually decoded instead of
  // popping in mid-layout — most noticeable scrolling fast through a
  // freshly-loaded page of cards, where several images finish loading
  // at once.
  const [isCoverImageLoaded, setIsCoverImageLoaded] = React.useState(false);
  React.useEffect(() => {
    setIsCoverImageLoaded(false);
  }, [coverImage]);

  const coverButtonRef = React.useRef<HTMLButtonElement>(null);
  const coverLabelRef = React.useRef<HTMLDivElement>(null);
  const [isCoverCompact, setIsCoverCompact] = React.useState(true);

  React.useLayoutEffect(() => {
    const button = coverButtonRef.current;
    const label = coverLabelRef.current;
    if (!button || !label) return;

    const update = () => {
      setIsCoverCompact(button.offsetHeight <= label.offsetHeight + 10);
    };
    update();

    const observer = new ResizeObserver(update);
    observer.observe(button);
    observer.observe(label);
    return () => observer.disconnect();
  }, []);

  // Comment/Reaction only make sense once this Moment actually has an
  // audience: it lives in a Circle-owned thread, or it names a mention
  // (FEATURES.md, Mention + Response Terms). /moments/{id}/audience's
  // mention_allowed is true for the owner unconditionally (they may
  // always act in their own mention context, even an empty one), so it
  // can't be used as an existence check for the owner — the owner path
  // checks the actual mention list instead. For a non-owner,
  // mention_allowed already only turns true when they're genuinely
  // mentioned, so it's used as-is.
  const audience = useMomentAudience(id ?? '', { enabled: !!id });
  const ownMentionsQuery = useGetMomentsIdMentions(id ?? '', {
    query: { enabled: !!id && isOwnedByCurrentUser },
  });
  const hasCircleAudience = audience.contexts.some(
    (context) => context.type === 'circle'
  );
  const hasMentionAudience = isOwnedByCurrentUser
    ? (ownMentionsQuery.data?.mentions?.length ?? 0) > 0
    : audience.contexts.some((context) => context.type === 'mention');
  const hasAudience = hasCircleAudience || hasMentionAudience;
  const canManage = isOwnedByCurrentUser;
  const canLeaveMention =
    !isOwnedByCurrentUser &&
    audience.contexts.some((context) => context.type === 'mention');
  const showMenu = canManage || canLeaveMention;

  const leaveMention = usePostMomentsIdMentionsLeave();

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

  const handleLeaveMention = async () => {
    if (!id) return;
    setIsLeavingMention(true);
    try {
      await leaveMention.mutateAsync({ id });
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: getGetMentionsQueryKey() }),
        queryClient.invalidateQueries({ queryKey: getGetMomentsQueryKey() }),
      ]);
      // Leaving removes this viewer's access — if they were looking at
      // its detail page, that page is no longer reachable for them.
      if (pathname === `/moment/${id}`) {
        navigate({ to: '/mentions' });
      }
    } finally {
      setIsLeavingMention(false);
    }
  };

  return (
    <Item size="xs">
      <ItemContent className="flex flex-col gap-4">
        <div className="flex group relative select-text hover:cursor-default selection:bg-primary selection:text-primary-foreground gap-4">
          {showMenu && (
            <div className="absolute top-0 -right-6 z-20">
              <DropdownMenu>
                <DropdownMenuTrigger
                  render={<Button variant="ghost" size="icon" />}
                >
                  <HugeiconsIcon strokeWidth={2} icon={MoreVerticalIcon} />
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuGroup>
                    {canManage ? (
                      <>
                        <DropdownMenuItem onClick={startEditingNote}>
                          <HugeiconsIcon strokeWidth={2} icon={Edit04Icon} />
                          Edit content
                        </DropdownMenuItem>
                        <DropdownMenuSub>
                          <DropdownMenuSubTrigger>
                            <HugeiconsIcon
                              strokeWidth={2}
                              icon={PaintBoardIcon}
                            />
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
                      </>
                    ) : (
                      <DropdownMenuItem
                        onClick={handleLeaveMention}
                        disabled={isLeavingMention}
                      >
                        <HugeiconsIcon strokeWidth={2} icon={Logout01Icon} />
                        Leave mention
                      </DropdownMenuItem>
                    )}
                  </DropdownMenuGroup>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          )}
          <button
            ref={coverButtonRef}
            type="button"
            disabled={!coverImage}
            onClick={() => setIsImagesDialogOpen(true)}
            style={{
              backgroundColor: colorHex ? hexToRgba(colorHex, 1) : undefined,
            }}
            className={cn(
              'max-w-2xs flex flex-none flex-col rounded-xl w-full overflow-hidden relative isolate text-left outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50',
              !colorHex && 'bg-gray-500',
              coverImage && 'cursor-pointer'
            )}
          >
            <div
              ref={coverLabelRef}
              className="items-end text-right text-white flex flex-col gap-1 flex-none relative z-10 px-4 py-3.5"
            >
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
              <div
                className={cn(
                  'absolute inset-x-0 bottom-0 z-0',
                  isCoverCompact ? 'top-0' : 'h-2/3'
                )}
              >
                <div
                  className="absolute inset-0 z-10"
                  style={
                    isCoverCompact
                      ? {
                          backgroundColor: colorHex
                            ? hexToRgba(colorHex, 0.8)
                            : 'color-mix(in srgb, #6b7280 40%, transparent)',
                        }
                      : {
                          backgroundImage: colorHex
                            ? `linear-gradient(to bottom, ${hexToRgba(colorHex, 1)}, ${hexToRgba(colorHex, 0.4)})`
                            : `linear-gradient(to bottom, #6b7280, color-mix(in srgb, #6b7280 40%, transparent))`,
                        }
                  }
                />
                <img
                  src={coverImage}
                  alt=""
                  loading="lazy"
                  decoding="async"
                  onLoad={() => setIsCoverImageLoaded(true)}
                  className={cn(
                    'absolute inset-0 size-full object-cover transition-opacity duration-300',
                    isCoverImageLoaded ? 'opacity-100' : 'opacity-0'
                  )}
                />
              </div>
            )}
          </button>
          <Item className="grow pt-1 pb-0">
            {showHeader && (
              <ItemHeader className="-ml-1">
                <Link
                  to="/@{$username}"
                  params={{ username: user.username }}
                  className="flex shrink-0 gap-2 items-center"
                >
                  <Avatar>
                    <AvatarImage src={user.imageSrc} alt={user.imageAlt} />
                    <AvatarFallback>
                      {user.name.charAt(0).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <p className="text-base">{user.username}</p>
                  </div>
                </Link>
                <HugeiconsIcon
                  icon={ArrowRight01Icon}
                  className="size-4 shrink-0 font-bold text-muted-foreground"
                />
                <div className="min-w-0 flex-1 font-medium">
                  {thread.id ? (
                    <Link
                      to="/thread/$id"
                      params={{ id: thread.id }}
                      className="block truncate hover:underline"
                    >
                      {thread.name}
                    </Link>
                  ) : (
                    <p className="truncate">{thread.name}</p>
                  )}
                </div>
                <div className="shrink-0 flex gap-1 items-center">
                  <p className="font-medium text-muted-foreground">
                    {formatDistanceToNow(createdAt, {
                      addSuffix: true,
                    })}
                  </p>
                </div>
              </ItemHeader>
            )}
            <ItemContent className="max-w-none gap-2">
              {isEditingNote ? (
                <div className="flex flex-col gap-2">
                  <Textarea
                    ref={noteTextareaRef}
                    value={noteDraft}
                    onChange={mentionAutocomplete.handleChange}
                    onKeyDown={mentionAutocomplete.handleKeyDown}
                    onClick={mentionAutocomplete.handleSelectionChange}
                    onKeyUp={mentionAutocomplete.handleSelectionChange}
                    maxLength={10000}
                    autoFocus
                    className="text-base min-h-24"
                    disabled={isSaving}
                  />
                  <MentionAutocompletePopover
                    anchorRef={noteTextareaRef}
                    open={mentionAutocomplete.isOpen}
                    onOpenChange={(open) =>
                      !open && mentionAutocomplete.closeSuggestions()
                    }
                    users={mentionAutocomplete.suggestions}
                    activeIndex={mentionAutocomplete.activeIndex}
                    isLoading={mentionAutocomplete.isLoading}
                    onSelect={mentionAutocomplete.handleSelect}
                  />
                  <div className="flex gap-2 justify-between">
                    <AddMentionButton
                      onInsert={mentionAutocomplete.insertAtCursor}
                    />
                    <div className="flex gap-2">
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
                </div>
              ) : (
                <p className="whitespace-pre-wrap text-base/7">
                  {renderTextWithMentions(content)}
                </p>
              )}
              {actionError && (
                <Alert variant="destructive">
                  <AlertDescription>{actionError}</AlertDescription>
                </Alert>
              )}
            </ItemContent>
            {id && hasAudience && (
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
                <CommentAuthorsAvatarGroup momentId={id} />
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
