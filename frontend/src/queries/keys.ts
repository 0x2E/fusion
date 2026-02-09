export interface ItemFilters {
  feedId?: number | null;
  groupId?: number | null;
  unread?: boolean;
}

export interface NormalizedItemFilters {
  feedId: number | null;
  groupId: number | null;
  unread: boolean;
}

export function normalizeItemFilters(
  filters: ItemFilters,
): NormalizedItemFilters {
  return {
    feedId: filters.feedId ?? null,
    groupId: filters.groupId ?? null,
    unread: filters.unread ?? false,
  };
}

export const queryKeys = {
  groups: {
    all: ["groups"] as const,
    list: () => [...queryKeys.groups.all, "list"] as const,
  },
  feeds: {
    all: ["feeds"] as const,
    list: () => [...queryKeys.feeds.all, "list"] as const,
  },
  items: {
    all: ["items"] as const,
    lists: () => [...queryKeys.items.all, "list"] as const,
    list: (filters: ItemFilters) =>
      [...queryKeys.items.all, "list", normalizeItemFilters(filters)] as const,
    details: () => [...queryKeys.items.all, "detail"] as const,
    detail: (id: number) => [...queryKeys.items.all, "detail", id] as const,
  },
  bookmarks: {
    all: ["bookmarks"] as const,
    list: () => [...queryKeys.bookmarks.all, "list"] as const,
  },
};
