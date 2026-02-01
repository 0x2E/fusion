# Fusion Backend Design

## 1. Tech Stack

| Item         | Choice                     | Notes                         |
| ------------ | -------------------------- | ----------------------------- |
| Framework    | Gin                        | Popular, well-documented      |
| Database     | modernc.org/sqlite         | Pure Go, no CGO               |
| ORM          | Pure SQL                   | Clearer, more control         |
| Foreign Keys | None                       | Cascade logic in app layer    |
| Timestamps   | Unix (INTEGER)             | Efficient, no timezone issues |
| Feed Parser  | github.com/mmcdole/gofeed  | RSS/Atom parsing              |
| Feed Finder  | github.com/0x2E/feedfinder | Discover feeds from HTML      |

## 2. Database Schema

```sql
-- groups
CREATE TABLE groups (
    id         INTEGER PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

INSERT INTO groups (id, name) VALUES (1, 'Default');

-- feeds
CREATE TABLE feeds (
    id         INTEGER PRIMARY KEY,
    group_id   INTEGER NOT NULL DEFAULT 1,
    name       TEXT NOT NULL,
    link       TEXT NOT NULL UNIQUE,
    site_url   TEXT DEFAULT '',
    last_build INTEGER DEFAULT 0,
    failure    TEXT DEFAULT '',
    failures   INTEGER DEFAULT 0,
    suspended  INTEGER DEFAULT 0,
    proxy      TEXT DEFAULT '',
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX idx_feeds_group_id ON feeds(group_id);

-- items
CREATE TABLE items (
    id         INTEGER PRIMARY KEY,
    feed_id    INTEGER NOT NULL,
    guid       TEXT NOT NULL,
    title      TEXT DEFAULT '',
    link       TEXT DEFAULT '',
    content    TEXT DEFAULT '',
    pub_date   INTEGER DEFAULT 0,
    unread     INTEGER DEFAULT 1,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE UNIQUE INDEX idx_items_feed_guid ON items(feed_id, guid);
CREATE INDEX idx_items_unread ON items(unread) WHERE unread = 1;
CREATE INDEX idx_items_pub_date ON items(pub_date DESC);

-- bookmarks (stores full snapshot)
CREATE TABLE bookmarks (
    id         INTEGER PRIMARY KEY,
    item_id    INTEGER,
    link       TEXT NOT NULL UNIQUE,
    title      TEXT DEFAULT '',
    content    TEXT DEFAULT '',
    pub_date   INTEGER DEFAULT 0,
    feed_name  TEXT DEFAULT '',
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX idx_bookmarks_created_at ON bookmarks(created_at DESC);
```

## 3. Cascade Logic (App Layer)

No foreign key constraints. Cascade handled in application:

```go
// Delete Group: move feeds to default group
UPDATE feeds SET group_id = 1 WHERE group_id = ?
DELETE FROM groups WHERE id = ?

// Delete Feed: preserve bookmark snapshots
UPDATE bookmarks SET item_id = NULL WHERE item_id IN (SELECT id FROM items WHERE feed_id = ?)
DELETE FROM items WHERE feed_id = ?
DELETE FROM feeds WHERE id = ?

```

## 4. API Endpoints

| Method | Path                  | Description              |
| ------ | --------------------- | ------------------------ |
| POST   | /api/sessions         | Login                    |
| DELETE | /api/sessions         | Logout                   |
| GET    | /api/groups           | List groups              |
| POST   | /api/groups           | Create group             |
| PATCH  | /api/groups/:id       | Update group             |
| DELETE | /api/groups/:id       | Delete group             |
| GET    | /api/feeds            | List feeds               |
| GET    | /api/feeds/:id        | Get feed                 |
| POST   | /api/feeds            | Create feed              |
| POST   | /api/feeds/batch      | Batch create feeds       |
| POST   | /api/feeds/validation | Validate feed URL        |
| PATCH  | /api/feeds/:id        | Update feed              |
| DELETE | /api/feeds/:id        | Delete feed              |
| POST   | /api/feeds/refresh    | Refresh feeds            |
| GET    | /api/items            | List items               |
| GET    | /api/items/:id        | Get item                 |
| PATCH  | /api/items/-/unread   | Batch update read status |
| GET    | /api/bookmarks        | List bookmarks           |
| POST   | /api/bookmarks        | Add bookmark             |
| DELETE | /api/bookmarks/:id    | Delete bookmark          |

