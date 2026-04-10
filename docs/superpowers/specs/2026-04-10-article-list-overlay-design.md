# Article List Overlay and Pagination Design

## Summary

Fusion currently has two related classes of article-list bugs:

1. The `starred` view only loads the first 100 bookmarks because the frontend fetches a single `bookmarks` page with `limit=100` and never paginates.
2. The `unread` view can stop loading additional unread items after local read-state changes because frontend pagination uses `offset`, while the filtered server result set shrinks after articles are marked read.

The intended UX is preserved:

- After marking an article read in the `unread` view, the article should remain visible in the current list.
- The article should change visually to its new state.
- The article should not disappear automatically.
- The list should only fully reconcile with server membership when the user explicitly refreshes or changes filters.

This design fixes those bugs without removing the current optimistic and undo-friendly interaction model.

## Goals

- Fix `starred` so it supports pagination beyond 100 entries.
- Fix `unread` so local read-state changes do not break further pagination.
- Preserve local visual state for read/star operations until explicit refresh or filter change.
- Keep React Query as the source of server data and avoid moving full article collections into a global store.
- Make the solution reusable across `all`, `unread`, and `starred` article views.

## Non-Goals

- No backend cursor API in this phase.
- No broad rewrite of feed/group state management.
- No new persistence format for pending UI state.
- No change to the product decision that modified entries remain visible until refresh or filter change.

## Root Cause

### Starred

`bookmarks` are fetched via a single query that requests `limit=100, offset=0`. The `starred` list, the sidebar starred count, and starred navigation all derive from that local array, so they are all capped by the first page.

### Unread

`items` uses infinite pagination with `offset = total number of fetched rows`. That is valid only if the filtered server result set remains stable between requests. In the `unread` view, local and server read-state changes shrink the matching result set. When the next request still uses the old offset, later unread items are skipped.

The bug is not caused by optimistic UI itself. It is caused by combining:

- membership-changing filters (`unread`, `starred`),
- `offset` pagination, and
- a UI rule that keeps modified rows visible in the current list.

## Design Overview

Introduce a lightweight article-list session overlay layer that sits on top of paged server data.

The model has two distinct parts:

1. **Server baseline**
   React Query stores paginated `items` and `bookmarks` responses exactly as returned by the API.
2. **Session overlay**
   A small frontend session store tracks local read/star overrides and whether already-visible entries should remain visible in the current list even if they no longer satisfy the active filter.

This separation lets pagination continue from server pages while the rendered list continues to honor the current undo-friendly UX.

## Data Model

Add an article list session store keyed by the active list context.

List context fields:

- top-level filter: `all | unread | starred`
- `feedId`
- `groupId`

Per-context state:

- `readOverrides: Record<number, boolean>`
- `starOverrides: Record<number, boolean>`
- `stickyVisibleIds: Record<number, true>`

Rules:

- In `all`, overrides only affect visuals.
- In `unread`, marking an already-visible item as read sets `readOverrides[id] = false` and `stickyVisibleIds[id] = true`.
- In `starred`, un-starring an already-visible item sets `starOverrides[id] = false` and `stickyVisibleIds[id] = true`.
- Refreshing the list, changing the top-level filter, changing the selected feed, or changing the selected group clears the current context overlay.

The overlay is session-local and intentionally non-persistent.

## Rendering Rules

### All

- Render paged server items.
- Apply read/star overrides only to display state.
- Do not change list membership locally.

### Unread

- Baseline membership comes from paged server items with `unread=true`.
- If an item is already visible and later marked read, keep it rendered in the current list using `stickyVisibleIds`.
- Its visual state reflects the local read override.
- The item remains until explicit refresh or filter/context change.

### Starred

- Baseline membership comes from paged bookmark results.
- If an item is already visible and later unstarred, keep it rendered in the current list using `stickyVisibleIds`.
- Its visual state reflects the local star override.
- The item remains until explicit refresh or filter/context change.

## Pagination Rules

