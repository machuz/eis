#!/usr/bin/env python3
"""Sync dev.to article URLs into the Library landing page (docs/index.html).

Background
----------
The Library landing page lists each article twice: an English variant
(``data-lang="en"``) that should point at the published dev.to article, and a
Japanese variant (``data-lang="ja"``) that points at the internal page. The
dev.to URL only exists *after* CI publishes the article (the slug carries a
random suffix), so it cannot be hand-written when the entry is authored.

This script closes that gap. Author the English ``<a>`` with a
``data-devto-src="<md filename>"`` attribute and any placeholder ``href``; this
script rewrites the ``href`` to the real dev.to URL recorded in
``docs/.blog-mapping.json``. It runs in CI right after publishing, so new
entries get their links filled automatically — no manual follow-up.

Usage
-----
    python scripts/sync-library-links.py [--check]

``--check`` exits non-zero if any href would change (for CI verification);
without it, the file is rewritten in place.
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
MAPPING_FILE = REPO_ROOT / "docs" / ".blog-mapping.json"
INDEX_FILE = REPO_ROOT / "docs" / "index.html"

SRC_RE = re.compile(r'data-devto-src="([^"]+)"')
HREF_RE = re.compile(r'href="[^"]*"')


def main() -> int:
    check_only = "--check" in sys.argv[1:]

    mapping = json.loads(MAPPING_FILE.read_text())
    lines = INDEX_FILE.read_text().splitlines(keepends=True)

    changed: list[str] = []
    missing: list[str] = []

    for i, line in enumerate(lines):
        m = SRC_RE.search(line)
        if not m:
            continue
        src = m.group(1)
        url = (mapping.get(src) or {}).get("devto_url")
        if not url:
            missing.append(src)
            continue
        new_line, n = HREF_RE.subn(f'href="{url}"', line, count=1)
        if n and new_line != line:
            lines[i] = new_line
            changed.append(f"{src} -> {url}")

    for src in missing:
        print(f"WARN: no devto_url in mapping for {src} (not published yet?)")

    if not changed:
        print("Library links already in sync.")
        return 0

    if check_only:
        print("Out-of-sync Library links (run without --check to fix):")
        for c in changed:
            print(f"  {c}")
        return 1

    INDEX_FILE.write_text("".join(lines))
    print(f"Updated {len(changed)} Library link(s):")
    for c in changed:
        print(f"  {c}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
