package store

import (
	"database/sql"
	"errors"
	"slices"
	"testing"

	"github.com/0x2E/fusion/internal/model"
)

func setItemCreatedAt(t *testing.T, store *Store, id, createdAt int64) {
	t.Helper()

	if _, err := store.db.Exec(
		`UPDATE items SET created_at = :created_at WHERE id = :id`,
		sql.Named("created_at", createdAt), sql.Named("id", id),
	); err != nil {
		t.Fatalf("set item created_at: %v", err)
	}
}

func TestGReaderUnreadCounts(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	tech := mustCreateGroup(t, store, "Tech")
	news := mustCreateGroup(t, store, "News")
	feed1 := mustCreateFeed(t, store, tech.ID, "Feed 1", "https://example.com/1", "https://example.com", "")
	feed2 := mustCreateFeed(t, store, news.ID, "Feed 2", "https://example.com/2", "https://example.com", "")

	item1 := mustCreateItem(t, store, feed1.ID, "g1", "One", "https://example.com/1/1", "<p>1</p>", 100)
	item2 := mustCreateItem(t, store, feed1.ID, "g2", "Two", "https://example.com/1/2", "<p>2</p>", 200)
	item3 := mustCreateItem(t, store, feed2.ID, "g3", "Three", "https://example.com/2/3", "<p>3</p>", 300)

	setItemCreatedAt(t, store, item1.ID, 1000)
	setItemCreatedAt(t, store, item2.ID, 2000)
	setItemCreatedAt(t, store, item3.ID, 3000)

	if err := store.BatchUpdateItemsUnread([]int64{item3.ID}, false); err != nil {
		t.Fatalf("mark item read: %v", err)
	}

	counts, err := store.GReaderUnreadCounts()
	if err != nil {
		t.Fatalf("GReaderUnreadCounts() failed: %v", err)
	}

	if counts.Total.Count != 2 || counts.Total.NewestCreatedAt != 2000 {
		t.Fatalf("expected total unread 2 with newest 1000, got %+v", counts.Total)
	}

	// A feed with no unread items still reports its idle slot.
	if got := counts.ByFeed[feed2.ID]; got.Count != 0 || got.NewestCreatedAt != 0 {
		t.Fatalf("expected empty feed summary, got %+v", got)
	}
	if got := counts.ByFeed[feed1.ID]; got.Count != 2 || got.NewestCreatedAt != 2000 {
		t.Fatalf("expected feed1 unread 2 with newest 1000, got %+v", got)
	}
	if got := counts.ByGroup[tech.ID]; got.Count != 2 || got.NewestCreatedAt != 2000 {
		t.Fatalf("expected tech unread 2 with newest 1000, got %+v", got)
	}
	if got := counts.ByGroup[news.ID]; got.Count != 0 {
		t.Fatalf("expected news group to report zero unread, got %+v", got)
	}
}

func TestGetGroupByName(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	tech := mustCreateGroup(t, store, "Tech")

	group, err := store.GetGroupByName("Tech")
	if err != nil {
		t.Fatalf("GetGroupByName() failed: %v", err)
	}
	if group.ID != tech.ID {
		t.Fatalf("expected group %d, got %d", tech.ID, group.ID)
	}

	if _, err := store.GetGroupByName("Missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing group, got %v", err)
	}
}

func TestGetFeedByLink(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Tech")
	created := mustCreateFeed(t, store, group.ID, "Feed", "https://example.com/rss.xml", "https://example.com", "")

	feed, err := store.GetFeedByLink("https://example.com/rss.xml")
	if err != nil {
		t.Fatalf("GetFeedByLink() failed: %v", err)
	}
	if feed.ID != created.ID {
		t.Fatalf("expected feed %d, got %d", created.ID, feed.ID)
	}

	if _, err := store.GetFeedByLink("https://example.com/missing.xml"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for unknown link, got %v", err)
	}
}

