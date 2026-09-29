# Fusion Frontend Design

## 1. Goals

- Keep interactions fast and keyboard-friendly.
- Keep state predictable by encoding major UI state in URL params.
- Prioritize readability and simple information architecture.

## 2. Tech stack

| Area                | Choice                   |
| ------------------- | ------------------------ |
| Framework           | React 19 + TypeScript    |
| Build               | Vite                     |
| Router              | TanStack Router          |
| Data fetching/cache | TanStack Query           |
| State               | Zustand (UI-only state)  |
| UI system           | shadcn/ui (Base UI) + Tailwind CSS |

## 3. Route map

| Route pattern              | Purpose                         |
| -------------------------- | ------------------------------- |
| `/`                        | Canonical redirect to `/unread` |
| `/:filter`                 | Top-level reading view          |
| `/feeds/:feedId/:filter`   | Feed-scoped reading view        |
| `/groups/:groupId/:filter` | Group-scoped reading view       |
| `/feeds`                   | Feed/group management           |
| `/login`                   | Password/OIDC login             |

## 4. URL-driven app state

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

This keeps list context stable while opening/closing article detail.

## 5. Layout system

### Desktop

- Fixed left sidebar (`300px`)
- Main content on the right
- Article detail opens in right-side drawer

### Mobile

- Sidebar is a left sheet/drawer
- Main content remains single-column
- Modals/drawers share the same UI flow as desktop

## 6. Core UI areas

### Sidebar

- App branding
- Search entry (`Cmd/Ctrl + K`)
- Feed tree (All, groups, feeds)
- Footer actions: Manage Feeds, Settings

### Main reading view (`/:filter`, `/feeds/:feedId/:filter`, `/groups/:groupId/:filter`)

- Header with page title and "Mark all as read"
- Filter tabs: All / Unread / Starred
- Infinite article list (load more)
- Article cards with quick actions (read/unread, star)

### Article drawer

- Shows full article content (sanitized HTML)
- Supports previous/next navigation
- Includes source link and feed metadata

### Feed management (`/feeds`)

- Grouped feed list with search + status filter
- Group actions: rename, delete (except default group)
- Feed actions: edit
- Bulk/system actions: refresh all, OPML import/export, add feed/group

## 7. Data flow

- API layer lives in `frontend/src/lib/api/`
- Query logic lives in `frontend/src/queries/`
- TanStack Query handles:
  - cursor-based pagination for items and bookmarks (opaque `next_cursor` passed as the `before` query param; `next_cursor` is null when no more pages exist)
  - optimistic read/unread updates
  - cache invalidation after mutations
- Zustand stores transient UI state only (dialogs, mobile sidebar, edit targets)

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

## 8. Search and bookmarks

### Unified search

- Endpoint: `GET /api/search`
- Searches feeds and items in one request
- Results open feed context or article drawer

### Starred model

- "Starred" view is powered by bookmarks
- Bookmarks are content snapshots, so starred items survive source deletion

## 9. Keyboard interactions

The app is keyboard-first. Shortcut categories: search/dialog toggles, article navigation (next/previous), read/star toggles, view jumps (`g u` / `g a` / `g s` / `g f`), and `?` for in-app help. The authoritative binding list lives in the shortcuts help dialog and its source; avoid duplicating it here so docs and code do not drift.

## 10. Authentication UX

- Password login is available when password auth is enabled
- If password is empty and OIDC is not configured, the UI is directly accessible without `/login`
- When OIDC is enabled, login page shows "Sign in with OIDC"
- OIDC callback failure is surfaced as `/login?error=oidc_failed`
