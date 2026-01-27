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
  last_build: number;
  failure?: string;
  failures: number;
  suspended: boolean;
  proxy?: string;
  created_at: number;
  updated_at: number;
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
  created_at: number;
}

// API response wrappers
export interface APIResponse<T> {
  data?: T;
  error?: string;
}

export interface ListAPIResponse<T> {
  data: T[];
  total: number;
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
  site_url?: string;
  proxy?: string;
}

export interface ValidateFeedRequest {
  url: string;
}

export interface CreateBookmarkRequest {
  item_id?: number;
  link: string;
  title: string;
  content: string;
  pub_date: number;
  feed_name: string;
}

export interface MarkItemsReadRequest {
  ids: number[];
}

export interface ListItemsParams {
  feed_id?: number;
  group_id?: number;
  unread?: boolean;
  limit?: number;
  offset?: number;
  order_by?: string;
}

export interface ImportOpmlResponse {
  imported: number;
  failed: number;
  errors?: string[];
}
