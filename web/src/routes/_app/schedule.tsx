import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_app/schedule')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/_app/schedule"!</div>
}
