import { api } from "./client";
import type {
  APIResponse,
  ListAPIResponse,
  LoginRequest,
  Group,
  Feed,
  Item,
  Bookmark,
  CreateGroupRequest,
  UpdateGroupRequest,
  CreateFeedRequest,
  UpdateFeedRequest,
  ValidateFeedRequest,
  ValidateFeedResponse,
  CreateBookmarkRequest,
  MarkItemsReadRequest,
  ListItemsParams,
  ImportOpmlResponse,
  BatchCreateFeedsRequest,
  BatchCreateFeedsResponse,
  SearchResponse,
  OIDCStatusResponse,
  OIDCLoginResponse,
} from "./types";

// Session APIs
export const sessionAPI = {
  login: (data: LoginRequest) =>
    api.post<APIResponse<{ message: string }>>("/sessions", data),

  logout: () => api.delete<APIResponse<{ message: string }>>("/sessions"),
};

// OIDC APIs
export const oidcAPI = {
  status: () => api.get<APIResponse<OIDCStatusResponse>>("/oidc/enabled"),

  login: () => api.get<APIResponse<OIDCLoginResponse>>("/oidc/login"),
};

// Group APIs
export const groupAPI = {
  list: () => api.get<ListAPIResponse<Group>>("/groups"),

  get: (id: number) => api.get<APIResponse<Group>>(`/groups/${id}`),

  create: (data: CreateGroupRequest) =>
    api.post<APIResponse<Group>>("/groups", data),

  update: (id: number, data: UpdateGroupRequest) =>
    api.patch<APIResponse<Group>>(`/groups/${id}`, data),

  delete: (id: number) =>
    api.delete<APIResponse<{ message: string }>>(`/groups/${id}`),
};

// Feed APIs
export const feedAPI = {
  list: () => api.get<ListAPIResponse<Feed>>("/feeds"),

  get: (id: number) => api.get<APIResponse<Feed>>(`/feeds/${id}`),

  create: (data: CreateFeedRequest) =>
    api.post<APIResponse<Feed>>("/feeds", data),

  update: (id: number, data: UpdateFeedRequest) =>
    api.patch<APIResponse<Feed>>(`/feeds/${id}`, data),

  delete: (id: number) =>
    api.delete<APIResponse<{ message: string }>>(`/feeds/${id}`),

  validate: (data: ValidateFeedRequest) =>
    api.post<APIResponse<ValidateFeedResponse>>("/feeds/validate", data),

  refresh: () =>
    api.post<APIResponse<{ message: string }>>("/feeds/refresh", {}),

  batchCreate: (data: BatchCreateFeedsRequest) =>
    api.post<APIResponse<BatchCreateFeedsResponse>>("/feeds/batch", data),
};

// Item APIs
export const itemAPI = {
  list: (params?: ListItemsParams) => {
    const query = new URLSearchParams();
    if (params?.feed_id) query.set("feed_id", params.feed_id.toString());
    if (params?.group_id) query.set("group_id", params.group_id.toString());
    if (params?.unread !== undefined)
      query.set("unread", params.unread.toString());
    if (params?.limit) query.set("limit", params.limit.toString());
    if (params?.offset) query.set("offset", params.offset.toString());
    if (params?.order_by) query.set("order_by", params.order_by);

    const queryString = query.toString();
    return api.get<ListAPIResponse<Item>>(
      `/items${queryString ? `?${queryString}` : ""}`,
    );
  },

  get: (id: number) => api.get<APIResponse<Item>>(`/items/${id}`),

  markRead: (data: MarkItemsReadRequest) =>
    api.patch<APIResponse<{ message: string }>>("/items/-/read", data),

  markUnread: (data: MarkItemsReadRequest) =>
    api.patch<APIResponse<{ message: string }>>("/items/-/unread", data),
};

// Bookmark APIs
export const bookmarkAPI = {
  list: (limit = 50, offset = 0) => {
    const query = new URLSearchParams({
      limit: limit.toString(),
      offset: offset.toString(),
    });
    return api.get<ListAPIResponse<Bookmark>>(`/bookmarks?${query}`);
  },

  get: (id: number) => api.get<APIResponse<Bookmark>>(`/bookmarks/${id}`),

  create: (data: CreateBookmarkRequest) =>
    api.post<APIResponse<Bookmark>>("/bookmarks", data),

  delete: (id: number) =>
    api.delete<APIResponse<{ message: string }>>(`/bookmarks/${id}`),
};

// Search APIs
export const searchAPI = {
  search: (q: string, limit = 10) =>
    api.get<APIResponse<SearchResponse>>(
      `/search?q=${encodeURIComponent(q)}&limit=${limit}`,
    ),
};

// OPML APIs
const API_BASE = import.meta.env.VITE_API_BASE_URL || "/api";

export const opmlAPI = {
  import: async (file: File): Promise<APIResponse<ImportOpmlResponse>> => {
    const formData = new FormData();
    formData.append("file", file);

    const response = await fetch(`${API_BASE}/opml/import`, {
      method: "POST",
      credentials: "include",
      body: formData,
    });

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ error: "Unknown error" }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return response.json();
  },
};

export * from "./types";
export { APIError, setUnauthorizedCallback } from "./client";
