import { ThemeProvider } from '@/components/theme-provider';
import { createRootRoute, Outlet } from '@tanstack/react-router';
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools';

const RootLayout = () => (
  <ThemeProvider defaultTheme="system" storageKey="vite-ui-theme">
    <Outlet />
    <TanStackRouterDevtools position="bottom-right" />
  </ThemeProvider>
);

export const Route = createRootRoute({ component: RootLayout });
