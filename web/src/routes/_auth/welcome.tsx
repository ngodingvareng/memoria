import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { WelcomeForm } from '@/features/auth';
import { getSession } from '@/lib/session';
import { createFileRoute, redirect } from '@tanstack/react-router';

export const Route = createFileRoute('/_auth/welcome')({
  beforeLoad: () => {
    // Reached right after register, which starts the session — no
    // session means this wasn't reached through that flow (e.g. a
    // direct link after a reload, since the access token lives in
    // memory only).
    if (!getSession()) {
      throw redirect({ to: '/signin' });
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <Card className="px-6">
      <CardHeader className="text-center">
        <CardTitle className="text-xl">Choose a username</CardTitle>
        <CardDescription>
          This is how others find and mention you on Memoria
        </CardDescription>
      </CardHeader>
      <CardContent>
        <WelcomeForm />
      </CardContent>
    </Card>
  );
}
