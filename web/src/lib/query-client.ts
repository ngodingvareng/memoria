import { ApiError } from '@/lib/api-client';
import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      // ApiError means the server actually answered (4xx/5xx) — retrying
      // that blindly makes no sense. Anything else (network hiccup) gets
      // a couple of attempts.
      retry: (failureCount, error) =>
        !(error instanceof ApiError) && failureCount < 2,
    },
    mutations: {
      retry: false,
    },
  },
});
