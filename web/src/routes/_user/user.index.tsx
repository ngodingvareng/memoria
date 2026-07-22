import { createFileRoute, redirect } from '@tanstack/react-router';

export const Route = createFileRoute('/_user/user/')({
  beforeLoad: () => {
    throw redirect({
      to: '/user/account',
    });
  },
});
