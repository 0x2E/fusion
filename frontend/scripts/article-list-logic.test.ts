import test from "node:test";
import assert from "node:assert/strict";
import {
  countStickyRemovedRows,
  countFetchedRows,
  getAdjustedOffset,
  mergeStickyArticles,
} from "../src/queries/article-list-logic.ts";

test("countFetchedRows ignores optimistic bookmark inserts", () => {
  const fetched = countFetchedRows([
    {
      fetchedCount: 100,
      data: Array.from({ length: 101 }, (_, id) => ({ id })),
    },
  ]);

  assert.equal(fetched, 100);
  assert.equal(getAdjustedOffset(fetched, 0), 100);
});

test("mergeStickyArticles keeps a sticky snapshot without cached detail", () => {
  const articles = mergeStickyArticles({
    baselineArticles: [],
    stickyVisibleIds: { 42: true },
    stickyArticles: {
      42: {
        id: 42,
        title: "Saved snapshot",
      },
    },
    getArticleId: (article) => article.id,
    getCachedArticle: () => undefined,
  });

  assert.deepEqual(articles, [
    {
      id: 42,
      title: "Saved snapshot",
    },
  ]);
});

test("countStickyRemovedRows keeps removed starred rows in offset math", () => {
  const removed = countStickyRemovedRows(
    { 10: true, 11: true },
    { 10: false, 11: true, 12: false },
  );

  assert.equal(removed, 1);
  assert.equal(getAdjustedOffset(100, removed), 99);
});
