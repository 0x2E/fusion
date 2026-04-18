import test from "node:test";
import assert from "node:assert/strict";
import type { Bookmark } from "../src/lib/api/types.ts";
import {
  buildSavedBookmarkLookup,
  resolveBookmarkFeedId,
  resolveSavedBookmarkId,
  removeSavedBookmarkRef,
  upsertSavedBookmarkRef,
} from "../src/queries/bookmark-lookup-logic.ts";

function makeBookmark(overrides: Partial<Bookmark>): Bookmark {
  return {
    id: 1,
    item_id: 42,
    link: "https://example.com/article",
    title: "Saved article",
    content: "content",
    pub_date: 1,
    feed_name: "Feed",
    created_at: 1,
    ...overrides,
  };
}

test("buildSavedBookmarkLookup returns delete targets without loaded bookmark pages", () => {
  const lookup = buildSavedBookmarkLookup([
    { bookmark_id: 7, item_id: 42 },
    { bookmark_id: 9, item_id: 99 },
  ]);

  assert.equal(lookup.get(42), 7);
  assert.equal(lookup.get(99), 9);
});

test("upsertSavedBookmarkRef replaces the existing saved entry for an item", () => {
  const updated = upsertSavedBookmarkRef(
    [
      { bookmark_id: 1, item_id: 42 },
      { bookmark_id: 2, item_id: 99 },
    ],
    makeBookmark({ id: 11, item_id: 42 }),
  );

  assert.deepEqual(updated, [
    { bookmark_id: 11, item_id: 42 },
    { bookmark_id: 2, item_id: 99 },
  ]);
});

test("removeSavedBookmarkRef removes only the deleted bookmark", () => {
  const updated = removeSavedBookmarkRef(
    [
      { bookmark_id: 1, item_id: 42 },
      { bookmark_id: 2, item_id: 99 },
    ],
    2,
  );

  assert.deepEqual(updated, [{ bookmark_id: 1, item_id: 42 }]);
});

test("resolveSavedBookmarkId falls back to loaded orphan bookmarks", () => {
  const savedLookup = buildSavedBookmarkLookup([]);
  const loadedBookmarks = new Map([[-7, { id: 7 }]]);

  assert.equal(
    resolveSavedBookmarkId({
      savedLookup,
      loadedBookmarks,
      itemId: -7,
    }),
    7,
  );
});

test("resolveBookmarkFeedId prefers cached item feed ids over ambiguous feed names", () => {
  const feedId = resolveBookmarkFeedId({
    bookmark: makeBookmark({ item_id: 42, feed_name: "Shared Feed" }),
    uniqueFeedIdByName: new Map(),
    getCachedFeedId: (itemId) => (itemId === 42 ? 99 : undefined),
  });

  assert.equal(feedId, 99);
});
