#!/usr/bin/env python3
"""Seed bookmarks via the Fusion HTTP API for manual testing.

Collects item IDs by paginating GET /api/items, then creates a bookmark
for each via POST /api/bookmarks with {"item_id": <id>}. Re-running is
safe: items already bookmarked are skipped (the server rejects duplicate
links with HTTP 500 due to the UNIQUE constraint on bookmarks.link).

Examples:
    # 100 bookmarks, anonymous (no password configured on server)
    python3 scripts/seed_bookmarks.py

    # 50 bookmarks against a password-protected server
    python3 scripts/seed_bookmarks.py --password secret --count 50

Flags may also be set via env vars: FUSION_URL, FUSION_PASSWORD.
"""

import argparse
import http.cookiejar
import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request

DEFAULT_URL = "http://localhost:8080"
PAGE_SIZE = 100  # mirrors the server's maxListLimit


class FusionClient:
    """Thin HTTP client that persists the session cookie automatically."""

    def __init__(self, base_url):
        self.base_url = base_url.rstrip("/")
        self.cookiejar = http.cookiejar.CookieJar()
        self.opener = urllib.request.build_opener(
            urllib.request.HTTPCookieProcessor(self.cookiejar)
        )

    def request(self, method, path, body=None):
        url = self.base_url + path
        data = None
        headers = {"Accept": "application/json"}
        if body is not None:
            data = json.dumps(body).encode("utf-8")
            headers["Content-Type"] = "application/json"
        req = urllib.request.Request(url, data=data, headers=headers, method=method)
        try:
            with self.opener.open(req) as resp:
                raw = resp.read()
                return resp.status, (json.loads(raw) if raw else None)
        except urllib.error.HTTPError as e:
            raw = e.read().decode("utf-8", "replace")
            try:
                return e.code, json.loads(raw)
            except ValueError:
                return e.code, {"error": raw}
        except urllib.error.URLError as e:
            sys.exit(f"error: cannot reach {url}: {e.reason}")

    def login(self, password):
        status, body = self.request("POST", "/api/sessions", {"password": password})
        if status != 200:
            sys.exit(f"error: login failed (HTTP {status}): {body}")
        print("login: ok", file=sys.stderr)

    def list_items_page(self, limit, before=None):
        params = [("limit", str(limit))]
        if before:
            params.append(("before", before))
        status, body = self.request(
            "GET", "/api/items?" + urllib.parse.urlencode(params)
        )
        if status != 200:
            sys.exit(f"error: list items failed (HTTP {status}): {body}")
        return body

    def create_bookmark(self, item_id):
        return self.request("POST", "/api/bookmarks", {"item_id": item_id})


def collect_item_ids(client, count):
    """Page through GET /api/items until we have `count` IDs (or run out)."""
    ids = []
    cursor = None
    while len(ids) < count:
        page = client.list_items_page(PAGE_SIZE, before=cursor)
        data = page.get("data") or []
        if not data:
            break
        ids.extend(item["id"] for item in data)
        cursor = page.get("next_cursor")
        if not cursor:
            break
    return ids[:count]


def main():
    parser = argparse.ArgumentParser(description="Seed bookmarks for manual testing.")
    parser.add_argument(
        "--url",
        default=os.environ.get("FUSION_URL", DEFAULT_URL),
        help=f"Fusion base URL (default: {DEFAULT_URL} or $FUSION_URL)",
    )
    parser.add_argument(
        "--password",
        default=os.environ.get("FUSION_PASSWORD", ""),
        help="login password; omit for anonymous access (env: FUSION_PASSWORD)",
    )
    parser.add_argument(
        "--count",
        type=int,
        default=100,
        help="number of bookmarks to create (default: 100)",
    )
    parser.add_argument(
        "--sleep",
        type=float,
        default=0.0,
        help="seconds to wait between bookmark requests (default: 0)",
    )
    args = parser.parse_args()

    client = FusionClient(args.url)
    if args.password:
        client.login(args.password)

    ids = collect_item_ids(client, args.count)
    if not ids:
        print(
            "no items found; add some feeds and fetch items first.",
            file=sys.stderr,
        )
        return
    if len(ids) < args.count:
        print(
            f"warning: only {len(ids)} items available (requested {args.count}).",
            file=sys.stderr,
        )

    created = skipped = failed = 0
    for i, item_id in enumerate(ids, 1):
        status, _ = client.create_bookmark(item_id)
        if status == 200:
            created += 1
        elif status == 500:
            # UNIQUE(link) conflict => item is already bookmarked
            skipped += 1
        else:
            failed += 1
            print(f"  item {item_id}: unexpected HTTP {status}", file=sys.stderr)
        if i % 10 == 0 or i == len(ids):
            print(f"  progress: {i}/{len(ids)}", file=sys.stderr)
        if args.sleep:
            time.sleep(args.sleep)

    print(
        f"done: created={created}, skipped(already)={skipped}, failed={failed}",
        file=sys.stderr,
    )


if __name__ == "__main__":
    main()
