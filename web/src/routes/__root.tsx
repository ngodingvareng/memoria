import { ThemeProvider } from '@/components/theme-provider';
import { Toaster } from '@/components/ui/sonner';
import { createRootRoute, Outlet } from '@tanstack/react-router';
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

const RootLayout = () => (
  <ThemeProvider defaultTheme="system" storageKey="vite-ui-theme">
    <Outlet />
    <Toaster />
    {/* <TanStackRouterDevtools position="bottom-right" /> */}
    {/* <ReactQueryDevtools buttonPosition="bottom-left" /> */}
  </ThemeProvider>
);

export const Route = createRootRoute({ component: RootLayout });
