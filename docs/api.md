# Fusion API

## Overview

| Item           | Value                |
| -------------- | -------------------- |
| Base URL       | `/api`               |
| Authentication | Session Cookie       |
| Content-Type   | `application/json`   |
| Timestamps     | Unix epoch (seconds) |

## Response Format

### Success Response

```json
{
  "data": { ... }
}
```

For list endpoints:

```json
{
  "data": [ ... ],
  "total": 100
}
```

### Error Response

```json
{
  "error": "error message"
}
```

### HTTP Status Codes

| Code | Description           |
| ---- | --------------------- |
| 200  | Success               |
| 201  | Created               |
| 400  | Bad Request           |
| 401  | Unauthorized          |
| 404  | Not Found             |
| 500  | Internal Server Error |

## Authentication

### Login

```
POST /api/sessions
```

**Request Body**

```json
{
  "password": "string"
}
```

**Response** `200 OK`

```json
{
  "data": {
    "message": "ok"
  }
}
```

Sets `session` cookie for subsequent requests.

### Logout

```
DELETE /api/sessions
```

**Response** `200 OK`

```json
{
  "data": {
    "message": "ok"
  }
}
```

Clears session cookie.

## Groups

### List Groups

```
GET /api/groups
```

**Response** `200 OK`

```json
{
  "data": [
    {
      "id": 1,
      "name": "Default",
      "created_at": 1703001600,
      "updated_at": 1703001600
    }
  ]
}
```

### Create Group

```
POST /api/groups
```

**Request Body**

```json
{
  "name": "Tech"
}
```

**Response** `201 Created`

```json
{
  "data": {
    "id": 2,
    "name": "Tech",
    "created_at": 1703001600,
    "updated_at": 1703001600
  }
}
```

### Update Group

```
PATCH /api/groups/:id
```

**Request Body**

```json
{
  "name": "Technology"
}
```

**Response** `200 OK`

```json
{
  "data": {
    "id": 2,
    "name": "Technology",
    "created_at": 1703001600,
    "updated_at": 1703088000
  }
}
```

### Delete Group

```
DELETE /api/groups/:id
```

Deleting a group moves its feeds to the Default group (id=1).

**Response** `200 OK`

```json
{
  "data": {
    "message": "ok"
  }
}
```

## Feeds

### List Feeds

```
GET /api/feeds
```

**Query Parameters**

| Parameter | Type   | Description     |
| --------- | ------ | --------------- |
| group_id  | number | Filter by group |

**Response** `200 OK`

```json
{
  "data": [
    {
      "id": 1,
      "group_id": 1,
      "name": "Hacker News",
      "link": "https://news.ycombinator.com/rss",
      "site_url": "https://news.ycombinator.com",
      "last_build": 1703001600,
      "failure": "",
      "failures": 0,
      "suspended": false,
      "proxy": "",
      "unread_count": 42,
      "created_at": 1703001600,
      "updated_at": 1703001600
    }
  ]
}
```

### Get Feed

```
GET /api/feeds/:id
```

**Response** `200 OK`

```json
{
  "data": {
    "id": 1,
    "group_id": 1,
    "name": "Hacker News",
    "link": "https://news.ycombinator.com/rss",
    "site_url": "https://news.ycombinator.com",
    "last_build": 1703001600,
    "failure": "",
    "failures": 0,
    "suspended": false,
    "proxy": "",
    "unread_count": 42,
    "created_at": 1703001600,
    "updated_at": 1703001600
  }
}
```

### Create Feed

```
POST /api/feeds
```

**Request Body**

```json
{
  "link": "https://example.com/feed.xml",
  "group_id": 1
}
```

Optional fields: `name`, `proxy`

**Response** `201 Created`

```json
{
  "data": {
    "id": 2,
    "group_id": 1,
    "name": "Example Blog",
    "link": "https://example.com/feed.xml",
    "site_url": "https://example.com",
    "last_build": 0,
    "failure": "",
    "failures": 0,
    "suspended": false,
    "proxy": "",
    "unread_count": 0,
    "created_at": 1703001600,
    "updated_at": 1703001600
  }
}
```

### Batch Create Feeds

```
POST /api/feeds/batch
```

Creates multiple feeds in a single transaction. Skips feeds with duplicate links.

**Request Body**

```json
{
  "feeds": [
    {
      "group_id": 1,
      "name": "Feed One",
      "link": "https://example.com/feed1.xml"
    },
    {
      "group_id": 2,
      "name": "Feed Two",
      "link": "https://example.com/feed2.xml"
    }
  ]
}
```

**Response** `200 OK`

```json
{
  "data": {
    "created": 2,
    "failed": 0,
    "errors": []
  }
}
```

If some feeds fail (e.g., duplicate links):

```json
{
  "data": {
    "created": 1,
    "failed": 1,
    "errors": ["duplicate feed: https://example.com/feed1.xml"]
  }
}
```

### Validate Feed URL

```
POST /api/feeds/validation
```

Discovers feeds from a URL. If the URL is an HTML page, returns discovered feed links.

**Request Body**

```json
{
  "url": "https://example.com"
}
```

