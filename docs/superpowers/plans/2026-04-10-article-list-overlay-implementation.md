# Article List Overlay Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix unread pagination and starred pagination while preserving the current undo-friendly behavior that keeps modified rows visible until explicit refresh or filter change.

**Architecture:** Keep React Query as the baseline source of paged server data, then layer a small context-aware session overlay on top for local read/star overrides and sticky visibility. Convert bookmarks to infinite pagination and route both `ArticleList` and `ArticleDrawer` through one derived article-list hook so visible rows, navigation, and counters stay consistent.

**Tech Stack:** React 19, TypeScript, TanStack Query, TanStack Router, Zustand, Vite, Go backend verification

---

### Task 1: Bring the Approved Design Into the Branch

**Files:**
- Create: `docs/superpowers/specs/2026-04-10-article-list-overlay-design.md`
- Create: `docs/superpowers/plans/2026-04-10-article-list-overlay-implementation.md`
- Modify: `.gitignore`

- [ ] **Step 1: Copy the approved spec into the feature branch**

```md
# Article List Overlay and Pagination Design

Use the approved spec from the main workspace unchanged so the branch contains the same source of truth that implementation follows.
```

- [ ] **Step 2: Keep local worktrees ignored in the branch**

```gitignore
.claude
.worktrees/
tmp/
build/
```

- [ ] **Step 3: Verify the branch contains the spec and ignore rule**

Run: `git status --short`
Expected: shows the copied spec, the new plan, and the `.gitignore` update in this branch


### Task 2: Replace the Flat Article Session Store With a Context-Aware Overlay

**Files:**
- Create: `frontend/src/lib/article-list-context.ts`
- Modify: `frontend/src/store/article-session.ts`
- Modify: `frontend/src/hooks/use-url-state.ts`

- [ ] **Step 1: Add a reusable list-context key helper**

```ts
import type { ArticleFilter } from "@/lib/article-filter";

export interface ArticleListContext {
  filter: ArticleFilter;
  feedId: number | null;
  groupId: number | null;
}

export function getArticleListContextKey(context: ArticleListContext): string {
  return `${context.filter}:${context.feedId ?? "all"}:${context.groupId ?? "all"}`;
}
```

- [ ] **Step 2: Expand the Zustand store from one global override map to per-context overlays**

```ts
interface ArticleListOverlay {
  readOverrides: Record<number, boolean>;
  starOverrides: Record<number, boolean>;
  stickyVisibleIds: Record<number, true>;
}

interface ArticleSessionState {
  overlays: Record<string, ArticleListOverlay>;
  setReadOverride: (contextKey: string, itemId: number, unread: boolean) => void;
  setStarOverride: (contextKey: string, itemId: number, starred: boolean) => void;
  keepVisible: (contextKey: string, itemId: number) => void;
  clearContext: (contextKey: string) => void;
}
```

- [ ] **Step 3: Keep `useUrlState()` as the place that exposes active list context values**

```ts
return {
  selectedFeedId,
  selectedGroupId,
  selectedArticleId,
  articleFilter,
  articleListContext: {
    filter: articleFilter,
    feedId: selectedFeedId,
    groupId: selectedGroupId,
  },
  setSelectedFeed,
  setSelectedGroup,
  setSelectedArticle,
  setArticleFilter,
  selectTopLevelFilter,
  selectAll,
};
```

- [ ] **Step 4: Verify the new types compile cleanly**

Run: `npx tsr generate && npx tsc -b --noEmit`
Expected: PASS with no TypeScript errors from the new overlay types


### Task 3: Convert Bookmarks to Infinite Pagination and Stable Totals

**Files:**
- Modify: `frontend/src/queries/bookmarks.ts`
- Modify: `frontend/src/lib/api/index.ts`
- Modify: `frontend/src/components/feed/feed-list.tsx`
- Modify: `frontend/src/queries/keys.ts`

- [ ] **Step 1: Add a paginated bookmarks query that mirrors the item-query shape**

```ts
const bookmarkPageSize = 100;

export const bookmarkQueries = {
  list: () =>
    infiniteQueryOptions({
      queryKey: queryKeys.bookmarks.list(),
      queryFn: async ({ pageParam }) => bookmarkAPI.list(bookmarkPageSize, pageParam),
      initialPageParam: 0,
      getNextPageParam: (lastPage, allPages) => {
        const fetched = allPages.reduce((count, page) => count + page.data.length, 0);
        return fetched < lastPage.total ? fetched : undefined;
      },
    }),
};
```

- [ ] **Step 2: Flatten bookmark pages for lookup helpers but keep `total` separate**

```ts
const pages = data?.pages ?? [];
const bookmarks = pages.flatMap((page) => page.data);
const total = pages.at(-1)?.total ?? 0;

return {
  bookmarks,
  total,
  hasNextPage,
  fetchNextPage,
  isFetchingNextPage,
  isItemStarred,
  getBookmarkByItemId,
};
```

- [ ] **Step 3: Switch the sidebar starred count to the server total**

```ts
const { total: starredCount } = useBookmarkLookup();
```

- [ ] **Step 4: Verify the starred query path still compiles**