func greaderFixture(t *testing.T, store *Store) (tech, news *model.Group, feed1, feed2 *model.Feed) {
	t.Helper()

	tech = mustCreateGroup(t, store, "Tech")
	news = mustCreateGroup(t, store, "News")
	feed1 = mustCreateFeed(t, store, tech.ID, "Feed 1", "https://example.com/1", "https://example.com/1", "")
	feed2 = mustCreateFeed(t, store, news.ID, "Feed 2", "https://example.com/2", "https://example.com/2", "")
	return tech, news, feed1, feed2
}

func TestListGReaderItemsStreamFilters(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	tech, _, feed1, feed2 := greaderFixture(t, store)

	item1 := mustCreateItem(t, store, feed1.ID, "g1", "One", "https://example.com/1/1", "<p>1</p>", 100)
	item2 := mustCreateItem(t, store, feed1.ID, "g2", "Two", "https://example.com/1/2", "<p>2</p>", 200)
	item3 := mustCreateItem(t, store, feed2.ID, "g3", "Three", "https://example.com/2/3", "<p>3</p>", 300)
	setItemCreatedAt(t, store, item1.ID, 1000)
	setItemCreatedAt(t, store, item2.ID, 2000)
	setItemCreatedAt(t, store, item3.ID, 3000)

	if err := store.BatchUpdateItemsUnread([]int64{item2.ID}, false); err != nil {
		t.Fatalf("mark item2 read: %v", err)
	}
	if _, err := store.CreateBookmark(&item3.ID, &item3.FeedID, item3.Link, item3.Title, item3.Content, item3.PubDate, feed2.Name); err != nil {
		t.Fatalf("bookmark item3: %v", err)
	}

	listIDs := func(t *testing.T, params GReaderItemsParams) []int64 {
		t.Helper()
		items, err := store.ListGReaderItems(params)
		if err != nil {
			t.Fatalf("ListGReaderItems() failed: %v", err)
		}
		ids := make([]int64, len(items))
		for i, item := range items {
			ids[i] = item.ID
		}
		return ids
	}

	t.Run("feed stream", func(t *testing.T) {
		got := listIDs(t, GReaderItemsParams{FeedID: &feed1.ID})
		if !slices.Equal(got, []int64{item2.ID, item1.ID}) {
			t.Fatalf("unexpected ids %v", got)
		}
	})

	t.Run("label stream", func(t *testing.T) {
		got := listIDs(t, GReaderItemsParams{GroupID: &tech.ID})
		if !slices.Equal(got, []int64{item2.ID, item1.ID}) {
			t.Fatalf("unexpected ids %v", got)
		}
	})

	t.Run("unread only", func(t *testing.T) {
		got := listIDs(t, GReaderItemsParams{UnreadOnly: true})
		if !slices.Equal(got, []int64{item3.ID, item1.ID}) {
			t.Fatalf("unexpected ids %v", got)
		}
	})

	t.Run("read only", func(t *testing.T) {
		got := listIDs(t, GReaderItemsParams{ReadOnly: true})
		if !slices.Equal(got, []int64{item2.ID}) {
			t.Fatalf("unexpected ids %v", got)
		}
	})

	t.Run("read and unread filters intersect to nothing", func(t *testing.T) {
		got := listIDs(t, GReaderItemsParams{ReadOnly: true, UnreadOnly: true})
		if len(got) != 0 {
			t.Fatalf("expected no ids, got %v", got)
		}
	})

	t.Run("starred only", func(t *testing.T) {
		got := listIDs(t, GReaderItemsParams{Starred: true})
		if !slices.Equal(got, []int64{item3.ID}) {
			t.Fatalf("unexpected ids %v", got)
		}
	})

	t.Run("not starred", func(t *testing.T) {
		got := listIDs(t, GReaderItemsParams{NotStarred: true})
		if !slices.Equal(got, []int64{item2.ID, item1.ID}) {
			t.Fatalf("unexpected ids %v", got)
		}
	})

	t.Run("starred and not starred intersect to nothing", func(t *testing.T) {
		got := listIDs(t, GReaderItemsParams{Starred: true, NotStarred: true})
		if len(got) != 0 {
			t.Fatalf("expected no ids, got %v", got)
		}
	})

	t.Run("created_at bounds are exclusive", func(t *testing.T) {
		before, after := int64(3000), int64(1000)
		got := listIDs(t, GReaderItemsParams{CreatedBefore: &before, CreatedAfter: &after})
		if !slices.Equal(got, []int64{item2.ID}) {
			t.Fatalf("unexpected ids %v", got)
		}
	})

	t.Run("items carry feed and group metadata", func(t *testing.T) {
		items, err := store.ListGReaderItems(GReaderItemsParams{Starred: true})
		if err != nil {
			t.Fatalf("ListGReaderItems() failed: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected one starred item, got %d", len(items))
		}
		item := items[0]
		if item.FeedName != "Feed 2" || item.GroupName != "News" || item.FeedSiteURL != "https://example.com/2" {
			t.Fatalf("unexpected metadata %+v", item)
		}
		if !item.Starred {
			t.Fatalf("expected starred state, got %+v", item)
		}
	})
}

