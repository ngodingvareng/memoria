import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import {
  Edit04Icon,
  EyeIcon,
  Search01Icon,
  Setting06Icon,
  Share01Icon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';

interface ThreadHeaderProps {
  threadId: string;
  threadName: string;
  isReadMode: boolean;
  onToggleMode: () => void;
  onShare: () => void;
}

export function ThreadHeader({
  threadId,
  threadName,
  isReadMode,
  onToggleMode,
  onShare,
}: ThreadHeaderProps) {
  return (
    <div className="flex flex-col gap-2 mt-4">
      <div className="flex items-center justify-end gap-1">
        <div className="grow">
          <h1 className="text-2xl font-semibold font-heading">{threadName}</h1>
        </div>

        <ButtonGroup>
          <Button variant="outline">
            <HugeiconsIcon icon={Search01Icon} />
          </Button>
          <Button variant="outline" onClick={onShare}>
            <HugeiconsIcon icon={Share01Icon} />
          </Button>
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
            render={
              <Link to="/thread/$id/manage" params={{ id: threadId }}>
                <HugeiconsIcon icon={Setting06Icon} /> Manage
              </Link>
            }
          />
        </ButtonGroup>
      </div>
    </div>
  );
}
