import type { Bookmark, SavedBookmarkRef } from "@/lib/api";

export function buildSavedBookmarkLookup(
  savedRefs: SavedBookmarkRef[],
): Map<number, number> {
  return new Map(savedRefs.map((ref) => [ref.item_id, ref.bookmark_id]));
}

export function resolveSavedBookmarkId(args: {
  savedLookup: Map<number, number>;
  loadedBookmarks: Map<number, Pick<Bookmark, "id">>;
  itemId: number;
}): number | undefined {
  return args.savedLookup.get(args.itemId) ?? args.loadedBookmarks.get(args.itemId)?.id;
}

export function resolveBookmarkFeedId(args: {
  bookmark: Pick<Bookmark, "item_id" | "feed_name">;
  uniqueFeedIdByName: Map<string, number>;
  getCachedFeedId?: (itemId: number) => number | undefined;
}): number {
  if (args.bookmark.item_id && args.bookmark.item_id > 0) {
    const cachedFeedId = args.getCachedFeedId?.(args.bookmark.item_id);
    if (cachedFeedId !== undefined) {
      return cachedFeedId;
    }
  }

  return args.uniqueFeedIdByName.get(args.bookmark.feed_name) ?? 0;
}

export function upsertSavedBookmarkRef(
  savedRefs: SavedBookmarkRef[] | undefined,
  bookmark: Pick<Bookmark, "id" | "item_id">,
): SavedBookmarkRef[] | undefined {
  if (bookmark.item_id === null || bookmark.item_id <= 0) {
    return savedRefs;
  }

  const nextRef: SavedBookmarkRef = {
    bookmark_id: bookmark.id,
    item_id: bookmark.item_id,
  };

  if (!savedRefs) {
    return [nextRef];
  }

  const index = savedRefs.findIndex((ref) => ref.item_id === bookmark.item_id);
  if (index === -1) {
    return [nextRef, ...savedRefs];
  }

  const next = [...savedRefs];
  next[index] = nextRef;
  return next;
}

export function removeSavedBookmarkRef(
  savedRefs: SavedBookmarkRef[] | undefined,
  bookmarkId: number,
): SavedBookmarkRef[] | undefined {
  return savedRefs?.filter((ref) => ref.bookmark_id !== bookmarkId);
}