## 5. Project Structure

```
backend/
├── cmd/fusion/main.go
├── internal/
│   ├── config/config.go
│   ├── auth/auth.go
│   ├── handler/
│   │   ├── handler.go
│   │   ├── session.go
│   │   ├── feed.go
│   │   ├── group.go
│   │   ├── item.go
│   │   └── bookmark.go
│   ├── store/
│   │   ├── db.go
│   │   ├── migrate.go
│   │   ├── feed.go
│   │   ├── group.go
│   │   ├── item.go
│   │   └── bookmark.go
│   ├── model/model.go
│   └── pull/
│       ├── puller.go
│       ├── fetcher.go
│       └── backoff.go
├── pkg/httpc/httpc.go
└── go.mod
```

## 6. Key Design Decisions

| Decision                | Choice             | Rationale                           |
| ----------------------- | ------------------ | ----------------------------------- |
| No soft delete          | Hard delete        | Simpler, bookmarks preserve content |
| No foreign keys         | App-layer cascade  | Avoids SQLite FK complexity         |
| Unix timestamps         | INTEGER            | Efficient, no timezone issues       |
| Bookmark snapshots      | Full content copy  | Survives feed/item deletion         |
| Partial index on unread | `WHERE unread = 1` | Smaller index, faster queries       |
| Single model file       | `model/model.go`   | Simple structs, no logic            |

## 7. Feed Pull Strategy

### 7.1 Configuration

| Parameter       | Value      | Notes                        |
| --------------- | ---------- | ---------------------------- |
| Pull interval   | 30 minutes | Fixed interval for all feeds |
| Request timeout | 30 seconds | Per-request timeout          |
| Max concurrency | 10         | Concurrent feed fetches      |
| Max backoff     | 7 days     | Maximum retry delay          |
| Backoff base    | 1.8        | Exponential backoff base     |

### 7.2 Exponential Backoff

Formula: `backoff = interval × (1.8 ^ failures)`

| Failures | Backoff Time |
| -------- | ------------ |
| 1        | 54 min       |
| 2        | 97 min       |
| 3        | 175 min      |
| 5        | 9.5 hours    |
| 10+      | 7 days (max) |

### 7.3 Pull Decision Flow

```
Start Pull
  ↓
suspended = true? → Skip
  ↓
failures > 0?
  ├─ Yes → Check cooldown period → Not expired → Skip
  ↓
Last update < 30 min ago? → Skip
  ↓
Fetch feed
  ├─ Success → failures = 0, update last_build
  └─ Failure → failures++, record error in failure field
```

### 7.4 HTTP Client

```go
http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        DisableKeepAlives: true,   // Avoid connection pool overhead
        ForceAttemptHTTP2: true,   // Prefer HTTP/2
        Proxy: /* Use feed.proxy if set */
    },
}

// Headers
User-Agent: fusion/1.0
Connection: close
```

### 7.5 RSS Parsing Rules

| Field    | Priority                  | Notes                       |
| -------- | ------------------------- | --------------------------- |
| guid     | GUID → Link               | Unique identifier for dedup |
| content  | Content → Description     | Prefer full content         |
| pub_date | PublishedParsed → Updated | Publication timestamp       |
| link     | Resolve relative URLs     | Convert to absolute URLs    |

### 7.6 Feed Discovery

When adding a new feed:

```
User inputs URL
  ↓
Try parsing as feed directly
  ├─ Success → Return single feed
  ↓
Parse HTML, find <link type="application/rss+xml"> etc.
  ↓
Return discovered feeds for user selection
```
