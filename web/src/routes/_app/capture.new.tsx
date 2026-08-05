import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_app/capture/new')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/_app/capture/new"!</div>
}
