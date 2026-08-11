import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import { Edit04Icon, EyeIcon, Setting06Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';
import { ThreadCircleBadge } from './thread-circle-badge';

interface ThreadHeaderProps {
  threadId: string;
  threadName: string;
  circleId?: string;
  isReadMode: boolean;
  onToggleMode: () => void;
}

export function ThreadHeader({
  threadId,
  threadName,
  circleId,
  isReadMode,
  onToggleMode,
}: ThreadHeaderProps) {
  return (
    <div className="mt-4 flex flex-col gap-2">
      <div className="flex items-start justify-end gap-1">
        <div className="flex grow flex-col gap-1">
          <h1 className="font-heading text-2xl font-semibold">{threadName}</h1>
          {circleId && <ThreadCircleBadge circleId={circleId} />}
        </div>

        <ButtonGroup>
          <Button variant="outline" onClick={onToggleMode}>
            {isReadMode ? (
              <>
                <HugeiconsIcon icon={EyeIcon} /> Read mode
              </>
            ) : (
              <>
                <HugeiconsIcon icon={Edit04Icon} /> Edit mode
              </>
            )}
          </Button>
          <Button
            variant="outline"
            render={<Link to="/thread/$id/manage" params={{ id: threadId }} />}
          >
            <HugeiconsIcon icon={Setting06Icon} /> Manage
          </Button>
        </ButtonGroup>
      </div>
    </div>
  );
}
