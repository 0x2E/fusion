# Fusion Frontend Design

## 1. Tech Stack

- **Framework**: React 19 + TypeScript
- **Routing**: TanStack Router
- **UI Components**: shadcn/ui + Tailwind CSS
- **State Management**: Zustand

## 2. Layout Design

### 2.1 Two-Column Layout (Desktop)

```
┌─────────────┬──────────────────────────────────┐
│   Sidebar   │        Main Content              │
│    240px    │           flex-1                 │
│             │                                  │
│             │  ┌─────────────────────────┐    │
│             │  │   Article List          │    │
│             │  │   (Tabs + Cards)        │    │
│             │  └─────────────────────────┘    │
│             │                                  │
│             │  ┌─────────────────────────┐    │
│             │  │  Article Detail Drawer  │    │
│             │  │  (Sheet, slide right)   │    │
│             │  └─────────────────────────┘    │
└─────────────┴──────────────────────────────────┘
```

- **Sidebar (240px)**: Collapsible, contains logo, search, feed list, settings, and version
- **Main Content (flex-1)**: Article list with filter tabs, article detail drawer slides from right

### 2.2 Responsive Breakpoints

| Breakpoint      | Layout            | Description                        |
| --------------- | ----------------- | ---------------------------------- |
| lg (≥1024px)    | Full two-column   | Sidebar + Main content both shown  |
| md (768-1023px) | Collapsed sidebar | Sidebar as icons, main content full width |
| sm (<768px)     | Mobile drawer     | Sidebar as drawer, single column   |

### 2.3 Responsive Behavior

**Desktop (≥1024px)**

- Sidebar and main content both visible
- Article detail opens as drawer from right side
- Each area scrollable independently

**Tablet (768-1023px)**

- Sidebar collapses to icon mode
- Main content takes full width
- Click sidebar icon to expand

**Mobile (<768px)**

- Sidebar becomes a drawer (Sheet)
- Single column view for article list
- Article detail opens as full-screen overlay

## 3. Component Structure

### 3.1 Sidebar

```
Sidebar (collapsible="icon", 240px)
├── Header
│   ├── Logo + App Name ("Fusion")
│   └── Search Button
├── Feed List
│   ├── Create Group Button (at list title)
│   ├── Default "All" Group (non-collapsible)
│   └── Custom Groups (collapsible)
│       ├── Group Header
│       │   ├── Group Name
│       │   ├── Add Feed Button (show on hover)
│       │   └── Context Menu Button (show on hover / right-click)
│       └── Feeds
│           └── Feed Item
│               ├── Logo + Name + Unread Badge
│               └── Context Menu Button (show on hover / right-click)
└── Footer
    ├── Settings Button
    └── Version Number
```

**Features**:

- **Header**: Contains logo, app name, and search button for quick access
- **Create Group Button**: Positioned at feed list title, triggers dialog for creating new groups
- **Default "All" Group**: Always visible, shows all items, cannot be deleted or renamed
- **Custom Groups**: User-created groups, collapsible/expandable
- **Hover Interactions**:
  - Group headers show "Add Feed" button and menu button on hover
  - Feed items show menu button on hover
  - Right-click also opens context menu for both groups and feeds
- **Unread Badges**: Each feed displays unread count (SidebarMenuBadge)
- **Active Highlighting**: Currently selected feed/group is highlighted
- **Keyboard Navigation**: Support arrow keys (↑↓) for navigation

### 3.2 Main Content Area

```
Main Content (flex-1)
├── Filter Tabs
│   ├── All
│   ├── Unread
│   └── Bookmarked
└── Article List (scrollable)
    └── Article Card
        ├── Title
        ├── Summary (excerpt)
        └── Quick Actions (show on hover, right-aligned)
            ├── Mark as read/unread (icon button)
            ├── Bookmark/unbookmark (icon button)
            └── Open original link (icon button)
```

**Features**:

- **Filter Tabs**: Switch between All, Unread, and Bookmarked articles
- **Article Cards**: Display title and summary for each article
- **Hover Actions**: Quick action buttons appear on right side when hovering over article card
- **Click Behavior**: Clicking article card opens detail drawer
- **Keyboard Navigation**: Support j/k for next/previous article
- **Empty States**: Show helpful messages when no articles match current filter

### 3.3 Article Detail Drawer

```
Sheet (slide from right, ~60-70% viewport width)
├── Header
│   ├── Close Button (X)
│   └── Action Buttons
│       ├── Mark as read/unread
│       ├── Bookmark/unbookmark
│       └── Open original link (external icon)
├── Article Content (scrollable)
│   ├── Title
│   ├── Feed Name + Publish Date (metadata)
│   ├── Original Link (clickable)
│   └── Body (sanitized HTML)
└── Navigation Buttons (overlay on margins)
    ├── Previous Article (left margin)
    └── Next Article (right margin)
```

