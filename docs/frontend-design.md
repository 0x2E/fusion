# Fusion Frontend Design

## 1. Goals

- Keep interactions fast and keyboard-friendly.
- Keep state predictable by encoding major UI state in URL params.
- Prioritize readability and simple information architecture.

## 2. Stack decisions

Dependencies live in `frontend/package.json` — do not restate them here.
The choices worth knowing: TanStack Router because routes are typed files
under `frontend/src/routes/` (the route inventory is those files, not a
table here); TanStack Query owns server-state caching so Zustand holds
only transient UI state; shadcn/ui components are vendored into
`src/components/ui/` and regenerated via the CLI rather than hand-edited
(see AGENTS.md).

## 3. URL-driven app state

Reading state is split between path params and search params:

| Location     | Key       | Type                       | Meaning                  |
| ------------ | --------- | -------------------------- | ------------------------ |
| Path param   | `filter`  | `all \| unread \| starred` | Active article filter    |
| Path param   | `feedId`  | number                     | Selected feed scope      |
| Path param   | `groupId` | number                     | Selected group scope     |
| Search param | `article` | number                     | Opened article in drawer |

Examples:

- `/unread`
- `/feeds/6/unread`
- `/groups/3/starred?article=289`

This keeps list context stable while opening/closing article detail. `g u` /
`g a` / `g s` / `g f` view jumps are navigation sugar over the same URLs.

## 4. Layout system

### Desktop

- Fixed left sidebar (`300px`)
- Main content on the right
- Article detail opens in right-side drawer

### Mobile

- Sidebar is a left sheet/drawer
- Main content remains single-column
- Modals/drawers share the same UI flow as desktop

The reading view is an infinite list with filter tabs and per-card
read/star quick actions; the article drawer shows sanitized full content
(`lib/content.ts`) with source link and prev/next navigation; feed
management (`/feeds`) hosts grouped feed/group CRUD, refresh-all, and
OPML import/export. Component names map to these areas directly — browse
`frontend/src/components/` for the implementation.

## 5. Data flow

- API layer lives in `frontend/src/lib/api/`
- Query logic lives in `frontend/src/queries/`
- TanStack Query handles:
  - cursor-based pagination for items and bookmarks (opaque `next_cursor` passed as the `before` query param; `next_cursor` is null when no more pages exist)
  - optimistic read/unread updates
  - cache invalidation after mutations
- Zustand stores transient UI state (dialogs, mobile sidebar, edit targets).
  The preferences store is the exception: it is a localStorage-backed local
  cache of the user's synced settings (`lib/settings-sync.ts` pulls the
  server values once per page load before the first authenticated screen
  renders, server non-null values win, and each settings-dialog change
  PATCHes only the changed key). Theme lives in next-themes' own
  localStorage key and is mirrored through the same sync.

### Read-state changes: grayed rows are an undo affordance, not stale UI

Marking an item read in the **unread** view keeps the row in the list, grayed out, instead of removing it. This is deliberate: the row can be toggled back to unread in place, without hunting it down in the all view. Items marked read from any other view (all, starred, search) are committed immediately.

The "still undoable" state is scoped to the exact list being viewed (feed + group + filter identity):

- Read-state mutations update cached items in place and deliberately never invalidate item list queries — list membership stays frozen at fetch time.
- The unread view renders `unread || pinned`. Pins are per-list-identity view state (`frontend/src/store/article-pins.ts`), written only when marking read inside an unread view.
- Leaving the list — switching filter or feed/group scope, or unmounting the reading view — clears the pins and commits: read rows drop out of the unread view without a refetch. The all view still shows them, so undo remains possible after the fact.
- Keyboard navigation is unaffected: pinned gray rows stay in the article id list, keeping `j`/`k` indices stable.

Known edge: marking an item unread from the all view does not insert it into an already-cached unread list; cache staleness/refetch covers it eventually.

Star/bookmark changes are deliberately the opposite pattern: un-starring removes the row from the starred view immediately (optimistic removal from every bookmark list cache, then invalidation to reconcile — see `useDeleteBookmark`), because a bookmark toggle is its own undo and needs no lingering row. Only read-state changes defer list membership.

Before filing "marked-read items still show up in the unread list" as a bug, check which half is reported: grayed rows while staying in the same unread list are this design working; rows surviving a filter/scope switch was the actual defect, fixed by the pin scoping (#262, #266).

## 7. Search and bookmarks

- Unified search (`GET /api/search`) returns feeds and items in one
  request; results open the feed context or article drawer.
- The "starred" view is powered by bookmarks, which are content
  snapshots — starred items survive deletion of the source feed/item
  (see the bookmarks rationale in `backend-design.md` §5).

## 8. Keyboard interactions

The app is keyboard-first. Shortcut categories: search/dialog toggles, article navigation (next/previous), read/star toggles, view jumps (`g u` / `g a` / `g s` / `g f`), and `?` for in-app help. The authoritative binding list lives in the shortcuts help dialog and its source; avoid duplicating it here so docs and code do not drift.

## 9. Authentication UX

- Password login is available when password auth is enabled
- If password is empty and OIDC is not configured, the UI is directly accessible without `/login`
- When OIDC is enabled, login page shows "Sign in with OIDC"
- OIDC callback failure is surfaced as `/login?error=oidc_failed`
