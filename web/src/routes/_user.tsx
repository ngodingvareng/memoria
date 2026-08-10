import AppHeader from '@/components/app-header';
import AppSidebar from '@/components/app-sidebar';
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';
import { UserSettingsSidebar } from '@/features/users';
import { createFileRoute, Outlet } from '@tanstack/react-router';

export const Route = createFileRoute('/_user')({
  component: RouteComponent,
});

function RouteComponent() {
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
