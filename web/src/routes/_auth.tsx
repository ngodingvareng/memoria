import { Brand } from '@/components/brand';
import { FieldDescription } from '@/components/ui/field';
import { GoogleOAuthProvider } from '@react-oauth/google';
import { createFileRoute, Outlet } from '@tanstack/react-router';

export const Route = createFileRoute('/_auth')({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <GoogleOAuthProvider clientId={import.meta.env.VITE_GOOGLE_CLIENT_ID}>
      <div className="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
        <div className="flex w-full max-w-lg flex-col gap-6">
          <div className="flex justify-center">
            <Brand />
          </div>
          <Outlet />
          <FieldDescription className="px-6 text-center">
            By clicking continue, you agree to our{' '}
            <a href="#">Terms of Service</a> and <a href="#">Privacy Policy</a>.
          </FieldDescription>
        </div>
      </div>
    </GoogleOAuthProvider>
  );
}