**Features**:

- **Sheet Component**: Slides in from right side, overlays main content
- **Width**: 60-70% of viewport width for comfortable reading
- **Content Styling**: Max-width constrained with prose styling for readability
- **HTML Sanitization**: Article content cleaned with DOMPurify
- **Image Loading**: Lazy loaded images for performance
- **External Links**: Open in new tab
- **Side Navigation**: Previous/Next buttons positioned in left/right margins (floating style)
- **Close Behavior**: Click overlay, close button, or press Esc to close drawer
- **URL Sync**: Opening/closing drawer updates URL parameter `?item={id}`

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

**Clicking an Article Card**:

1. Add `?item={id}` to URL
2. Keep `tab`, `feed`, `group` parameters
3. Open article detail drawer (Sheet)

**Closing Article Drawer**:

1. Remove `?item` parameter from URL
2. Keep `tab`, `feed`, `group` parameters
3. Close drawer with animation

## 5. Interaction Design

### 5.1 Global Search

- **Trigger**:
  - Sidebar search button
  - Keyboard shortcut: `Cmd/Ctrl + K`
- **Style**: Dialog with Command component (spotlight-style)
- **Search Types**:
  - **Feeds**: Search by feed name or URL
  - **Articles**: Search by article title and content
  - **Unified**: Search across both feeds and articles
- **UI**:
  - Dialog overlay with search input
  - Tab or radio group to switch search type
  - Live results displayed as list
- **Behavior**: Click result to navigate to feed or open article drawer

### 5.2 Keyboard Shortcuts

| Key            | Action                              |
| -------------- | ----------------------------------- |
| `j`            | Next article in list                |
| `k`            | Previous article in list            |
| `o` or `Enter` | Open selected article               |
| `u`            | Toggle read/unread                  |
| `b`            | Toggle bookmark                     |
| `v`            | Open original link                  |
| `←` / `→`      | Previous/Next article (in drawer)   |
| `Esc`          | Close article drawer                |
| `Cmd/Ctrl + K` | Open search                         |
| `?`            | Show shortcuts help                 |

### 5.3 Hover Interactions

**Article Card Hover**:
- Quick action buttons appear on right side
- Buttons: Mark read/unread, Bookmark, Open link
- Icon-only buttons with tooltips

**Feed Item Hover**:
- Context menu button (three dots) appears
- Right-click also shows context menu

**Group Header Hover**:
- "Add Feed" button appears
- Context menu button appears
- Right-click also shows context menu

### 5.4 Feed Click Behavior

When clicking a feed in the sidebar:

- **Preserve**: Current tab (All/Unread/Bookmarked)
- **Filter**: Show only items from that feed
- **Clear**: Previously selected item (close drawer if open)

### 5.5 Empty States

| State             | Message                              |
| ----------------- | ------------------------------------ |
| No feeds          | "Add your first feed to get started" |
| No items in feed  | "No items in this feed"              |
| No unread items   | "All caught up!"                     |
| No bookmarks      | "No bookmarked items yet"            |
| No search results | "No results found"                   |

## 6. Theme

### 6.1 Design Philosophy

**Notion-inspired Design**:
- Clean, minimal interface
- Generous white space and breathing room
- Subtle borders and shadows
- Smooth transitions and interactions
- Focus on content readability
- Professional yet approachable aesthetic

### 6.2 Color Scheme

