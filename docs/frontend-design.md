# Fusion Frontend Design

## 1. Tech Stack

- **Framework**: React 19 + TypeScript
- **Routing**: TanStack Router
- **UI Components**: shadcn/ui + Tailwind CSS
- **State Management**: Zustand

## 2. Layout Design

### 2.1 Three-Column Layout (Desktop)

```
┌──────────┬─────────────┬──────────────────┐
│ Sidebar  │  Item List  │   Item Content   │
│  240px   │   320px     │     flex-1       │
└──────────┴─────────────┴──────────────────┘
```

- **Sidebar (240px)**: Navigation + Feed tree
- **Item List (320px)**: Tabs + article list
- **Item Content (flex-1)**: Article reading area

### 2.2 Responsive Breakpoints

| Breakpoint      | Layout        | Description                               |
| --------------- | ------------- | ----------------------------------------- |
| lg (≥1024px)    | Three columns | Full layout                               |
| md (768-1023px) | Two columns   | Sidebar as drawer, list + content visible |
| sm (<768px)     | Single column | List or content, toggle between views     |

### 2.3 Responsive Behavior

**Desktop (≥1024px)**

- All three columns visible
- Sidebar fixed, scrollable independently

**Tablet (768-1023px)**

- Sidebar hidden, accessible via hamburger menu (drawer)
- Item list and content side by side

**Mobile (<768px)**

- Single column view
- Toggle between list view and content view
- Back button to return to list from content

## 3. Component Structure

### 3.1 Sidebar

```
Sidebar
├── Header
│   ├── Logo
│   └── Theme Toggle
├── Add Feed Button
├── Navigation Menu
│   └── All
├── Feeds Section
│   └── Groups (collapsible)
│       ├── Group Header (name + action button)
│       └── Feeds
│           └── Feed Item (name + unread count + action button)
└── User Menu
    ├── Settings
    └── Logout
```

**Features**:

- Groups are collapsible/expandable
- Each feed shows unread count badge
- Active item is highlighted
- Keyboard navigation support (↑↓)
- **Feed/Group actions**: Available via both right-click context menu AND explicit action button (three-dot icon)

### 3.2 Item List (Middle Column)

```
Item List
├── Header
│   ├── Tabs: All | Unread | Bookmarks
│   └── Feed Filter Indicator (if filtered)
├── Item Entries
│   └── ItemCard (repeating)
│       ├── Title
│       ├── Feed name + publish date
│       └── Unread indicator
└── Pagination
```

**Features**:

- Tabs control which items to show
- Clicking a feed in sidebar filters current tab
- Selected item is highlighted
- Keyboard navigation (j/k)

### 3.3 Item Content (Right Column)

```
Item Content
├── Header
│   ├── Title
│   ├── Feed name + publish date
│   └── Action Buttons
│       ├── Mark as read/unread
│       ├── Bookmark/unbookmark
│       └── Open original link
├── Article Body (prose styling)
└── Navigation
    ├── Previous item
    └── Next item
```

**Features**:

- Max width constrained for readability (prose)
- HTML content sanitized (DOMPurify)
- Images lazy loaded
- Links open in new tab

## 4. Routing Design

### 4.1 Routes

| Route    | Description            |
| -------- | ---------------------- |
| `/`      | Main three-column view |
| `/login` | Login page             |

### 4.2 URL Parameters

All state is managed via URL search parameters:

| Parameter | Values                       | Description                     |
| --------- | ---------------------------- | ------------------------------- |
| `tab`     | `all`, `unread`, `bookmarks` | Current tab (default: `unread`) |
| `feed`    | Feed ID (number)             | Filter by specific feed         |
| `group`   | Group ID (number)            | Filter by specific group        |
| `item`    | Item ID (number)             | Currently viewed item           |

**Example URLs**:

- `/?tab=unread` - Show unread items
- `/?tab=all&feed=5` - Show all items from feed #5
- `/?tab=unread&item=123` - Show unread items, viewing item #123

### 4.3 Navigation Behavior

**Clicking a Feed**:

1. Add `?feed={id}` to URL
2. Keep current `tab` parameter
3. Clear `item` parameter

**Clicking a Group**:

1. Add `?group={id}` to URL
2. Keep current `tab` parameter
3. Clear `feed` and `item` parameters

**Clicking an Item**:

1. Add `?item={id}` to URL
2. Keep `tab`, `feed`, `group` parameters

## 5. Interaction Design

### 5.1 Global Search

- **Trigger**: `Cmd/Ctrl + K`
- **Style**: Spotlight/Command palette (shadcn Command component)
- **Scope**: Search item titles and content
- **Behavior**: Results shown in dropdown, click to navigate

