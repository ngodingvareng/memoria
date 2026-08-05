import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_app/recap/$period')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/_app/recap/$period"!</div>
}
