import { ConfirmDestructiveDialog } from '@/components/dialogs/confirm-destructive-dialog';
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
import { renderTextWithMentions } from '@/features/mentions';
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
import { queryClient } from '@/lib/query-client';
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
import { MomentCardCover } from './moment-card-cover';
import { MomentImagesDialog } from './moment-images-dialog';
import { MomentNoteEditor } from './moment-note-editor';
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
  const [isLeavingMention, setIsLeavingMention] = React.useState(false);
  const [colorChangeError, setColorChangeError] = React.useState<string | null>(
    null
  );
  const imagesQuery = useGetMomentsIdImages(id ?? '', {
    query: { enabled: !!id },
  });
  const images = (imagesQuery.data ?? [])
    .map((image) => image.url)
    .filter((url): url is string => !!url);
  const coverImage = images[images.length - 1];

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

  const handleColorChange = async (hex: string) => {
    setColorChangeError(null);
    try {
      await onEditColor?.(hex);
    } catch {
      setColorChangeError('Failed to update color. Please try again.');
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
        <div className="group selection:bg-primary selection:text-primary-foreground relative flex gap-4 select-text hover:cursor-default">
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
                        <DropdownMenuItem
                          onClick={() => setIsEditingNote(true)}
                        >
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
          <MomentCardCover
            colorHex={colorHex}
            capturedAt={capturedAt}
            coverImage={coverImage}
            onOpen={() => setIsImagesDialogOpen(true)}
          />
          <Item className="grow pt-1 pb-0">
            {showHeader && (
              <ItemHeader className="-ml-1">
                <Link
                  to="/@{$username}"
                  params={{ username: user.username }}
                  className="flex shrink-0 items-center gap-2"
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
                  className="text-muted-foreground size-4 shrink-0 font-bold"
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
                <div className="flex shrink-0 items-center gap-1">
                  <p className="text-muted-foreground font-medium">
                    {formatDistanceToNow(createdAt, {
                      addSuffix: true,
                    })}
                  </p>
                </div>
              </ItemHeader>
            )}
            <ItemContent className="max-w-none gap-2">
              {isEditingNote ? (
                <MomentNoteEditor
                  content={content}
                  onCancel={() => setIsEditingNote(false)}
                  onSave={async (note) => {
                    await onEditNote?.(note);
                    setIsEditingNote(false);
                  }}
                />
              ) : (
                <p className="text-base/7 whitespace-pre-wrap">
                  {renderTextWithMentions(content)}
                </p>
              )}
              {colorChangeError && (
                <Alert variant="destructive">
                  <AlertDescription>{colorChangeError}</AlertDescription>
                </Alert>
              )}
            </ItemContent>
            {id && hasAudience && (
              <ItemFooter className="-ml-3 flex items-center justify-start gap-2">
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
      <ConfirmDestructiveDialog
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        title="Delete this moment?"
        description="This can't be undone."
        confirmLabel="Delete"
        errorFallback="Failed to delete this moment. Please try again."
        onConfirm={async () => {
          await onDelete?.();
        }}
      />
    </Item>
  );
}
