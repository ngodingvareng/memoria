import { ItemGroup, ItemSeparator } from '@/components/ui/item';
import { MomentCard, type MomentCardParam } from './moment-card';

interface MomentListProps {
  moments: MomentCardParam[];
  withStoryLayout?: boolean;
}

export function MomentList({
  moments,
  withStoryLayout = false,
}: MomentListProps) {
  return (
    <ItemGroup>
      {moments.map((moment, index) => (
        <>
          <MomentCard
            user={moment.user}
            thread={moment.thread}
            stats={moment.stats}
            color={moment.color}
            content={moment.content}
            tags={moment.tags}
            createdAt={moment.createdAt}
            capturedAt={moment.capturedAt}
            isPublished={withStoryLayout || moment.isPublished}
            isOwnedByCurrentUser={
              !withStoryLayout || moment.isOwnedByCurrentUser
            }
          />
          {withStoryLayout && moments.length - 1 != index && <ItemSeparator />}
        </>
      ))}
    </ItemGroup>
  );
}
