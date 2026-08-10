import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { ResetPasswordForm } from '@/features/auth';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_auth/reset-password')({
  validateSearch: (
    search: Record<string, unknown>
  ): { email?: string; token?: string } => ({
    email: typeof search.email === 'string' ? search.email : undefined,
    token: typeof search.token === 'string' ? search.token : undefined,
  }),
  component: RouteComponent,
});

function RouteComponent() {
  const { email, token } = Route.useSearch();

  return (
    <Card className="px-6">
      <CardHeader className="text-center">
        <CardTitle className="text-xl">Choose a new password</CardTitle>
        <CardDescription>
          {email ? `Resetting the password for ${email}` : ''}
        </CardDescription>
      </CardHeader>
      <CardContent>
        {email && token ? (
          <ResetPasswordForm email={email} token={token} />
        ) : (
          <Alert variant="destructive">
            <AlertDescription>
              This reset link is missing or malformed. Request a new one from
              the sign-in page.
            </AlertDescription>
          </Alert>
        )}
      </CardContent>
    </Card>
  );
}
