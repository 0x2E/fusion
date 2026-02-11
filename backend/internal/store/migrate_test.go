package store

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestMigrate(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	// Verify all expected tables exist
	tables := []string{"groups", "feeds", "items", "bookmarks", "schema_migrations", "items_fts"}
	for _, table := range tables {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=:table"
		err := store.db.QueryRow(query, sql.Named("table", table)).Scan(&count)
		if err != nil {
			t.Errorf("failed to check table %s: %v", table, err)
			continue
		}
		if count != 1 {
			t.Errorf("expected table %s to exist, but it doesn't", table)
		}
	}
}

func TestMigrateIdempotent(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	// Get initial migration count
	var initialCount int
	err := store.db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&initialCount)
	if err != nil {
		t.Fatalf("failed to query initial migration count: %v", err)
	}

	// Run migrate again
	if err := store.migrate(); err != nil {
		t.Fatalf("second migrate() call failed: %v", err)
	}

	// Verify count hasn't changed
	var finalCount int
	err = store.db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&finalCount)
	if err != nil {
		t.Fatalf("failed to query final migration count: %v", err)
	}

	if finalCount != initialCount {
		t.Errorf("migrations were re-applied: initial=%d, final=%d", initialCount, finalCount)
	}
}

func TestMigrateLegacySchema(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy.db")
	createLegacyDBFixture(t, dbPath)

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() failed for legacy schema: %v", err)
	}
	defer closeStore(t, store)

	backupFiles, err := filepath.Glob(dbPath + ".bak.*")
	if err != nil {
		t.Fatalf("glob backup files failed: %v", err)
	}
	if len(backupFiles) == 0 {
		t.Fatal("expected backup file to be created before legacy migration")
	}

	var versionCount int
	err = store.db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = 1`).Scan(&versionCount)
	if err != nil {
		t.Fatalf("query schema version failed: %v", err)
	}
	if versionCount != 1 {
		t.Fatalf("expected baseline migration version to be recorded once, got %d", versionCount)
	}

	var group1Name string
	err = store.db.QueryRow(`SELECT name FROM groups WHERE id = 1`).Scan(&group1Name)
	if err != nil {
		t.Fatalf("query default group failed: %v", err)
	}
	if group1Name != "Legacy Default" {
		t.Errorf("expected default group name migrated from legacy, got %q", group1Name)
	}

	var feedCount int
	err = store.db.QueryRow(`SELECT COUNT(*) FROM feeds`).Scan(&feedCount)
	if err != nil {
		t.Fatalf("count feeds failed: %v", err)
	}
	if feedCount != 3 {
		t.Fatalf("expected 3 active deduplicated feeds, got %d", feedCount)
	}

	var (
		feed1GroupID     int64
		feed1Failures    int64
		feed1Suspended   int
		feed1Proxy       string
		feed1SiteURL     string
		feed1LastBuild   int64
		feed1LastFailure int64
	)
	err = store.db.QueryRow(`
		SELECT group_id, failures, suspended, proxy, site_url, last_build, last_failure_at
		FROM feeds
		WHERE link = 'https://legacy.example/feed-a.xml'
	`).Scan(&feed1GroupID, &feed1Failures, &feed1Suspended, &feed1Proxy, &feed1SiteURL, &feed1LastBuild, &feed1LastFailure)
	if err != nil {
		t.Fatalf("query migrated feed failed: %v", err)
	}
	if feed1GroupID != 2 {
		t.Errorf("expected feed-a group_id=2, got %d", feed1GroupID)
	}
	if feed1Failures != 3 {
		t.Errorf("expected failures=3, got %d", feed1Failures)
	}
	if feed1Suspended != 1 {
		t.Errorf("expected suspended=1, got %d", feed1Suspended)
	}
	if feed1Proxy != "http://127.0.0.1:8080" {
		t.Errorf("expected proxy migrated, got %q", feed1Proxy)
	}
	if feed1SiteURL != "" {
		t.Errorf("expected empty site_url default, got %q", feed1SiteURL)
	}
	if feed1LastBuild == 0 {
		t.Error("expected last_build to be converted to unix timestamp")
	}
	if feed1LastFailure != 0 {
		t.Errorf("expected last_failure_at default to 0, got %d", feed1LastFailure)
	}

	var orphanGroupID int64
	err = store.db.QueryRow(`SELECT group_id FROM feeds WHERE link = 'https://legacy.example/feed-b.xml'`).Scan(&orphanGroupID)
	if err != nil {
		t.Fatalf("query orphan feed failed: %v", err)
	}
	if orphanGroupID != 1 {
		t.Errorf("expected orphan feed moved to default group 1, got %d", orphanGroupID)
	}

	var dedupGroupID int64
	err = store.db.QueryRow(`SELECT group_id FROM feeds WHERE link = 'https://legacy.example/feed-c.xml'`).Scan(&dedupGroupID)
	if err != nil {
		t.Fatalf("query dedup-group feed failed: %v", err)
	}
	if dedupGroupID != 2 {
		t.Errorf("expected feed-c mapped to canonical group 2, got %d", dedupGroupID)
	}

	var itemCount int
	err = store.db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&itemCount)
	if err != nil {
		t.Fatalf("count items failed: %v", err)
	}
	if itemCount != 6 {
		t.Fatalf("expected 6 active deduplicated items, got %d", itemCount)
	}

	var remappedFeedID int64
	err = store.db.QueryRow(`SELECT feed_id FROM items WHERE guid = 'guid-a3'`).Scan(&remappedFeedID)
	if err != nil {
		t.Fatalf("query remapped item failed: %v", err)
	}
	if remappedFeedID != 1 {
		t.Errorf("expected guid-a3 remapped to canonical feed 1, got %d", remappedFeedID)
	}

	var generatedGUID string
	err = store.db.QueryRow(`SELECT guid FROM items WHERE id = 9`).Scan(&generatedGUID)
	if err != nil {
		t.Fatalf("query generated guid item failed: %v", err)
	}
	if generatedGUID != "legacy-item-9" {
		t.Errorf("expected generated guid legacy-item-9, got %q", generatedGUID)
	}

	var bookmarkCount int
	err = store.db.QueryRow(`SELECT COUNT(*) FROM bookmarks`).Scan(&bookmarkCount)
	if err != nil {
		t.Fatalf("count bookmarks failed: %v", err)
	}
	if bookmarkCount != 3 {
		t.Fatalf("expected 3 bookmarks migrated from item bookmark flag, got %d", bookmarkCount)
	}

	var ftsCount int
	err = store.db.QueryRow(`SELECT COUNT(*) FROM items_fts`).Scan(&ftsCount)
	if err != nil {
		t.Fatalf("count items_fts failed: %v", err)
	}
	if ftsCount != itemCount {
		t.Errorf("expected items_fts count=%d, got %d", itemCount, ftsCount)
	}
}

func TestMigrateLegacySchemaDespiteExistingMigrationRecord(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy_with_marker.db")
	createLegacyDBFixture(t, dbPath)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open legacy db failed: %v", err)
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at INTEGER NOT NULL DEFAULT (unixepoch())
		);
		INSERT OR IGNORE INTO schema_migrations (version) VALUES (999);
	`); err != nil {
		_ = db.Close()
		t.Fatalf("seed schema_migrations marker failed: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close legacy db failed: %v", err)
	}

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() failed for legacy schema with existing migration record: %v", err)
	}
	defer closeStore(t, store)

	var deletedAtColumnCount int
	err = store.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('groups') WHERE name = 'deleted_at'`).Scan(&deletedAtColumnCount)
	if err != nil {
		t.Fatalf("check groups.deleted_at column failed: %v", err)
	}
	if deletedAtColumnCount != 0 {
		t.Fatalf("expected migrated groups schema without deleted_at, got count=%d", deletedAtColumnCount)
	}

	var versionCount int
	err = store.db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = 1`).Scan(&versionCount)
	if err != nil {
		t.Fatalf("query baseline migration version failed: %v", err)
	}
	if versionCount != 1 {
		t.Fatalf("expected baseline migration version recorded once, got %d", versionCount)
	}
}

