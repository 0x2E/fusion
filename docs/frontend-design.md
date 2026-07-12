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
