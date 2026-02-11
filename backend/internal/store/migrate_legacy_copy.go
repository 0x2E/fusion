package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type legacyGroupRow struct {
	ID         int64
	Name       string
	CreatedAt  int64
	UpdatedAt  int64
	HasUpdated bool
}

type legacyFeedRow struct {
	ID        int64
	GroupID   int64
	Name      string
	Link      string
	LastBuild int64
	Failure   string
	Failures  int64
	Suspended int64
	Proxy     string
	CreatedAt int64
	UpdatedAt int64
}

type legacyItemRow struct {
	ID        int64
	FeedID    int64
	GUID      string
	Title     string
	Link      string
	Content   string
	PubDate   int64
	Unread    int64
	CreatedAt int64
}

type itemDedupKey struct {
	FeedID int64
	GUID   string
}

func copyLegacyGroups(tx *sql.Tx) (map[int64]int64, error) {
	nowTs := time.Now().Unix()
	rows, err := tx.Query(`
		SELECT id, name, created_at, updated_at
		FROM legacy.groups
		WHERE (deleted_at IS NULL OR CAST(deleted_at AS INTEGER) = 0)
		  AND name IS NOT NULL
		  AND TRIM(name) <> ''
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("query legacy groups: %w", err)
	}
	defer rows.Close()

	legacyRows := make([]legacyGroupRow, 0)
	for rows.Next() {
		var (
			id         int64
			nameRaw    any
			createdRaw any
			updatedRaw any
		)
		if err := rows.Scan(&id, &nameRaw, &createdRaw, &updatedRaw); err != nil {
			return nil, fmt.Errorf("scan legacy group row: %w", err)
		}

		name := legacyString(nameRaw, "")
		if strings.TrimSpace(name) == "" {
			continue
		}

		createdAt := legacyUnixSeconds(createdRaw, nowTs)
		updatedAt, hasUpdated := legacyUnixSecondsNullable(updatedRaw)
		legacyRows = append(legacyRows, legacyGroupRow{
			ID:         id,
			Name:       name,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
			HasUpdated: hasUpdated,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate legacy group rows: %w", err)
	}

	canonicalByName := make(map[string]int64, len(legacyRows))
	for _, row := range legacyRows {
		current, exists := canonicalByName[row.Name]
		if !exists || row.ID < current {
			canonicalByName[row.Name] = row.ID
		}
	}

	groupMap := make(map[int64]int64, len(legacyRows))
	for _, row := range legacyRows {
		groupMap[row.ID] = canonicalByName[row.Name]
	}

	insertStmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO groups (id, name, created_at, updated_at)
		VALUES (:id, :name, :created_at, :updated_at)
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare insert migrated groups: %w", err)
	}
	defer insertStmt.Close()

	for _, row := range legacyRows {
		if canonicalByName[row.Name] != row.ID {
			continue
		}
		updatedAt := row.CreatedAt
		if row.HasUpdated {
			updatedAt = row.UpdatedAt
		}
		if _, err := insertStmt.Exec(
			sql.Named("id", row.ID),
			sql.Named("name", row.Name),
			sql.Named("created_at", row.CreatedAt),
			sql.Named("updated_at", updatedAt),
		); err != nil {
			return nil, fmt.Errorf("insert migrated group %d: %w", row.ID, err)
		}
	}

	return groupMap, nil
}

func restoreDefaultGroupFromLegacy(tx *sql.Tx) error {
	nowTs := time.Now().Unix()

	var (
		nameRaw    any
		createdRaw any
		updatedRaw any
	)
	err := tx.QueryRow(`
		SELECT name, created_at, updated_at
		FROM legacy.groups
		WHERE id = 1
		  AND (deleted_at IS NULL OR CAST(deleted_at AS INTEGER) = 0)
		  AND name IS NOT NULL
		  AND TRIM(name) <> ''
		LIMIT 1
	`).Scan(&nameRaw, &createdRaw, &updatedRaw)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("query default legacy group: %w", err)
	}

	name := legacyString(nameRaw, "")
	if strings.TrimSpace(name) == "" {
		return nil
	}

	createdAt := legacyUnixSeconds(createdRaw, nowTs)
	updatedAt, hasUpdated := legacyUnixSecondsNullable(updatedRaw)
	if !hasUpdated {
		updatedAt = createdAt
	}

	if _, err := tx.Exec(
		`INSERT OR REPLACE INTO groups (id, name, created_at, updated_at) VALUES (1, :name, :created_at, :updated_at)`,
		sql.Named("name", name),
		sql.Named("created_at", createdAt),
		sql.Named("updated_at", updatedAt),
	); err != nil {
		return fmt.Errorf("restore default group: %w", err)
	}

	return nil
}

func copyLegacyFeeds(tx *sql.Tx, groupMap map[int64]int64) (map[int64]int64, error) {
	nowTs := time.Now().Unix()
	rows, err := tx.Query(`
		SELECT
			id,
			group_id,
			name,
			link,
			last_build,
			failure,
			consecutive_failures,
			suspended,
			req_proxy,
			created_at,
			updated_at
		FROM legacy.feeds
		WHERE (deleted_at IS NULL OR CAST(deleted_at AS INTEGER) = 0)
		  AND link IS NOT NULL
		  AND TRIM(link) <> ''
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("query legacy feeds: %w", err)
	}
	defer rows.Close()

	legacyRows := make([]legacyFeedRow, 0)
	for rows.Next() {
		var (
			idRaw        any
			groupIDRaw   any
			nameRaw      any
			linkRaw      any
			lastBuildRaw any
			failureRaw   any
			failuresRaw  any
			suspendedRaw any
			proxyRaw     any
			createdAtRaw any
			updatedAtRaw any
		)
		if err := rows.Scan(
			&idRaw,
			&groupIDRaw,
			&nameRaw,
			&linkRaw,
			&lastBuildRaw,
			&failureRaw,
			&failuresRaw,
			&suspendedRaw,
			&proxyRaw,
			&createdAtRaw,
			&updatedAtRaw,
		); err != nil {
			return nil, fmt.Errorf("scan legacy feed row: %w", err)
		}

		id := legacyInt64(idRaw, 0)
		link := legacyString(linkRaw, "")
		if id == 0 || strings.TrimSpace(link) == "" {
			continue
		}

		name := legacyString(nameRaw, "")
		if strings.TrimSpace(name) == "" {
			name = link
		}

		legacyGroupID := legacyInt64(groupIDRaw, 0)
		mappedGroupID := int64(1)
		if canonicalGroupID, ok := groupMap[legacyGroupID]; ok {
			mappedGroupID = canonicalGroupID
		}

		createdAt := legacyUnixSeconds(createdAtRaw, nowTs)
		updatedAt, hasUpdated := legacyUnixSecondsNullable(updatedAtRaw)
		if !hasUpdated {
			updatedAt = createdAt
		}

		legacyRows = append(legacyRows, legacyFeedRow{
			ID:        id,
			GroupID:   mappedGroupID,
			Name:      name,
			Link:      link,
			LastBuild: legacyUnixSeconds(lastBuildRaw, 0),
			Failure:   legacyString(failureRaw, ""),
			Failures:  legacyInt64(failuresRaw, 0),
			Suspended: boolToInt64(legacyInt64(suspendedRaw, 0) != 0),
			Proxy:     legacyString(proxyRaw, ""),
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate legacy feed rows: %w", err)
	}

	canonicalByLink := make(map[string]int64, len(legacyRows))
	for _, row := range legacyRows {
		if _, exists := canonicalByLink[row.Link]; !exists {
			canonicalByLink[row.Link] = row.ID
		}
	}

	feedMap := make(map[int64]int64, len(legacyRows))
	for _, row := range legacyRows {
		feedMap[row.ID] = canonicalByLink[row.Link]
	}

	insertStmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO feeds (
			id,
			group_id,
			name,
			link,
			site_url,
			last_build,
			last_failure_at,
			failure,
			failures,
			suspended,
			proxy,
			created_at,
			updated_at
		)
		VALUES (
			:id,
			:group_id,
			:name,
			:link,
			:site_url,
			:last_build,
			:last_failure_at,
			:failure,
			:failures,
			:suspended,
			:proxy,
			:created_at,
			:updated_at
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare insert migrated feeds: %w", err)
	}
	defer insertStmt.Close()

	for _, row := range legacyRows {
		if canonicalByLink[row.Link] != row.ID {
			continue
		}
		if _, err := insertStmt.Exec(
			sql.Named("id", row.ID),
			sql.Named("group_id", row.GroupID),
			sql.Named("name", row.Name),
			sql.Named("link", row.Link),
			sql.Named("site_url", ""),
			sql.Named("last_build", row.LastBuild),
			sql.Named("last_failure_at", int64(0)),
			sql.Named("failure", row.Failure),
			sql.Named("failures", row.Failures),
			sql.Named("suspended", row.Suspended),
			sql.Named("proxy", row.Proxy),
			sql.Named("created_at", row.CreatedAt),
			sql.Named("updated_at", row.UpdatedAt),
		); err != nil {
			return nil, fmt.Errorf("insert migrated feed %d: %w", row.ID, err)
		}
	}

	return feedMap, nil
}

func copyLegacyItems(tx *sql.Tx) error {
	if err := prepareLegacyItemMapTx(tx); err != nil {
		return err
	}

	nowTs := time.Now().Unix()
	rows, err := tx.Query(`
		SELECT
			i.id,
			fm.new_id AS mapped_feed_id,
			CASE
				WHEN i.guid IS NULL OR TRIM(i.guid) = '' THEN printf('legacy-item-%d', i.id)
				ELSE i.guid
			END AS mapped_guid,
			i.title,
			i.link,
			i.content,
			i.pub_date,
			i.unread,
			i.created_at
		FROM legacy.items i
		INNER JOIN legacy_feed_map fm ON fm.old_id = i.feed_id
		WHERE (i.deleted_at IS NULL OR CAST(i.deleted_at AS INTEGER) = 0)
		ORDER BY mapped_feed_id, mapped_guid, i.id
	`)
	if err != nil {
		return fmt.Errorf("query legacy items: %w", err)
	}
	defer rows.Close()

	insertItemStmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO items (
			id,
			feed_id,
			guid,
			title,
			link,
			content,
			pub_date,
			unread,
			created_at
		)
		VALUES (
			:id,
			:feed_id,
			:guid,
			:title,
			:link,
			:content,
			:pub_date,
			:unread,
			:created_at
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert migrated items: %w", err)
	}
	defer insertItemStmt.Close()

	insertItemMapStmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO legacy_item_map (old_id, new_id)
		VALUES (:old_id, :new_id)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert migrated item map: %w", err)
	}
	defer insertItemMapStmt.Close()

	flushCanonical := func(row legacyItemRow) error {
		_, err := insertItemStmt.Exec(
			sql.Named("id", row.ID),
			sql.Named("feed_id", row.FeedID),
			sql.Named("guid", row.GUID),
			sql.Named("title", row.Title),
			sql.Named("link", row.Link),
			sql.Named("content", row.Content),
			sql.Named("pub_date", row.PubDate),
			sql.Named("unread", row.Unread),
			sql.Named("created_at", row.CreatedAt),
		)
		if err != nil {
			return fmt.Errorf("insert migrated item %d: %w", row.ID, err)
		}
		return nil
	}

	var (
		hasCanonical bool
		canonicalKey itemDedupKey
		canonicalRow legacyItemRow
	)

	for rows.Next() {
		var (
			idRaw         any
			mappedFeedRaw any
			guidRaw       any
			titleRaw      any
			linkRaw       any
			contentRaw    any
			pubDateRaw    any
			unreadRaw     any
			createdAtRaw  any
		)
		if err := rows.Scan(
			&idRaw,
			&mappedFeedRaw,
			&guidRaw,
			&titleRaw,
			&linkRaw,
			&contentRaw,
			&pubDateRaw,
			&unreadRaw,
			&createdAtRaw,
		); err != nil {
			return fmt.Errorf("scan legacy item row: %w", err)
		}

		row := legacyItemRow{
			ID:        legacyInt64(idRaw, 0),
			FeedID:    legacyInt64(mappedFeedRaw, 0),
			GUID:      legacyString(guidRaw, ""),
			Title:     legacyString(titleRaw, ""),
			Link:      legacyString(linkRaw, ""),
			Content:   legacyString(contentRaw, ""),
			PubDate:   legacyUnixSeconds(pubDateRaw, 0),
			Unread:    boolToInt64(legacyInt64(unreadRaw, 1) != 0),
			CreatedAt: legacyUnixSeconds(createdAtRaw, nowTs),
		}

		if row.ID == 0 || row.FeedID == 0 || strings.TrimSpace(row.GUID) == "" {
			continue
		}

		key := itemDedupKey{FeedID: row.FeedID, GUID: row.GUID}
		if !hasCanonical || key != canonicalKey {
			if hasCanonical {
				if err := flushCanonical(canonicalRow); err != nil {
					return err
				}
			}
			hasCanonical = true
			canonicalKey = key
			canonicalRow = row
		}

		if _, err := insertItemMapStmt.Exec(
			sql.Named("old_id", row.ID),
			sql.Named("new_id", canonicalRow.ID),
		); err != nil {
			return fmt.Errorf("insert migrated item mapping %d->%d: %w", row.ID, canonicalRow.ID, err)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate legacy item rows: %w", err)
	}

	if hasCanonical {
		if err := flushCanonical(canonicalRow); err != nil {
			return err
		}
	}

	return nil
}

func copyLegacyBookmarks(tx *sql.Tx) error {
	rows, err := tx.Query(`
		SELECT
			o.id AS old_item_id,
			n.id AS item_id,
			n.link,
			n.title,
			n.content,
			n.pub_date,
			COALESCE(f.name, '') AS feed_name,
			n.created_at
		FROM legacy.items o
		INNER JOIN legacy_item_map m ON m.old_id = o.id
		INNER JOIN items n ON n.id = m.new_id
		LEFT JOIN feeds f ON f.id = n.feed_id
		WHERE (o.deleted_at IS NULL OR CAST(o.deleted_at AS INTEGER) = 0)
		  AND COALESCE(CAST(o.bookmark AS INTEGER), 0) <> 0
		  AND n.link IS NOT NULL
		  AND TRIM(n.link) <> ''
		ORDER BY n.link, o.id
	`)
	if err != nil {
		return fmt.Errorf("query migrated bookmarks source rows: %w", err)
	}
	defer rows.Close()

	insertStmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO bookmarks (item_id, link, title, content, pub_date, feed_name, created_at)
		VALUES (:item_id, :link, :title, :content, :pub_date, :feed_name, :created_at)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert bookmarks: %w", err)
	}
	defer insertStmt.Close()

	lastInsertedLink := ""
	for rows.Next() {
		var (
			oldItemID int64
			itemID    int64
			link      string
			title     string
			content   string
			pubDate   int64
			feedName  string
			createdAt int64
		)
		if err := rows.Scan(
			&oldItemID,
			&itemID,
			&link,
			&title,
			&content,
			&pubDate,
			&feedName,
			&createdAt,
		); err != nil {
			return fmt.Errorf("scan migrated bookmarks source row: %w", err)
		}

		if link == lastInsertedLink {
			continue
		}

		if _, err := insertStmt.Exec(
			sql.Named("item_id", itemID),
			sql.Named("link", link),
			sql.Named("title", title),
			sql.Named("content", content),
			sql.Named("pub_date", pubDate),
			sql.Named("feed_name", feedName),
			sql.Named("created_at", createdAt),
		); err != nil {
			return fmt.Errorf("insert migrated bookmark for legacy item %d: %w", oldItemID, err)
		}

		lastInsertedLink = link
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate migrated bookmarks source rows: %w", err)
	}

	return nil
}
