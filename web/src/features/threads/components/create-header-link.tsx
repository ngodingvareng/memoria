import { Button } from '@/components/ui/button';
import { PlusSignIcon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { Link } from '@tanstack/react-router';

export function CreateHeaderLink() {
  return (
    <div className="flex gap-2">
      <Button
        size="lg"
        render={
          <Link to="/thread/new">
            <HugeiconsIcon icon={PlusSignIcon} /> New
          </Link>
        }
      />
    </div>
  );
}
