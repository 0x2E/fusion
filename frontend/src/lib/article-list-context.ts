import type { ArticleFilter } from "@/lib/article-filter";

export interface ArticleListContext {
  filter: ArticleFilter;
  feedId: number | null;
  groupId: number | null;
}

export function getArticleListContextKey(context: ArticleListContext): string {
  return `${context.filter}:${context.feedId ?? "all"}:${context.groupId ?? "all"}`;
}
