# ReedMe Frontend Design

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
| UI system           | shadcn/ui + Tailwind CSS |

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
  - infinite pagination for items
  - optimistic read/unread updates
  - cache invalidation after mutations
- Zustand stores transient UI state only (dialogs, mobile sidebar, edit targets)

## 8. Search and bookmarks

### Unified search

- Endpoint: `GET /api/search`
- Searches feeds and items in one request
- Results open feed context or article drawer

### Starred model

- "Starred" view is powered by bookmarks
- Bookmarks are content snapshots, so starred items survive source deletion

## 9. Keyboard interactions

Implemented shortcuts:

- `Cmd/Ctrl + K`: toggle search dialog
- `Cmd/Ctrl + ,`: open settings dialog
- `Esc`: close search/settings/article drawer
- `/`: open search dialog
- `?`: open keyboard shortcuts help
- `j` / `n` / `ArrowDown`: next article
- `k` / `p` / `ArrowUp`: previous article
- `m`: toggle read/unread for current article
- `s` / `f`: toggle star for current article
- `o` / `v`: open current article in browser
- `g u`: go to unread
- `g a`: go to all
- `g s`: go to starred
- `g f`: go to feed management

Shortcut help entry points:

- Sidebar search button hint shows `Cmd+K / ?`
- Search dialog Quick Actions includes a "Keyboard Shortcuts" item
- Settings > Appearance includes a "Keyboard Shortcuts" section

## 10. Authentication UX

- Password login is always available
- When OIDC is enabled, login page shows "Sign in with OIDC"
- OIDC callback failure is surfaced as `/login?error=oidc_failed`

## 11. Key files

- `frontend/src/routes/$filter.lazy.tsx`: top-level reading page
- `frontend/src/routes/feeds_.$feedId.$filter.lazy.tsx`: feed-scoped reading page
- `frontend/src/routes/groups.$groupId.$filter.lazy.tsx`: group-scoped reading page
- `frontend/src/routes/feeds.lazy.tsx`: feed management page
- `frontend/src/routes/login.lazy.tsx`: login page
- `frontend/src/components/article/article-page.tsx`: shared reading page wrapper
- `frontend/src/components/article/article-list.tsx`: list + tabs + bulk read
- `frontend/src/components/article/article-drawer.tsx`: article detail
- `frontend/src/components/feed/feed-list.tsx`: sidebar feed tree

## 12. Route file naming note

TanStack file-based routes can infer parent-child nesting from file names. We intentionally keep management page `/feeds` and reading page `/feeds/:feedId/:filter` as separate route trees.

- `frontend/src/routes/feeds.lazy.tsx` maps to management page `/feeds`
- `frontend/src/routes/feeds_.$feedId.$filter.lazy.tsx` maps to reading page `/feeds/:feedId/:filter`

The `feeds_` prefix is a routing implementation detail to avoid accidental nesting under the management page route while preserving the final URL path as `/feeds/...`.

## 13. Release verification checklist

- Type check: `cd frontend && npx tsc -b --noEmit`
- Lint: `cd frontend && pnpm lint`
- Production build: `cd frontend && pnpm build`
- Smoke test flows:
  - login/logout
  - search open + result navigation
  - feed/group selection and URL sync
  - read/unread + starred updates
  - `/feeds` page operations (add/edit/delete/import/export)
