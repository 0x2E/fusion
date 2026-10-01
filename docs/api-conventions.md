# API Conventions

The HTTP contract lives in the code: route inventory in
`backend/internal/handler/handler.go` (`SetupRouter`), payload shapes in
`backend/internal/model/model.go` (JSON tags), and the frontend call surface in
`frontend/src/lib/api/index.ts`. This page records only the cross-cutting rules
that are distributed across handlers, so they do not need to be rediscovered
per endpoint. If code and this page disagree, the code wins — fix the page.

## Base path and authentication

- All endpoints are under `/api`. The Fever compatibility API (`/fever*`) is
  separate: see `docs/fever-api.md`.
- Authentication is a session cookie named `session` (`POST /api/sessions`
  with `{"password": ...}` sets it; `DELETE /api/sessions` clears it, 204).
- Login attempts are rate limited: excess attempts get `429` with a
  `Retry-After` header (seconds).
- OIDC: `GET /api/oidc/enabled` always exists; `/api/oidc/login` and
  `/api/oidc/callback` are registered only when OIDC is configured. Login
  returns `{data: {auth_url}}` for the client to navigate to; the callback
  redirects (307) to `/` on success or `/login?error=oidc_failed` on failure.

## Response envelopes

- Single resource: `{"data": <object>}`
- Group and feed lists (small, fully returned): `{"data": [...], "total": n}`
- Item and bookmark lists (cursor-paginated):
  `{"data": [...], "total": n, "next_cursor": <string|null>}`;
  `next_cursor` is `null` when there are no more pages.

## Cursor pagination

- Pass the previous page's `next_cursor` as the `before` query parameter;
  omit it on the first page.
- Cursor format: `"<value>_<id>"` (two int64s, underscore-separated). The
  value is `pub_date` (unix) for items and `created_at` (unix) for bookmarks;
  `id` breaks ties.
- Ordering is always DESC. Items support `order_by=pub_date|created_at`
  (default `pub_date`); any other value silently falls back to `pub_date`.
- `limit` is capped at 100 (`maxListLimit` in `backend/internal/handler/item.go`).

## Errors

Always `{"error": "<message>"}`:

- `400` — invalid input (message varies)
- `401` — `"unauthorized"`
- `404` — `"<resource> not found"`
- `429` — `"too many login attempts"` (plus `Retry-After`)
- `500` — `"internal server error"` (details are logged server-side only)

## Status codes

- `200` — reads and creates (body present)
- `202` — feed refresh requests (`/feeds/refresh`, `/feeds/{id}/refresh`):
  refresh happens asynchronously, no body
- `204` — body-less mutations: mark items read/unread, all deletes, logout

## Request quirks worth knowing

- `POST /api/bookmarks` has two modes: `{item_id}` snapshots the bookmark from
  that item server-side (preferred), or a full `{link, title, content,
  feed_name, pub_date?}` snapshot when there is no item. Sending `item_id`
  makes the server ignore all other fields.
- `PATCH /api/items/-/read` and `-/unread` accept 1–1000 ids
  (`maxBatchUpdateIDs`).
- `PATCH /api/settings` is a partial update: an absent field leaves the stored
  value unchanged, and there is deliberately no way to clear a preference
  back to null. Closed sets: locale `en|zh|de|fr|es|ru|pt|sv`, theme
  `light|dark|system`, `auto_mark_read` `off|open|5|10|30`,
  `article_page_size` 1–100.
