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
| UI system           | shadcn/ui + Tailwind CSS |

## 3. Route map

| Route    | Purpose               |
| -------- | --------------------- |
| `/`      | Main reading view     |
| `/feeds` | Feed/group management |
| `/login` | Password/OIDC login   |

## 4. URL-driven app state

Main page state is kept in URL search params:

| Param     | Type                       | Meaning                  |
| --------- | -------------------------- | ------------------------ |
| `feed`    | number                     | Selected feed            |
| `group`   | number                     | Selected group           |
| `filter`  | `all \| unread \| starred` | Article filter           |
| `article` | number                     | Opened article in drawer |

This makes navigation, refresh, and deep-linking predictable.

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

### Main reading view (`/`)

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
- `Esc`: close search/settings/article drawer
- `j` / `ArrowDown`: next article
- `k` / `ArrowUp`: previous article

## 10. Authentication UX

- Password login is always available
- When OIDC is enabled, login page shows "Sign in with OIDC"
- OIDC callback failure is surfaced as `/login?error=oidc_failed`

## 11. Key files

- `frontend/src/routes/index.lazy.tsx`: main page composition
- `frontend/src/routes/feeds.lazy.tsx`: feed management page
- `frontend/src/routes/login.lazy.tsx`: login page
- `frontend/src/components/article/article-list.tsx`: list + tabs + bulk read
- `frontend/src/components/article/article-drawer.tsx`: article detail
- `frontend/src/components/feed/feed-list.tsx`: sidebar feed tree

## 12. Release verification checklist

- Type check: `cd frontend && npx tsc -b --noEmit`
- Lint: `cd frontend && pnpm lint`
- Production build: `cd frontend && pnpm build`
- Smoke test flows:
  - login/logout
  - search open + result navigation
  - feed/group selection and URL sync
  - read/unread + starred updates
  - `/feeds` page operations (add/edit/delete/import/export)
