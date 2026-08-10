import AppHeader from '@/components/app-header';
import AppSidebar from '@/components/app-sidebar';
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';
import { UserSettingsSidebar } from '@/features/users';
import { getSession, useSession } from '@/lib/session';
import {
  createFileRoute,
  Outlet,
  redirect,
  useNavigate,
} from '@tanstack/react-router';
import { useEffect } from 'react';

export const Route = createFileRoute('/_user')({
  beforeLoad: () => {
    const session = getSession();
    if (!session) {
      throw redirect({ to: '/signin' });
    }
    // Picking a username is mandatory — an account created but never
    // finished onboarding (e.g. closed the tab right after a Google
    // sign-in that created it) must not reach any inner page.
    if (!session.user.username) {
      throw redirect({ to: '/welcome' });
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  const session = useSession();
  const navigate = useNavigate();

  // beforeLoad only guards navigation into this shell — it doesn't
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
      <SidebarProvider defaultOpen={false} className="flex flex-col">
        <AppHeader />
        <div className="flex flex-1 pt-16">
          <UserSettingsSidebar />
          <SidebarInset>
            <AppSidebar />
            <main>
              <Outlet />
            </main>
          </SidebarInset>
        </div>
      </SidebarProvider>
    </div>
  );
}
