#!/usr/bin/env python3
"""
Re-order a dev.to series by unpublish → republish of each article.

Forem's series ordering is `ORDER BY COALESCE(crossposted_at, published_at) ASC`
and has no API or UI override. But toggling `published: false` then
`published: true` resets `published_at` to *now* while keeping the article's
slug, ID, comments, and reactions intact — so we can rewrite the order
without breaking any external links.

Usage:
    DEVTO_API_KEY=... python scripts/reorder-devto-series.py <file> [<file> ...]

The order of filenames on the command line becomes the new series order
(earliest → latest). Example:

    python scripts/reorder-devto-series.py \\
        docs/blog-en-devto-structure-ch0.md \\
        docs/blog-en-devto-structure-ch1.md \\
        ...
        docs/blog-en-devto-structure-ch10.md
"""

import json
import os
import sys
import time
import urllib.request
import urllib.error
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
MAPPING_FILE = REPO_ROOT / "docs" / ".blog-mapping.json"


def devto_headers():
    api_key = os.environ.get("DEVTO_API_KEY", "")
    if not api_key:
        raise RuntimeError("DEVTO_API_KEY not set")
    return {
        "api-key": api_key,
        "Content-Type": "application/json",
        "Accept": "application/json",
        "User-Agent": "EIS-Blog-Reorderer/1.0",
    }


def devto_put(article_id: int, payload: dict) -> dict:
    """PUT to dev.to article. Retries on 429 and 5xx with exponential backoff."""
    url = f"https://dev.to/api/articles/{article_id}"
    data = json.dumps({"article": payload}).encode()
    max_attempts = 6

    for attempt in range(max_attempts):
        req = urllib.request.Request(url, data=data, headers=devto_headers(), method="PUT")
        try:
            with urllib.request.urlopen(req) as resp:
                return json.loads(resp.read())
        except urllib.error.HTTPError as e:
            body = e.read().decode()
            retriable = e.code == 429 or 500 <= e.code < 600
            if retriable and attempt < max_attempts - 1:
                retry_after = 0
                try:
                    retry_after = int(e.headers.get("Retry-After", 0) or 0)
                except (TypeError, ValueError):
                    retry_after = 0
                wait = retry_after if retry_after > 0 else 2 ** (attempt + 1)
                print(f"  {e.code} received; sleeping {wait}s then retrying (attempt {attempt + 2}/{max_attempts})", file=sys.stderr)
                time.sleep(wait)
                continue
            raise RuntimeError(f"HTTP {e.code}: {body}") from e
    raise RuntimeError("exhausted retry attempts")


def main():
    if len(sys.argv) < 2:
        print(__doc__, file=sys.stderr)
        sys.exit(1)

    filenames = [Path(p).name for p in sys.argv[1:]]
    mapping = json.loads(MAPPING_FILE.read_text())

    targets: list[tuple[str, int]] = []
    for fn in filenames:
        entry = mapping.get(fn)
        if not entry or "devto_id" not in entry:
            raise RuntimeError(f"{fn}: no devto_id in {MAPPING_FILE}")
        targets.append((fn, entry["devto_id"]))

    print(f"Reordering {len(targets)} articles in this order:")
    for idx, (fn, aid) in enumerate(targets, 1):
        print(f"  {idx:>2}. {fn}  (id {aid})")
    print()

    for fn, aid in targets:
        print(f"[{fn}] unpublish {aid} ...")
        devto_put(aid, {"published": False})
        time.sleep(2)
        print(f"[{fn}] republish {aid} ...")
        devto_put(aid, {"published": True})
        # Spacing between articles to stay well under the rate limit.
        time.sleep(2)
        print(f"[{fn}] done.\n")

    print("All articles reordered.")


if __name__ == "__main__":
    main()