The next page cursor must be derived from the number of server rows fetched, not from the number of rows currently rendered.

That means:

- sticky rows do not affect `offset`
- local membership overrides do not affect `offset`
- the query layer only counts baseline server pages when computing `getNextPageParam`

This keeps pagination stable even if the visible list contains items that no longer match the current filter.

## Query Changes

### Items

- Keep `useInfiniteQuery` for article items.
- Preserve current server pagination and totals.
- Ensure `getNextPageParam` is based only on accumulated server page lengths.
- Add a derived selector/hook that combines baseline pages with the session overlay into the rendered article list.

### Bookmarks

- Replace the single-page bookmarks query with an infinite query.
- Keep a derived bookmark lookup for fast `isItemStarred` and `getBookmarkByItemId` checks.
- Expose bookmark `total` separately from the loaded page count.
- Render `starred` articles from paginated bookmark pages plus the overlay.
- Use bookmark `total` for sidebar starred count.

## Component Boundaries

### Query Layer

Responsibilities:

- fetch baseline paged data from APIs
- expose baseline lookups and totals
- remain independent from visual sticky behavior

### Article List Session Store

Responsibilities:

- track session-local overrides
- track sticky visibility for the active context
- provide clear/reset operations on explicit reconciliation boundaries

### Derived Article List Hook

Responsibilities:

- merge baseline query results with the session overlay
- provide rendered rows for `ArticleList` and `ArticleDrawer`
- centralize membership and display rules per filter

This removes the current fragmentation where local state is partly held in component state and partly inferred from query caches.

## Reconciliation Boundaries

The overlay is cleared when the user intentionally leaves the current session view.

Clear on:

- manual refresh of the current list
- switching top-level filter
- switching feed
- switching group
- re-entering the page with a new context

Do not clear on:

- mark read/unread
- star/unstar
- loading more pages inside the same context

## Error Handling

- Keep optimistic mutation rollback for failed read/star operations.
- On mutation failure, restore both query cache changes and overlay state.
- Derived list hooks must tolerate missing item details or partially loaded bookmark pages.
- If a lookup cannot be resolved for a sticky row, keep rendering the best available snapshot rather than removing it abruptly.

## Testing Strategy

### Frontend query and state tests

- `unread` pagination continues to fetch later unread rows after local read changes.
- `starred` pagination loads beyond the first 100 bookmarks.
- sidebar starred count uses server `total`, not loaded page length.
- already-visible rows stay visible after local membership-changing actions.
- refresh/filter/context change clears sticky rows and overrides.

### Component behavior tests

- `ArticleList` shows modified entries with updated visuals while keeping them in place.
- `ArticleDrawer` navigation uses the same derived list as `ArticleList`.

### Regression tests

- batch mark-as-read in `unread` does not prevent loading more unread items.
- un-starring visible rows in `starred` does not break subsequent pagination.

## Implementation Plan

### Phase 1

- Add article list session overlay store.
- Introduce a shared derived article-list hook.
- Fix unread pagination to use baseline page counts only.
- Convert bookmarks to infinite pagination.
- Update starred count to use bookmark total.

### Phase 2

- Move remaining per-component override state into the shared overlay model.
- Ensure list, drawer, and keyboard navigation all consume the same derived article collection.
- Remove duplicated local starred/unread state logic from components.

## Alternatives Considered

### Local patch only

Rejected because it would fix the current symptoms without addressing the structural mismatch between sticky list semantics and membership-changing paged filters.

### Immediate server reconciliation after every action

Rejected because it conflicts with the intended undo-friendly UX.

### Backend cursor pagination first

Deferred because it increases scope across frontend and backend APIs. The current bugs can be fixed cleanly in the frontend with a smaller first step.

## Expected Outcome

After this design is implemented:

- `starred` can paginate through all bookmarks.
- `unread` can keep modified rows visible without breaking load-more behavior.
- article list behavior is consistent across list, drawer, and navigation.
- the current undo-friendly UX is preserved, but the pagination model no longer depends on unstable rendered membership.