**Response** `200 OK`

```json
{
  "data": {
    "feeds": [
      {
        "title": "Example Blog",
        "link": "https://example.com/feed.xml"
      },
      {
        "title": "Example Blog - Comments",
        "link": "https://example.com/comments/feed"
      }
    ]
  }
}
```

### Update Feed

```
PATCH /api/feeds/:id
```

**Request Body**

```json
{
  "name": "New Name",
  "group_id": 2,
  "suspended": false,
  "proxy": "http://proxy:8080"
}
```

All fields optional.

**Response** `200 OK`

```json
{
  "data": {
    "id": 1,
    "group_id": 2,
    "name": "New Name",
    "link": "https://example.com/feed.xml",
    "site_url": "https://example.com",
    "last_build": 1703001600,
    "failure": "",
    "failures": 0,
    "suspended": false,
    "proxy": "http://proxy:8080",
    "unread_count": 42,
    "created_at": 1703001600,
    "updated_at": 1703088000
  }
}
```

### Delete Feed

```
DELETE /api/feeds/:id
```

Deletes the feed and all its items. Bookmarks are preserved (item_id set to null).

**Response** `200 OK`

```json
{
  "data": {
    "message": "ok"
  }
}
```

### Refresh Feeds

```
POST /api/feeds/refresh
```

Triggers an asynchronous refresh of all feeds. Optionally refresh a single feed. The response returns immediately; feeds are refreshed in the background.

**Request Body** (optional)

```json
{
  "feed_id": 1
}
```

**Response** `200 OK`

```json
{
  "data": {
    "message": "ok"
  }
}
```

## Items

### List Items

```
GET /api/items
```

**Query Parameters**

| Parameter | Type   | Default     | Description                                                               |
| --------- | ------ | ----------- | ------------------------------------------------------------------------- |
| unread    | bool   | -           | Filter by unread status                                                   |
| feed_id   | number | -           | Filter by feed                                                            |
| group_id  | number | -           | Filter by group                                                           |
| sort      | string | `-pub_date` | Sort order (prefix `-` for descending). Allowed: `pub_date`, `created_at` |
| page      | number | 1           | Page number                                                               |
| limit     | number | 20          | Items per page (max 100)                                                  |

**Response** `200 OK`

```json
{
  "data": [
    {
      "id": 1,
      "feed_id": 1,
      "feed_name": "Hacker News",
      "guid": "https://example.com/post/123",
      "title": "Article Title",
      "link": "https://example.com/post/123",
      "content": "<p>Article content...</p>",
      "pub_date": 1703001600,
      "unread": true,
      "created_at": 1703001600
    }
  ],
  "total": 100
}
```

### Get Item

```
GET /api/items/:id
```

**Response** `200 OK`

```json
{
  "data": {
    "id": 1,
    "feed_id": 1,
    "feed_name": "Hacker News",
    "guid": "https://example.com/post/123",
    "title": "Article Title",
    "link": "https://example.com/post/123",
    "content": "<p>Full article content...</p>",
    "pub_date": 1703001600,
    "unread": true,
    "created_at": 1703001600
  }
}
```

### Batch Update Read Status

```
PATCH /api/items/-/unread
```

Update read/unread status for multiple items.

**Request Body**

```json
{
  "unread": false,
  "item_ids": [1, 2, 3]
}
```

Or mark all items in a feed/group:

```json
{
  "unread": false,
  "feed_id": 1
}
```

```json
{
  "unread": false,
  "group_id": 1
}
```

Or mark all items:

```json
{
  "unread": false,
  "all": true
}
```

**Response** `200 OK`

```json
{
  "data": {
    "updated": 3
  }
}
```

## Bookmarks

Bookmarks store a snapshot of item content, surviving feed/item deletion.

### List Bookmarks

```
GET /api/bookmarks
```

**Query Parameters**

| Parameter | Type   | Default | Description              |
| --------- | ------ | ------- | ------------------------ |
| page      | number | 1       | Page number              |
| limit     | number | 20      | Items per page (max 100) |

**Response** `200 OK`

```json
{
  "data": [
    {
      "id": 1,
      "item_id": 123,
      "link": "https://example.com/post/123",
      "title": "Article Title",
      "content": "<p>Article content...</p>",
      "pub_date": 1703001600,
      "feed_name": "Hacker News",
      "created_at": 1703001600
    }
  ],
  "total": 10
}
```

Note: `item_id` may be `null` if the original item was deleted.

### Add Bookmark

```
POST /api/bookmarks
```

**Request Body**

```json
{
  "item_id": 123
}
```

**Response** `201 Created`

```json
{
  "data": {
    "id": 1,
    "item_id": 123,
    "link": "https://example.com/post/123",
    "title": "Article Title",
    "content": "<p>Article content...</p>",
    "pub_date": 1703001600,
    "feed_name": "Hacker News",
    "created_at": 1703001600
  }
}
```

### Delete Bookmark

```
DELETE /api/bookmarks/:id
```

**Response** `200 OK`

```json
{
  "data": {
    "message": "ok"
  }
}
```
