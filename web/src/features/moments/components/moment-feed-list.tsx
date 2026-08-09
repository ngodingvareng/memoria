import { ItemGroup, ItemSeparator } from '@/components/ui/item';
import { MomentFeedItem } from './moment-feed-item';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMomentResponse } from '@/lib/api/generated/models';
import { Fragment } from 'react';

interface MomentFeedListProps {
  moments: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMomentResponse[];
}

export function MomentFeedList({ moments }: MomentFeedListProps) {
  return (
    <ItemGroup>
      {moments.map((moment, index) => (
        <Fragment key={moment.id}>
          <MomentFeedItem moment={moment} />
          {index !== moments.length - 1 && <ItemSeparator />}
        </Fragment>
      ))}
    </ItemGroup>
  );
}
