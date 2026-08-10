import { ItemGroup, ItemSeparator } from '@/components/ui/item';
import { MomentFeedItem } from './moment-feed-item';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMomentResponse } from '@/lib/api/generated/models';
import { Fragment } from 'react';

interface MomentFeedListProps {
  moments: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMomentResponse[];
  showHeader?: boolean;
  onEditNote?: (momentId: string, note: string) => void | Promise<void>;
  onEditColor?: (momentId: string, colorHex: string) => void | Promise<void>;
  onDelete?: (momentId: string) => void | Promise<void>;
}

export function MomentFeedList({
  moments,
  showHeader = false,
  onEditNote,
  onEditColor,
  onDelete,
}: MomentFeedListProps) {
  return (
    <ItemGroup>
      {moments.map((moment, index) => (
        <Fragment key={moment.id}>
          <MomentFeedItem
            moment={moment}
            showHeader={showHeader}
            onEditNote={onEditNote}
            onEditColor={onEditColor}
            onDelete={onDelete}
          />
          {index !== moments.length - 1 && <ItemSeparator />}
        </Fragment>
      ))}
    </ItemGroup>
  );
}