func createLegacyDBFixture(t *testing.T, dbPath string) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open legacy db failed: %v", err)
	}
	defer db.Close()

	legacySchema := `
		CREATE TABLE groups (
			id INTEGER PRIMARY KEY,
			created_at TEXT,
			updated_at TEXT,
			deleted_at INTEGER,
			name TEXT NOT NULL
		);
		CREATE UNIQUE INDEX idx_name ON groups(deleted_at, name);

		CREATE TABLE feeds (
			id INTEGER PRIMARY KEY,
			created_at TEXT,
			updated_at TEXT,
			deleted_at INTEGER,
			name TEXT NOT NULL,
			link TEXT NOT NULL,
			last_build TEXT,
			failure TEXT DEFAULT '',
			consecutive_failures INTEGER DEFAULT 0,
			suspended INTEGER DEFAULT 0,
			req_proxy TEXT,
			group_id INTEGER
		);
		CREATE UNIQUE INDEX idx_link ON feeds(deleted_at, link);

		CREATE TABLE items (
			id INTEGER PRIMARY KEY,
			created_at TEXT,
			updated_at TEXT,
			deleted_at INTEGER,
			title TEXT,
			guid TEXT,
			link TEXT,
			content TEXT,
			pub_date TEXT,
			unread INTEGER DEFAULT 1,
			bookmark INTEGER DEFAULT 0,
			feed_id INTEGER
		);
		CREATE UNIQUE INDEX idx_guid ON items(deleted_at, guid, feed_id);
	`
	if _, err := db.Exec(legacySchema); err != nil {
		t.Fatalf("create legacy schema failed: %v", err)
	}

	seedData := `
		INSERT INTO groups (id, created_at, updated_at, deleted_at, name) VALUES
			(1, '2024-01-01 00:00:00', '2024-01-02 00:00:00', NULL, 'Legacy Default'),
			(2, '2024-01-03 00:00:00', '2024-01-04 00:00:00', NULL, 'Tech'),
			(3, '2024-01-03 00:00:00', '2024-01-04 00:00:00', 1704500000, 'Archived'),
			(4, '2024-01-03 00:00:00', '2024-01-04 00:00:00', NULL, 'Tech');

		INSERT INTO feeds (id, created_at, updated_at, deleted_at, name, link, last_build, failure, consecutive_failures, suspended, req_proxy, group_id) VALUES
			(1, '2024-02-01 00:00:00', '2024-02-02 00:00:00', NULL, 'Feed A', 'https://legacy.example/feed-a.xml', '2024-02-03 00:00:00', 'timeout', 3, 1, 'http://127.0.0.1:8080', 2),
			(2, '2024-02-01 00:00:00', '2024-02-02 00:00:00', NULL, 'Feed B', 'https://legacy.example/feed-b.xml', NULL, '', 0, 0, NULL, 999),
			(3, '2024-02-01 00:00:00', '2024-02-02 00:00:00', 1707000000, 'Deleted Feed', 'https://legacy.example/feed-deleted.xml', NULL, '', 0, 0, NULL, 2),
			(4, '2024-02-01 00:00:00', '2024-02-02 00:00:00', NULL, 'Feed A Duplicate', 'https://legacy.example/feed-a.xml', NULL, '', 0, 0, NULL, 2),
			(5, '2024-02-01 00:00:00', '2024-02-02 00:00:00', NULL, 'Feed C', 'https://legacy.example/feed-c.xml', NULL, '', 0, 0, NULL, 4);

		INSERT INTO items (id, created_at, updated_at, deleted_at, title, guid, link, content, pub_date, unread, bookmark, feed_id) VALUES
			(1, '2024-03-01 00:00:00', '2024-03-01 00:00:00', NULL, 'Item A1', 'guid-a1', 'https://legacy.example/a1', 'A1 content', '2024-03-01 10:00:00', 1, 1, 1),
			(2, '2024-03-01 00:00:00', '2024-03-01 00:00:00', NULL, 'Item A2', 'guid-a2', 'https://legacy.example/a2', 'A2 content', '2024-03-01 11:00:00', 0, 0, 1),
			(3, '2024-03-01 00:00:00', '2024-03-01 00:00:00', NULL, 'Item A1 Dup', 'guid-a1', 'https://legacy.example/a1-dup', 'A1 dup content', '2024-03-01 12:00:00', 1, 1, 1),
			(4, '2024-03-01 00:00:00', '2024-03-01 00:00:00', NULL, 'Item B1', 'guid-b1', 'https://legacy.example/b1', 'B1 content', '2024-03-01 13:00:00', 1, 1, 2),
			(5, '2024-03-01 00:00:00', '2024-03-01 00:00:00', NULL, 'Item Deleted Feed', 'guid-x1', 'https://legacy.example/x1', 'X1 content', '2024-03-01 14:00:00', 1, 1, 3),
			(6, '2024-03-01 00:00:00', '2024-03-01 00:00:00', 1708000000, 'Item Soft Deleted', 'guid-d1', 'https://legacy.example/d1', 'D1 content', '2024-03-01 15:00:00', 1, 1, 1),
			(7, '2024-03-01 00:00:00', '2024-03-01 00:00:00', NULL, 'Item From Duplicate Feed', 'guid-a3', 'https://legacy.example/a3', 'A3 content', '2024-03-01 16:00:00', 1, 1, 4),
			(8, '2024-03-01 00:00:00', '2024-03-01 00:00:00', NULL, 'Item C1', 'guid-c1', 'https://legacy.example/c1', 'C1 content', '2024-03-01 17:00:00', 1, 0, 5),
			(9, '2024-03-01 00:00:00', '2024-03-01 00:00:00', NULL, 'Item Missing Guid', NULL, 'https://legacy.example/missing-guid', 'Missing guid content', '2024-03-01 18:00:00', 1, 0, 2);
	`
	if _, err := db.Exec(seedData); err != nil {
		t.Fatalf("seed legacy data failed: %v", err)
	}
}
