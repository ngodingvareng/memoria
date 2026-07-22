import { Button } from '@/components/ui/button';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_app/')({
  component: Index,
});

function Index() {
  return (
    <div className="p-2">
      <h1 className="font-heading">Margarin</h1>
      <h3>Welcome Home!</h3>
      <Button>Hello World</Button>
    </div>
  );
}
