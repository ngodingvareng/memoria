# Web Coding Style

This document standardizes how React components are structured and placed in
`web/`. It extends the `Features vs. components` note in the root `CLAUDE.md`
with concrete, enforceable rules. Every rule below is grounded in the
patterns already established in this codebase — this isn't a generic React
style guide, it's specific to how *this* project is built.

Treat this as a living document: when a new pattern gets agreed on in
review, add it here in the same PR.

## 1. One component per `.tsx` file

Each `.tsx` file defines exactly one React component. If a file grows a
second component — even a small private one only used via `.map()` inside
the first — extract it into its own file, named for the component in
kebab-case, next to the original.

```
features/threads/components/comment-list.tsx     # CommentList
features/threads/components/comment-row.tsx       # CommentRow (extracted)
```

The extracted file gets a normal named `export function`. It does **not**
need to be re-exported from the feature's `index.ts` barrel unless something
outside the file it was extracted from actually imports it — most extracted
sub-components stay private to their parent.

**Exception: `src/components/ui/`.** This directory is the ShadCN-generated
primitive library (per root `CLAUDE.md`, "treat these as generated/vendored
and prefer composing over heavily editing them"). ShadCN emits multiple
related parts (e.g. `Accordion`, `AccordionItem`, `AccordionTrigger`,
`AccordionContent`) into one file by convention upstream — do not split
these, so future `shadcn` CLI regeneration stays a clean diff.

**Not covered by this rule:** `.tsx` files that don't export a component at
all — e.g. `features/mentions/lib/render-text-with-mentions.tsx` exports a
lowercase helper function (`renderTextWithMentions`) that happens to return
`React.ReactNode`. It's a render helper, not a component, so it's exempt.
The rule targets component files under a `components/` directory (capitalized
`export function`/`export const` returning JSX as the file's own UI unit),
not every file that touches JSX.

## 2. Placement: `src/components/` vs. `src/features/<domain>/components/`

- **`src/components/`** — components that are general-purpose or used
  across more than one feature domain: the app shell (`app-header.tsx`,
  `app-sidebar.tsx`), standalone dialogs not tied to one feature's domain
  types (`components/dialogs/`), and infrastructure like
  `theme-provider.tsx`. `src/components/ui/` is reserved for the ShadCN
  primitive library specifically (see §1's exception).
- **`src/features/<domain>/components/`** — components specific to one
  product domain (threads, moments, circles, mentions, notifications,
  users, album, auth), operating on that domain's entities/DTOs. If a
  component only makes sense in the context of one feature, it lives there,
  even if it's currently only rendered from one place.

When in doubt: a component that imports domain-specific API hooks/types
(e.g. `useGetMomentsIdComments`, a `*Response` DTO) belongs in that
feature's `components/`. A component whose props are generic (ids, strings,
children, style props) and that has no feature-specific data dependency
belongs in `src/components/`.

## 3. Component props: `interface`, never `type`

Every component that takes props declares an `interface`, named
`<ComponentName>Props`, immediately above the component:

```tsx
interface CommentRowProps {
  momentId: string;
  comment: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoCommentResponse;
  momentOwnerId?: string;
}

function CommentRow({ momentId, comment, momentOwnerId }: CommentRowProps) {
  ...
}
```

Inline prop types (`function Foo({ x }: { x: string })`) and `type FooProps
= {...}` aliases are not allowed — extract a named interface even for a
single-field props shape.

This applies only to **component props**. Other `type` aliases colocated
with a component — a union like `export type Theme = 'dark' | 'light' |
'system'`, or a non-props shape like a context value type
(`ThemeProviderState`) — are unaffected and should stay `type`, since
`interface` can't express a union and a context value isn't a props
contract.

When props extend another component's props with additions (e.g. wrapping
`Avatar`), use `interface X extends Omit<ComponentProps<typeof Avatar>,
'children'> { ... }` rather than a `type` intersection (`Omit<...> &
{...}`) — `interface extends` works for any base that resolves to an
object type.
