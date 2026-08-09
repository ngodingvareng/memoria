import AppHeader from '@/components/app-header';
import AppSidebar from '@/components/app-sidebar';
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';
import { getSession, useSession } from '@/lib/session';
import { createFileRoute, Outlet, redirect, useNavigate } from '@tanstack/react-router';
import { useEffect } from 'react';

export const Route = createFileRoute('/_app')({
  beforeLoad: () => {
    if (!getSession()) {
      throw redirect({ to: '/signin' });
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  const session = useSession();
  const navigate = useNavigate();

  // beforeLoad only guards navigation into the app shell — it doesn't
  // react to the session dying while already inside it (e.g. the API
  // mutator's 401 refresh-and-retry failing because the refresh cookie
  // itself expired). Without this, a request would just fail silently
  // instead of sending the user back to sign in.
  useEffect(() => {
    if (!session) {
      navigate({ to: '/signin' });
    }
  }, [session, navigate]);

  return (
    <div className="[--header-height:calc(--spacing(16))]">
      <SidebarProvider className="flex flex-col">
        <AppHeader />
        <div className="flex pt-16 flex-1">
          <AppSidebar />
          <SidebarInset>
            <main>
              <Outlet />
            </main>
          </SidebarInset>
        </div>
      </SidebarProvider>
    </div>
  );
}
