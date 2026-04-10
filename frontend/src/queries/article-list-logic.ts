export function countFetchedRows<T extends { fetchedCount: number }>(
  pages: T[],
): number {
  return pages.reduce((count, page) => count + page.fetchedCount, 0);
}

export function getAdjustedOffset(
  loadedCount: number,
  removedCount: number,
): number {
  return Math.max(0, loadedCount - removedCount);
}

export function countStickyRemovedRows(
  stickyVisibleIds: Record<number, true>,
  membershipOverrides: Record<number, boolean>,
): number {
  return Object.keys(stickyVisibleIds).filter(
    (id) => membershipOverrides[Number(id)] === false,
  ).length;
}

export function mergeStickyArticles<T>(args: {
  baselineArticles: T[];
  stickyVisibleIds: Record<number, true>;
  stickyArticles: Record<number, T>;
  getArticleId: (article: T) => number;
  getCachedArticle?: (id: number) => T | undefined;
}): T[] {
  const articleById = new Map<number, T>();

  for (const article of args.baselineArticles) {
    articleById.set(args.getArticleId(article), article);
  }

  for (const stickyId of Object.keys(args.stickyVisibleIds)) {
    const articleId = Number(stickyId);
    if (articleById.has(articleId)) {
      continue;
    }

    const article =
      args.stickyArticles[articleId] ?? args.getCachedArticle?.(articleId);
    if (article) {
      articleById.set(articleId, article);
    }
  }

  return Array.from(articleById.values());
}
