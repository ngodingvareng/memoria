import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_app/mentions')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/_app/mentions"!</div>
}
