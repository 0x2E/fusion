# Google Reader API Compatibility

Fusion implements the Google Reader (GReader) sync protocol, so clients that
offer a "Google Reader" account type — ReadYou, NetNewsWire, Reeder, FeedMe,
FocusReader and others — can read, mark and manage subscriptions without a
Fusion-specific client. Fever stays available next to it; both are independent
front doors onto the same data. The route inventory lives in `SetupRouter` in
`backend/internal/handler/handler.go`.

## Why a second compatibility API

Fever covers reading and saved state but not subscription management. GReader
is the protocol mobile and desktop clients implement for a full sync,
including subscribing to and organizing feeds, which is what lets one of them
replace the web UI instead of complementing it.

The implementation mirrors Fever's structure and auth handling
(`backend/internal/handler/fever.go`): the same store, the same auth-mode
switch in `New` and the same login rate limiter.

## Client setup

- Account type: `Google Reader`. Clients without a generic Google Reader type,
  such as NetNewsWire, work with `FreshRSS`: enter the same base URL as its
  API URL.
- Server URL: the Fusion base URL, for example `https://rss.example.com` or
  `http://192.168.1.10:8080`. Do not append `/api`, `/reader` or
  `/accounts/ClientLogin`; the client composes those itself.
- Username: `FUSION_GREADER_USERNAME`. It is separate from the Fever username:
  the two protocols serve different clients, and changing one should not sign
  out the other.
- Password: `FUSION_PASSWORD`.

Defaults and the full variable list live in [`.env.example`](../.env.example);
`Load` in `backend/internal/config/config.go` is authoritative.

### Quick connectivity check

`POST /accounts/ClientLogin` with `Email` and `Passwd` returns the token lines
(`SID`, `LSID`, `Auth`) when the credentials are accepted:

```bash
curl -sS -X POST 'https://your-domain/accounts/ClientLogin' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data 'Email=fusion&Passwd=your_password'
```

Then reuse the `Auth=` value as the credential for every other request:

```bash
curl -sS 'https://your-domain/reader/api/0/user-info' \
  -H 'Authorization: GoogleLogin auth=<Auth value>'
```

## Authentication

Every GReader route requires `Authorization: GoogleLogin auth=<token>`. The
token is a stateless HMAC of the username keyed by the password — see
`deriveGReaderToken` in `backend/internal/handler/greader.go`. It survives a
restart and is invalidated by a password change, which matches how clients
cache credentials.

The same token is returned by `POST /accounts/ClientLogin` and by the token
endpoint. Clients send it in the `Authorization` header; a POST without that
header may carry it as the `T` form value instead, because some clients only
send `T` on writes. `T` is accepted as a credential, not checked as a CSRF
token on top of the header: a browser cannot attach that header cross-site, so
a second check would add no protection. `T` in a GET query string is not
accepted, since query strings end up in access logs.

Behavior depends on how Fusion is configured for authentication (see `New` in
`backend/internal/handler/handler.go`):

| Fusion mode | ClientLogin | Reader endpoints |
|---|---|---|
| Password configured | requires the configured username and password | token checked |
| Anonymous (no password, no OIDC) | accepts any credentials | token checked, like every mode; the token derives from the empty password, so any client can obtain it |
| OIDC only | always fails | always `401` |

ClientLogin is rate limited by the same limiter as the password login and
shares its budget per client IP; excess attempts get `429` with `Retry-After`.
Token checks are not rate limited, because the token is a 256-bit secret
rather than a guessable password.

## Mapping to Fusion

| GReader concept | Fusion |
|---|---|
| `user/-/state/com.google/reading-list` | all items |
| `feed/<id>` | one feed |
| `user/-/label/<name>` | one group; a feed belongs to exactly one group, so its single label is that group |
| `user/-/state/com.google/read` | `items.unread = 0` |
| `user/-/state/com.google/kept-unread` | unread items (`items.unread = 1`) |
| `user/-/state/com.google/starred` | a bookmark with a non-null `item_id` |
| item id `tag:google.com,2005:reader/item/<hex>` | `items.id`; input also accepts the bare 16-digit hex and the short decimal form (`parseGReaderItemID`) |

Stream order and every timestamp the protocol exposes (`timestampUsec`,
`crawlTimeMsec`, and the `ot`/`nt`/`ts` filters) use `items.created_at`
(crawl time), not `pub_date`. Clients sync incrementally with `nt`; using
publication time there would silently skip items whose `pub_date` is older
than their crawl time. `published`/`updated` inside the item JSON do use
`pub_date`, falling back to `created_at`.

Continuations are opaque to the client: Fusion encodes the last returned item
id. Do not parse or construct one.

## Known limitations

- One label per feed: GReader allows several labels per subscription, Fusion
  has a single group. Adding a label moves the feed.
- Subscription edits are idempotent: subscribing to a feed that is already
  known updates its title and label instead of failing, quickadd returns the
  existing subscription, and unsubscribing a feed that is gone is a no-op.
  This keeps retries from clients harmless.
- Starred streams expose only bookmarks that still point at a live item;
  bookmarks orphaned by feed deletion remain visible in Fusion's own bookmarks
  API but not through GReader.
- Items have no author field in Fusion, so `author` is empty.
- Tags on individual items are only understood for read and starred state
  (`edit-tag`, `it`, `xt`). `broadcast` and other tags are ignored, as are
  the protocol's social endpoints, which Fusion does not implement.
- Renaming or deleting a label operates on the group: feeds stay subscribed
  and move to `Default` when their group is deleted.

## Development notes

- Handler and client-flow tests: `backend/internal/handler/greader_test.go`.
- Store queries and tests: `backend/internal/store/greader.go` and
  `greader_test.go`.
- The GReader routes are registered outside `/api`, next to Fever. An
  unimplemented endpoint under `/reader/api/0/` answers an authenticated GET or
  POST with `200 []`: clients probe endpoints such as `preference/list` during
  setup and treat a `404` as a broken server. Other unregistered paths under
  `/reader` and `/accounts` return `404` instead of the SPA shell (`NoRoute`
  in `SetupRouter`, `isBackendPath` in `backend/internal/handler/frontend.go`).