func TestListGReaderItemsCursorAndOrder(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	_, _, feed1, _ := greaderFixture(t, store)

	first := mustCreateItem(t, store, feed1.ID, "g1", "One", "https://example.com/1/1", "<p>1</p>", 100)
	second := mustCreateItem(t, store, feed1.ID, "g2", "Two", "https://example.com/1/2", "<p>2</p>", 200)
	third := mustCreateItem(t, store, feed1.ID, "g3", "Three", "https://example.com/1/3", "<p>3</p>", 300)

	ids := func(items []*GReaderItem) []int64 {
		result := make([]int64, len(items))
		for i, item := range items {
			result[i] = item.ID
		}
		return result
	}

	t.Run("newest first by default", func(t *testing.T) {
		items, err := store.ListGReaderItems(GReaderItemsParams{})
		if err != nil {
			t.Fatalf("ListGReaderItems() failed: %v", err)
		}
		if !slices.Equal(ids(items), []int64{third.ID, second.ID, first.ID}) {
			t.Fatalf("unexpected order %v", ids(items))
		}
	})

	t.Run("oldest first when ascending", func(t *testing.T) {
		items, err := store.ListGReaderItems(GReaderItemsParams{Ascending: true})
		if err != nil {
			t.Fatalf("ListGReaderItems() failed: %v", err)
		}
		if !slices.Equal(ids(items), []int64{first.ID, second.ID, third.ID}) {
			t.Fatalf("unexpected order %v", ids(items))
		}
	})

	t.Run("descending continuation", func(t *testing.T) {
		items, err := store.ListGReaderItems(GReaderItemsParams{Limit: 2})
		if err != nil {
			t.Fatalf("ListGReaderItems() failed: %v", err)
		}
		last := items[len(items)-1].ID
		next, err := store.ListGReaderItems(GReaderItemsParams{BeforeID: &last})
		if err != nil {
			t.Fatalf("ListGReaderItems() failed: %v", err)
		}
		if !slices.Equal(ids(next), []int64{first.ID}) {
			t.Fatalf("unexpected continuation page %v", ids(next))
		}
	})

	t.Run("ascending continuation", func(t *testing.T) {
		items, err := store.ListGReaderItems(GReaderItemsParams{Ascending: true, Limit: 1})
		if err != nil {
			t.Fatalf("ListGReaderItems() failed: %v", err)
		}
		last := items[len(items)-1].ID
		next, err := store.ListGReaderItems(GReaderItemsParams{Ascending: true, AfterID: &last})
		if err != nil {
			t.Fatalf("ListGReaderItems() failed: %v", err)
		}
		if !slices.Equal(ids(next), []int64{second.ID, third.ID}) {
			t.Fatalf("unexpected continuation page %v", ids(next))
		}
	})
}