### 5.2 Keyboard Shortcuts

| Key            | Action                |
| -------------- | --------------------- |
| `j`            | Next item in list     |
| `k`            | Previous item in list |
| `o` or `Enter` | Open selected item    |
| `u`            | Toggle read/unread    |
| `b`            | Toggle bookmark       |
| `v`            | Open original link    |
| `Cmd/Ctrl + K` | Open search           |
| `?`            | Show shortcuts help   |

### 5.3 Feed Click Behavior

When clicking a feed in the sidebar:

- **Preserve**: Current tab (All/Unread/Bookmarks)
- **Filter**: Show only items from that feed
- **Clear**: Previously selected item

### 5.4 Empty States

| State             | Message                              |
| ----------------- | ------------------------------------ |
| No feeds          | "Add your first feed to get started" |
| No items in feed  | "No items in this feed"              |
| No unread items   | "All caught up!"                     |
| No bookmarks      | "No bookmarked items yet"            |
| No search results | "No results found"                   |

## 6. Theme

### 6.1 Color Scheme

- Light mode (default)
- Dark mode
- System preference detection

### 6.2 Theme Toggle

- Located in sidebar header
- Persisted in localStorage
- Sun/Moon icon toggle

## 7. Feed & Group Management

Feed and Group management is **contextual** rather than centralized in settings.

### 7.1 Access Methods

| Method        | Trigger                   | Device      |
| ------------- | ------------------------- | ----------- |
| Context menu  | Right-click on feed/group | Desktop     |
| Action button | Click three-dot icon (⋮)  | All devices |

### 7.2 Feed Actions (Context Menu / Dropdown)

- Mark all as read
- Rename
- Move to group
- Edit (URL, proxy settings)
- Delete

### 7.3 Group Actions (Context Menu / Dropdown)

- Mark all as read
- Rename
- Delete (with confirmation)

### 7.4 Sidebar Header Actions

- **Add Feed**: Button to subscribe to new feed
- **Import OPML**: Import feeds from file
- **Export OPML**: Export all feeds

### 7.5 Create Group

- Triggered from "Move to group" action
- Option to create new group inline

## 8. Settings

Settings presented as a Dialog, containing only app-wide preferences.

### 8.1 Settings Sections

**Appearance**

- Theme selection (Light/Dark/System)

**Data**

- Import OPML
- Export OPML

**Account**

- Logout

## 9. Component Library

Using shadcn/ui components:

| Component      | Usage                            |
| -------------- | -------------------------------- |
| Button         | Actions, navigation              |
| Input          | Forms, search                    |
| Tabs           | List filtering                   |
| Dialog         | Settings, confirmations          |
| Sheet          | Mobile sidebar, settings drawer  |
| DropdownMenu   | User menu, feed/group actions    |
| ContextMenu    | Right-click menus for feed/group |
| ScrollArea     | Scrollable lists                 |
| Separator      | Visual dividers                  |
| Command        | Global search (Spotlight)        |
| Sonner (Toast) | Notifications                    |

## 10. State Management

Using Zustand for global state:

```typescript
interface AppState {
  // Sidebar
  sidebarOpen: boolean;

  // Data
  groups: Group[];
  feeds: Feed[];

  // Actions
  toggleSidebar: () => void;
  setGroups: (groups: Group[]) => void;
  setFeeds: (feeds: Feed[]) => void;
  updateFeedUnreadCount: (feedId: number, delta: number) => void;
}
```

**Note**: Most UI state (tab, feed filter, selected item) is in URL, not Zustand.

## 11. File Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── ui/                 # shadcn/ui components
│   │   ├── layout/
│   │   │   ├── Sidebar.tsx
│   │   │   ├── ItemList.tsx
│   │   │   └── ItemContent.tsx
│   │   ├── feed/
│   │   │   ├── FeedItem.tsx
│   │   │   └── AddFeedDialog.tsx
│   │   ├── item/
│   │   │   ├── ItemCard.tsx
│   │   │   └── ItemActions.tsx
│   │   └── settings/
│   │       └── SettingsDialog.tsx
│   ├── routes/
│   │   ├── __root.tsx          # Root layout
│   │   ├── index.tsx           # Main view (/)
│   │   └── login.tsx           # Login page
│   ├── lib/
│   │   ├── api/                # API client
│   │   └── utils.ts            # Utility functions
│   ├── store/
│   │   └── app.ts              # Zustand store
│   └── main.tsx
├── package.json
└── vite.config.ts
```