**Light Mode (Default)**:
- **Background**: Warm whites (#FFFFFF, #FAFAFA) - Notion-like warmth
- **Text**: Dark grays (#1F1F1F, #37352F) - High contrast for readability
- **Accents**: Subtle blues/grays (#0F62FE, #8E8E8E)
- **Borders**: Light grays (#E9E9E7, #EFEFEF)
- **Hover States**: Soft gray backgrounds (#F7F6F3)

**Dark Mode**:
- **Background**: Dark grays (#191919, #2F2F2F) - Not pure black for reduced eye strain
- **Text**: Light grays (#E3E2E0, #CBCAC8)
- **Accents**: Muted colors (#5B9EFF, #ABABAB)
- **Borders**: Medium grays (#373737, #404040)
- **Hover States**: Lighter dark backgrounds (#3A3A3A)

### 6.3 Theme Toggle

- Located in sidebar header or settings
- Persisted in localStorage
- Sun/Moon icon toggle button
- System preference detection on first load
- Smooth transition between themes

## 7. Feed & Group Management

Feed and Group management is **contextual** rather than centralized in settings.

### 7.1 Create Group

**Trigger**:
- "Create Group" button at feed list title

**Interaction**:
1. Click button opens Dialog
2. Enter group name in input field
3. Confirm to create, cancel to dismiss

**UI Components**:
- Dialog with title "Create New Group"
- Text input for group name
- Cancel and Create buttons

### 7.2 Add Feed to Group

**Trigger Methods**:
- Hover over group header → "Add Feed" button appears
- Click button opens "Add Feed" dialog

**Interaction**:
1. Enter feed URL in input field
2. Optionally select target group (defaults to current group)
3. Confirm to add feed

**UI Components**:
- Dialog with title "Add Feed"
- URL input field
- Group selector (optional)
- Cancel and Add buttons

### 7.3 Access Methods for Feed/Group Actions

| Method        | Trigger                      | Device      |
| ------------- | ---------------------------- | ----------- |
| Context menu  | Right-click on feed/group    | Desktop     |
| Action button | Hover → Click three-dot icon | All devices |

### 7.4 Feed Actions (Context Menu / Dropdown)

- Mark all as read
- Rename
- Move to group
- Edit (URL, refresh interval, proxy settings)
- Delete (with confirmation)

### 7.5 Group Actions (Context Menu / Dropdown)

- Mark all as read
- Rename
- Delete (with confirmation, cannot delete "All" group)

### 7.6 Additional Actions

**Global Feed Actions** (future):
- Import OPML (in settings or toolbar)
- Export OPML (in settings or toolbar)

## 8. Settings

Settings presented as a **Dialog** (modal window), triggered from sidebar footer settings button.

### 8.1 Current Implementation

**Status**: Placeholder

Display a simple message or basic structure indicating settings are under development.

**Example UI**:
- Dialog title: "Settings"
- Placeholder message: "Settings coming soon..."
- Or basic tab structure: Appearance / Account / About

### 8.2 Future Sections (Planned)

**Appearance**:
- Theme selection (Light/Dark/System)
- Font size adjustment
- Density options (Compact/Comfortable)

**Data**:
- Import OPML
- Export OPML
- Clear cache/data

**Account**:
- User information
- Logout button

**About**:
- App version
- License information
- Links to documentation/support

## 9. Component Library

Using shadcn/ui components:

| Component      | Usage                                         |
| -------------- | --------------------------------------------- |
| Button         | Actions, navigation, quick actions            |
| Input          | Forms, search, feed URL input                 |
| Tabs           | Article list filtering (All/Unread/Bookmarked)|
| Dialog         | Settings, create group, add feed, search      |
| **Sheet**      | **Article detail drawer (slide from right)**  |
| DropdownMenu   | Feed/group action menus                       |
| ContextMenu    | Right-click menus for feed/group              |
| ScrollArea     | Scrollable lists (sidebar, article list)      |
| Separator      | Visual dividers                               |
| Command        | Global search (inside Dialog)                 |
| Sonner (Toast) | Notifications (success, error messages)       |
| Badge          | Unread count badges on feeds                  |
| Avatar         | Feed logos/icons                              |
| Collapsible    | Collapsible feed groups in sidebar            |
| Tooltip        | Hover tooltips for icon buttons               |

## 10. State Management

Using Zustand for global state:

```typescript
interface AppState {
  // Sidebar
  sidebarOpen: boolean;

  // Data
  groups: Group[];
  feeds: Feed[];

  // UI State (optional, can also be derived from URL)
  // Most UI state is managed via URL parameters

  // Actions
  toggleSidebar: () => void;
  setGroups: (groups: Group[]) => void;
  setFeeds: (feeds: Feed[]) => void;
  updateFeedUnreadCount: (feedId: number, delta: number) => void;
  addGroup: (group: Group) => void;
  removeGroup: (groupId: number) => void;
  addFeed: (feed: Feed) => void;
  removeFeed: (feedId: number) => void;
}
```

**State Management Philosophy**:

- **URL-first**: UI state (tab, feed filter, selected item, article drawer) managed via URL search parameters
- **Zustand for data**: Global data (groups, feeds) and sidebar state stored in Zustand
- **Server sync**: Data synced with backend API
- **Local updates**: Optimistic updates for better UX, rollback on error

**URL Parameters** (primary UI state):
- `tab`: Current filter (all/unread/bookmarked)
- `feed`: Selected feed ID
- `group`: Selected group ID
- `item`: Currently viewed article ID (opens drawer when present)

**Benefits**:
- Shareable URLs with exact app state
- Browser back/forward navigation works naturally
- Deep linking to specific articles
- Simpler state management (no complex global UI state)
