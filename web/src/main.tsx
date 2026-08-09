import { StrictMode } from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider, createRouter } from '@tanstack/react-router';
import { QueryClientProvider } from '@tanstack/react-query';

import { routeTree } from './routeTree.gen';
import { queryClient } from '@/lib/query-client';
import { restoreSession } from '@/features/auth/api/session-bootstrap';
import { getSession } from '@/lib/session';

const router = createRouter({ routeTree });

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}

async function bootstrap() {
  const rootElement = document.getElementById('root')!;
  if (rootElement.innerHTML) return;

  // Exchange the httpOnly refresh cookie for an access token before the
  // router ever renders, so a page reload doesn't bounce a still-logged-in
  // user to /signin just because the in-memory access token was lost.
  if (!getSession()) {
    await restoreSession();
  }

  const root = ReactDOM.createRoot(rootElement);
  root.render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </StrictMode>
  );
}

bootstrap();
