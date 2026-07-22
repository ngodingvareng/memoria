import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import { Input } from '@/components/ui/input';
import Wrapper from '@/components/wrapper';
import { Search01Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/stories')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <>
      <Wrapper>
        <ButtonGroup className="w-full">
          <Input className="h-10 text-base!" placeholder="Search..." />
          <Button variant="outline" aria-label="Search" className="h-10">
            <HugeiconsIcon icon={Search01Icon} />
          </Button>
        </ButtonGroup>
      </Wrapper>
    </>
  );
}
