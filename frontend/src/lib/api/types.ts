// Core data models (matching backend/internal/model/model.go)
export interface Group {
  id: number;
  name: string;
  created_at: number;
  updated_at: number;
}

export interface Feed {
  id: number;
  group_id: number;
  name: string;
  link: string;
  site_url?: string;
  suspended: boolean;
  proxy?: string;
  created_at: number;
  updated_at: number;
  fetch_state: FeedFetchState;
  unread_count: number;
  item_count: number;
}

export interface FeedFetchState {
  etag?: string;
  last_modified?: string;
  cache_control?: string;
  expires_at: number;
  last_checked_at: number;
  next_check_at: number;
  last_http_status: number;
  retry_after_until: number;
  last_success_at: number;
  last_error_at: number;
  last_error?: string;
  consecutive_failures: number;
}

export interface Item {
  id: number;
  feed_id: number;
  guid: string;
  title: string;
  link: string;
  content: string;
  pub_date: number;
  unread: boolean;
  created_at: number;
}

export interface Bookmark {
  id: number;
  item_id: number | null;
  link: string;
  title: string;
  content: string;
  pub_date: number;
  feed_name: string;
  feed_id: number | null;
  unread: boolean;
  created_at: number;
}

// User preferences synced via the backend. A null field means the user has
// never explicitly chosen a value; clients fall back to their own defaults.
export interface Settings {
  locale: string | null;
  article_page_size: number | null;
  theme: string | null;
  auto_mark_read: string | null;
  updated_at: number;
}

// Only the fields being changed are sent; a missing field leaves the stored
// value unchanged (indistinguishable from an explicit null on the wire).
export interface UpdateSettingsRequest {
  locale?: string;
  article_page_size?: number;
  theme?: string;
  auto_mark_read?: string;
}

// API response wrappers
export interface APIResponse<T> {
  data?: T;
  error?: string;
}

// Group and feed lists: small, fully returned in one response
export interface ListResponse<T> {
  data: T[];
  total: number;
}

// Item and bookmark lists: cursor-paginated (see docs/api-conventions.md)
export interface PaginatedListResponse<T> {
  data: T[];
  total: number;
  next_cursor: string | null;
}

// Request types
export interface LoginRequest {
  password: string;
}

export interface CreateGroupRequest {
  name: string;
}

export interface UpdateGroupRequest {
  name: string;
}

export interface CreateFeedRequest {
  group_id: number;
  name: string;
  link: string;
  site_url?: string;
  proxy?: string;
}

export interface UpdateFeedRequest {
  group_id?: number;
  name?: string;
  link?: string;
  site_url?: string;
  suspended?: boolean;
  proxy?: string;
}

export interface ValidateFeedRequest {
  url: string;
  proxy?: string;
}

export interface DiscoveredFeed {
  title: string;
  link: string;
}

export interface ValidateFeedResponse {
  feeds: DiscoveredFeed[];
}

// Two mutually exclusive modes: either reference an item by id (the backend
// snapshots all fields from it and ignores everything else — preferred), or
// provide a full snapshot for bookmarks without an backing item.
export type CreateBookmarkRequest =
  | { item_id: number }
  | {
      link: string;
      title: string;
      content: string;
      feed_name: string;
      pub_date?: number;
    };

export interface MarkItemsReadRequest {
  ids: number[];
}

export interface ListItemsParams {
  feed_id?: number;
  group_id?: number;
  unread?: boolean;
  limit?: number;
  before?: string;
  order_by?: "pub_date" | "created_at";
}

export interface ListBookmarksParams {
  feed_id?: number;
  group_id?: number;
  limit?: number;
  before?: string;
}

export interface BatchCreateFeedsRequest {
  feeds: Array<{
    group_id: number;
    name: string;
    link: string;
    site_url?: string;
  }>;
}

export interface SearchFeed {
  id: number;
  name: string;
  link: string;
  site_url: string;
}

export interface SearchItem {
  id: number;
  feed_id: number;
  title: string;
  pub_date: number;
}

export interface SearchResponse {
  feeds: SearchFeed[];
  items: SearchItem[];
}

export interface BatchCreateFeedsResponse {
  created: number;
  failed: number;
  errors?: string[];
}

// OIDC
export interface OIDCStatusResponse {
  enabled: boolean;
}

export interface OIDCLoginResponse {
  auth_url: string;
}