Run: `npx tsr generate && npx tsc -b --noEmit`
Expected: PASS and no references remain to a single-page bookmarks query


### Task 4: Add a Shared Derived Article-List Hook

**Files:**
- Create: `frontend/src/queries/article-list.ts`
- Modify: `frontend/src/queries/items.ts`
- Modify: `frontend/src/queries/bookmarks.ts`

- [ ] **Step 1: Create one derived hook that merges baseline pages with the active overlay**

```ts
export function useArticleListData(context: ArticleListContext) {
  const itemsQuery = useItems({
    feedId: context.feedId,
    groupId: context.groupId,
    unread: context.filter === "unread" ? true : undefined,
  });
  const bookmarks = useStarredItems({
    feedId: context.feedId,
    groupId: context.groupId,
  });
  const overlay = useArticleListOverlay(context);

  return buildVisibleArticleList({
    filter: context.filter,
    itemPages: itemsQuery.data?.pages ?? [],
    bookmarks,
    overlay,
  });
}
```

- [ ] **Step 2: Keep pagination based on baseline server pages only**

```ts
getNextPageParam: (lastPage, allPages) => {
  const fetched = allPages.reduce((n, page) => n + page.data.length, 0);
  return fetched < lastPage.total ? fetched : undefined;
}
```

- [ ] **Step 3: Make sticky rows affect rendering, not the fetch cursor**

```ts
if (filter === "unread" && overlay.stickyVisibleIds[item.id]) {
  visible.push(applyOverrides(item, overlay));
  continue;
}
```

- [ ] **Step 4: Verify the derived hook leaves load-more math unchanged for baseline pages**

Run: `npx tsr generate && npx tsc -b --noEmit`
Expected: PASS and the `getNextPageParam` logic still counts only `page.data.length`


### Task 5: Rewire ArticleList to Use the Shared Derived List and Overlay Actions

**Files:**
- Modify: `frontend/src/components/article/article-list.tsx`
- Modify: `frontend/src/components/feed/feed-list.tsx`
- Modify: `frontend/src/hooks/use-keyboard.ts`

- [ ] **Step 1: Replace per-component starred/unread override state with the shared hook**

```ts
const { articleListContext } = useUrlState();
const articleList = useArticleListData(articleListContext);

const articles = articleList.articles;
const hasMore = articleList.hasMore;
const isLoading = articleList.isLoading;
const isLoadingMore = articleList.isLoadingMore;
```

- [ ] **Step 2: When local actions change membership, update overlay instead of dropping rows**

```ts
if (articleFilter === "unread") {
  articleSession.keepVisible(contextKey, article.id);
  articleSession.setReadOverride(contextKey, article.id, false);
}

if (articleFilter === "starred") {
  articleSession.keepVisible(contextKey, article.id);
  articleSession.setStarOverride(contextKey, article.id, false);
}
```

- [ ] **Step 3: Clear the current context overlay only on explicit reconciliation boundaries**

```ts
useEffect(() => {
  return () => {
    articleSession.clearContext(contextKey);
  };
}, [articleSession, contextKey]);
```

- [ ] **Step 4: Verify list rendering and load-more wiring compile**

Run: `npx tsr generate && npx tsc -b --noEmit`
Expected: PASS and `ArticleList` uses the shared hook for rows and pagination


### Task 6: Rewire ArticleDrawer and Final Verification

**Files:**
- Modify: `frontend/src/components/article/article-drawer.tsx`
- Modify: `frontend/src/components/article/article-list.tsx`
- Modify: `frontend/src/queries/bookmarks.ts`
- Test: `backend/internal/store/...` (no backend changes expected; run regression suite only)

- [ ] **Step 1: Make the drawer navigate over the same derived article collection as the list**

```ts
const { articleListContext } = useUrlState();
const articleList = useArticleListData(articleListContext);
const listArticles = articleList.articles;
```

- [ ] **Step 2: Ensure batch mark-as-read uses derived unread state but preserves current-row visibility**

```ts
const unreadIds = articleList.articles
  .filter((article) => article.unread && article.id > 0)
  .map((article) => article.id);
```

- [ ] **Step 3: Run frontend and backend verification**

Run: `npx tsr generate && npx tsc -b --noEmit`
Expected: PASS

Run: `go test ./...`
Expected: PASS

- [ ] **Step 4: Inspect the branch diff before review**

Run: `git status --short && git diff --stat`
Expected: only the planned frontend files, docs, and `.gitignore` have changed

- [ ] **Step 5: Commit the completed work**

```bash
git add .gitignore docs/superpowers/specs/2026-04-10-article-list-overlay-design.md docs/superpowers/plans/2026-04-10-article-list-overlay-implementation.md frontend/src/lib/article-list-context.ts frontend/src/store/article-session.ts frontend/src/hooks/use-url-state.ts frontend/src/queries/bookmarks.ts frontend/src/queries/items.ts frontend/src/queries/article-list.ts frontend/src/components/article/article-list.tsx frontend/src/components/article/article-drawer.tsx frontend/src/components/feed/feed-list.tsx frontend/src/hooks/use-keyboard.ts
git commit -m "fix(frontend): stabilize article list pagination state"
```
