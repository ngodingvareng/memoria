import { ItemGroup, ItemSeparator } from '@/components/ui/item';
import { MomentCard, type MomentCardParam } from './moment-card';

interface MomentListProps {
  moments: MomentCardParam[];
  withStoryLayout?: boolean;
  isEditMode?: boolean;
  onEditNote?: (id: string, note: string) => void | Promise<void>;
  onEditColor?: (id: string, colorHex: string) => void | Promise<void>;
  onDelete?: (id: string) => void | Promise<void>;
}

export function MomentList({
  moments,
  withStoryLayout = false,
  isEditMode = false,
  onEditNote,
  onEditColor,
  onDelete,
}: MomentListProps) {
  return (
    <ItemGroup>
      {moments.map((moment, index) => (
        <>
          <MomentCard
            id={moment.id}
            user={moment.user}
            thread={moment.thread}
            colorHex={moment.colorHex}
            content={moment.content}
            images={moment.images}
            createdAt={moment.createdAt}
            capturedAt={moment.capturedAt}
            isPublished={withStoryLayout || moment.isPublished}
            isOwnedByCurrentUser={
              !withStoryLayout || moment.isOwnedByCurrentUser
            }
            isEditMode={isEditMode}
            onEditNote={
              moment.id && onEditNote
                ? (note) => onEditNote(moment.id!, note)
                : undefined
            }
            onEditColor={
              moment.id && onEditColor
                ? (colorHex) => onEditColor(moment.id!, colorHex)
                : undefined
            }
            onDelete={
              moment.id && onDelete ? () => onDelete(moment.id!) : undefined
            }
          />
          {withStoryLayout && moments.length - 1 != index && <ItemSeparator />}
        </>
      ))}
    </ItemGroup>
  );
}
