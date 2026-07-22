import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_app/activities/$id/stats')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/activities/$id"!</div>
}
