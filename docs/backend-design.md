# Fusion Backend Design

## 1. Goals

- Keep the backend small, easy to self-host, and easy to reason about.
- Prefer explicit SQL over ORM abstractions.
- Preserve user content via bookmark snapshots.

## 2. Runtime architecture

Fusion backend runs two long-lived services in one process:

1. HTTP API server (Gin)
2. Feed pull worker (periodic and manual refresh)

Both services share the same SQLite store.

## 3. Tech stack

| Area           | Choice                                |
| -------------- | ------------------------------------- |
| Language       | Go 1.25                               |
| HTTP framework | Gin                                   |
| Database       | SQLite (`modernc.org/sqlite`)         |
| Migrations     | Embedded SQL files                    |
| Feed parser    | `github.com/mmcdole/gofeed`           |
| Feed discovery | `github.com/0x2E/feedfinder`          |
| Auth           | Password session auth + optional OIDC |

## 4. Module layout

```text
backend/
├── cmd/fusion/main.go           # process startup and lifecycle
├── internal/
│   ├── config/                  # env parsing
│   ├── handler/                 # HTTP handlers + middleware
│   ├── store/                   # SQL persistence + migrations
│   ├── pull/                    # fetch/parse/schedule/backoff
│   ├── auth/                    # password + OIDC helpers
│   ├── model/                   # API/storage models
│   └── pkg/httpc/               # HTTP client + SSRF guards
```

## 5. Database schema (current)

Source of truth: `backend/internal/store/migrations/001_initial.sql`.

### groups

- `id`, `name`, `created_at`, `updated_at`
- `name` is unique
- Default group: `id=1`, `name='Default'`

### feeds

- Core: `id`, `group_id`, `name`, `link`, `site_url`
- Pull status: `last_build`, `last_failure_at`, `failure`, `failures`, `suspended`
- Network: `proxy`
- Meta: `created_at`, `updated_at`
- Unique: `link`

### items

- `id`, `feed_id`, `guid`, `title`, `link`, `content`, `pub_date`, `unread`, `created_at`
- Unique: `(feed_id, guid)`
- Indexes: unread partial index, `pub_date` index, `(feed_id, unread)` index

### items full-text search

- Virtual table: `items_fts` (FTS5 on `title`, `content`)
- Triggers keep `items_fts` synchronized with `items`

### bookmarks

- Snapshot table: `item_id`, `link`, `title`, `content`, `pub_date`, `feed_name`, `created_at`
- `link` is unique
- `item_id` is nullable to preserve snapshots after source item deletion

## 6. Data integrity and cascade strategy

- No database foreign-key constraints.
- Cascade rules are explicit in store transactions:
  - Delete group: move feeds to group `1`, then delete group.
  - Delete feed: set matching bookmarks `item_id=NULL`, delete items, then delete feed.

This keeps behavior explicit and avoids hidden DB-level side effects.

## 7. API surface (high level)

- Sessions: login/logout
- OIDC: enabled status, login URL, callback
- Groups: list/get/create/update/delete
- Feeds: list/get/create/update/delete/validate/batch create/refresh
- Items: list/get/mark read/mark unread
- Search: feed + item search
- Bookmarks: list/get/create/delete

Detailed contract: `docs/openapi.yaml`.

## 8. Feed pull strategy

### Scheduler

- Pull interval: `FUSION_PULL_INTERVAL` (default 1800s)
- Concurrency limit: `FUSION_PULL_CONCURRENCY` (default 10)
- Request timeout: `FUSION_PULL_TIMEOUT` (default 30s)

### Skip policy

Periodic pull skips feed when:

- Feed is suspended, or
- Feed is in exponential backoff window, or
- Last successful update is within the regular interval.

### Backoff

- Formula: `interval * (1.8 ^ failures)`
- Capped by `FUSION_PULL_MAX_BACKOFF` (default 7 days)
- Failure timestamp source: `last_failure_at` (fallback to `last_build`)

### Manual refresh

- `POST /feeds/refresh`: refresh all non-suspended feeds
- `POST /feeds/:id/refresh`: refresh one feed
- Manual refresh bypasses periodic skip logic

## 9. Security model

- Password auth with bcrypt hash computed at startup
- Login attempt rate limit (`FUSION_LOGIN_*`)
- Session cookie: `HttpOnly`, `SameSite=Lax`, `Secure` on HTTPS
- Optional OIDC SSO (`FUSION_OIDC_*`)
- URL validation + private-network blocking by default for feed fetches
- CORS allowlist via `FUSION_CORS_ALLOWED_ORIGINS`
- Trusted proxy list via `FUSION_TRUSTED_PROXIES`

## 10. Observability and logs

- Structured logging via `log/slog`
- Configurable log level (`FUSION_LOG_LEVEL`)
- Configurable output format (`FUSION_LOG_FORMAT`: `auto`, `text`, `json`)

## 11. Release verification checklist

- Backend tests: `cd backend && go test ./...`
- Build check: `cd backend && go build -o /dev/null ./cmd/fusion`
- Migration sanity check: start app on empty DB and ensure schema bootstraps correctly
- API smoke tests: login, create feed, manual refresh, search, bookmark create/delete
