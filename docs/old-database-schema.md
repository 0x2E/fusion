# Legacy Database Schema (Preserved)

> This document preserves the legacy schema snapshot for migration development and validation. SQL DDL is the single source of truth for table/column definitions.

## Full DDL (Authoritative)

```sql
CREATE TABLE `groups` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `created_at` datetime,
  `updated_at` datetime,
  `deleted_at` integer,
  `name` text NOT NULL
);

CREATE TABLE sqlite_sequence(name,seq);

CREATE UNIQUE INDEX `idx_name` ON `groups`(`deleted_at`,`name`);

CREATE TABLE `feeds` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `created_at` datetime,
  `updated_at` datetime,
  `deleted_at` integer,
  `name` text NOT NULL,
  `link` text NOT NULL,
  `last_build` datetime,
  `failure` text DEFAULT "",
  `group_id` integer,
  `suspended` numeric DEFAULT false,
  `req_proxy` text,
  `consecutive_failures` integer DEFAULT 0,
  CONSTRAINT `fk_feeds_group` FOREIGN KEY (`group_id`) REFERENCES `groups`(`id`)
);

CREATE TABLE `items` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `created_at` datetime,
  `updated_at` datetime,
  `deleted_at` integer,
  `title` text NOT NULL,
  `guid` text,
  `link` text,
  `content` text,
  `pub_date` datetime,
  `unread` numeric DEFAULT true,
  `feed_id` integer,
  `bookmark` numeric DEFAULT false,
  CONSTRAINT `fk_items_feed` FOREIGN KEY (`feed_id`) REFERENCES `feeds`(`id`)
);

CREATE INDEX `idx_items_feed_id` ON `items`(`feed_id`);
CREATE INDEX `idx_items_title` ON `items`(`title`);
CREATE INDEX `idx_items_unread` ON `items`(`unread`);
CREATE INDEX `idx_items_bookmark` ON `items`(`bookmark`);

CREATE UNIQUE INDEX `idx_guid` ON `items`(`deleted_at`,`guid`,`feed_id`);
CREATE UNIQUE INDEX `idx_link` ON `feeds`(`deleted_at`,`link`);
```

## Business Logic Notes

### 1) Entity relationships

- `groups -> feeds -> items` is a two-level one-to-many chain.
- `feeds.group_id` links each feed to a group.
- `items.feed_id` links each item to a feed.
- Legacy schema declares foreign keys (`fk_feeds_group`, `fk_items_feed`).

### 2) Soft-delete model

- All business tables use `deleted_at` (Unix timestamp) for logical delete.
- Deletes update `deleted_at`; rows are not physically removed.
- Unique indexes include `deleted_at`, so deleted historical rows can coexist with new active rows.

### 3) Deduplication and uniqueness semantics

- Group name uniqueness: `idx_name(deleted_at, name)`.
- Feed link uniqueness: `idx_link(deleted_at, link)`.
- Item deduplication key: `idx_guid(deleted_at, guid, feed_id)`.
- SQLite caveat: `NULL != NULL` in unique indexes, so multiple rows with `guid=NULL` are allowed (must be preserved in migrations).

### 4) Feed pull status fields (`feeds`)

- `last_build`: last content update time.
- `failure` + `consecutive_failures`: latest error and failure streak.
- `suspended`: pull disabled flag.
- `req_proxy`: per-feed proxy setting.

### 5) Reading state fields (`items`)

- `unread`: unread flag (default `true`).
- `bookmark`: bookmark flag (default `false`).
- Both are indexed for high-frequency filtering.

### 6) Application-level semantics

- App bootstrap ensures default group exists: `id=1`, `name="Default"`.
- `unread_count` is a virtual field in app models and is not persisted.

## Legacy Migration History

### After `v0.8.7`

- Added unique constraint for feed links via `idx_link`.
- Migration removed duplicate feeds and kept the row with the smallest `id`.
