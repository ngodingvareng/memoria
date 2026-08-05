import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_app/activity/$id/captures/$captureId')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/_app/activity/$id/captures/$captureId"!</div>
}