func TestListGReaderItemRefs(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	_, _, feed1, _ := greaderFixture(t, store)

	first := mustCreateItem(t, store, feed1.ID, "g1", "One", "https://example.com/1/1", "<p>1</p>", 100)
	second := mustCreateItem(t, store, feed1.ID, "g2", "Two", "https://example.com/1/2", "<p>2</p>", 200)
	setItemCreatedAt(t, store, first.ID, 1000)
	setItemCreatedAt(t, store, second.ID, 2000)

	refs, err := store.ListGReaderItemRefs(GReaderItemsParams{FeedID: &feed1.ID})
	if err != nil {
		t.Fatalf("ListGReaderItemRefs() failed: %v", err)
	}
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(refs))
	}
	if refs[0].ID != second.ID || refs[0].CreatedAt != 2000 {
		t.Fatalf("unexpected first ref %+v", refs[0])
	}
}

func TestListGReaderItemsByIDs(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	_, _, feed1, _ := greaderFixture(t, store)

	first := mustCreateItem(t, store, feed1.ID, "g1", "One", "https://example.com/1/1", "<p>1</p>", 100)
	second := mustCreateItem(t, store, feed1.ID, "g2", "Two", "https://example.com/1/2", "<p>2</p>", 200)

	// Unknown ids are skipped; order is the caller's job because handlers
	// rebuild it from the request.
	items, err := store.ListGReaderItemsByIDs([]int64{second.ID, 9999, first.ID})
	if err != nil {
		t.Fatalf("ListGReaderItemsByIDs() failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	found := map[int64]bool{}
	for _, item := range items {
		found[item.ID] = true
	}
	if !found[first.ID] || !found[second.ID] {
		t.Fatalf("expected both items, got %v", found)
	}

	empty, err := store.ListGReaderItemsByIDs(nil)
	if err != nil {
		t.Fatalf("ListGReaderItemsByIDs() failed: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected no items, got %d", len(empty))
	}
}

// mark-all-as-read uses crawl time, so an item with an old pub_date but a
// recent created_at must stay unread when ts sits between the two axes.
func TestMarkGReaderItemsReadUsesCreatedAt(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	tech, _, feed1, feed2 := greaderFixture(t, store)

	oldPubDate := mustCreateItem(t, store, feed1.ID, "g1", "Old", "https://example.com/1/1", "<p>1</p>", 100)
	newPubDate := mustCreateItem(t, store, feed1.ID, "g2", "New", "https://example.com/1/2", "<p>2</p>", 20000)
	otherFeed := mustCreateItem(t, store, feed2.ID, "g3", "Other", "https://example.com/2/3", "<p>3</p>", 300)
	setItemCreatedAt(t, store, oldPubDate.ID, 1000)
	setItemCreatedAt(t, store, newPubDate.ID, 3000)
	setItemCreatedAt(t, store, otherFeed.ID, 2000)

	if err := store.MarkGReaderItemsRead(MarkGReaderItemsReadParams{FeedID: &feed1.ID, CreatedAt: 2000}); err != nil {
		t.Fatalf("MarkGReaderItemsRead() failed: %v", err)
	}

	checkUnread := func(id int64, wantUnread bool) {
		t.Helper()
		item, err := store.GetItem(id)
		if err != nil {
			t.Fatalf("GetItem(%d) failed: %v", id, err)
		}
		if item.Unread != wantUnread {
			t.Fatalf("item %d: expected unread=%v, got %v", id, wantUnread, item.Unread)
		}
	}
	checkUnread(oldPubDate.ID, false)
	// A future pub_date must not keep it unread: the cutoff is crawl time.
	checkUnread(newPubDate.ID, true)
	checkUnread(otherFeed.ID, true)

	if err := store.MarkGReaderItemsRead(MarkGReaderItemsReadParams{GroupID: &tech.ID, CreatedAt: 5000}); err != nil {
		t.Fatalf("MarkGReaderItemsRead() failed: %v", err)
	}
	checkUnread(newPubDate.ID, false)
	checkUnread(otherFeed.ID, true)
}
