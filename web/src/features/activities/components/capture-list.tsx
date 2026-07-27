import { ItemGroup, ItemSeparator } from '@/components/ui/item';
import { CaptureCard, type CaptureCardParam } from './capture-card';

interface CaptureListProps {
  captures: CaptureCardParam[];
  withStoryLayout?: boolean;
}

export function CaptureList({
  captures,
  withStoryLayout = false,
}: CaptureListProps) {
  return (
    <ItemGroup>
      {captures.map((capture, index) => (
        <>
          <CaptureCard
            user={capture.user}
            activity={capture.activity}
            stats={capture.stats}
            color={capture.color}
            content={capture.content}
            tags={capture.tags}
            createdAt={capture.createdAt}
            capturedAt={capture.capturedAt}
            isPublished={withStoryLayout || capture.isPublished}
            isOwnedByCurrentUser={
              !withStoryLayout || capture.isOwnedByCurrentUser
            }
          />
          {withStoryLayout && captures.length - 1 != index && <ItemSeparator />}
        </>
      ))}
    </ItemGroup>
  );
}
