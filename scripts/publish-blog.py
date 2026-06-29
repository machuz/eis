#!/usr/bin/env python3
"""
Publish blog posts to dev.to and Hatena Blog.

Usage:
    # Publish a specific file
    python scripts/publish-blog.py docs/blog-en-devto-ch1.md

    # Publish all changed files (used by CI)
    python scripts/publish-blog.py --changed

    # Initialize mapping by fetching existing article IDs
    python scripts/publish-blog.py --init

    # dev.to maintenance (dev.to has no delete API; unpublish to de-dup)
    python scripts/publish-blog.py --devto-list                      # dump articles (id/state/title)
    python scripts/publish-blog.py --devto-unpublish <id> [...]      # unpublish (published=false)
    python scripts/publish-blog.py --devto-publish <id> [...]        # re-publish (published=true)

    # Hatena maintenance
    python scripts/publish-blog.py --hatena-list                     # dump entries (title/id/url)
    python scripts/publish-blog.py --hatena-delete <edit_id> [...]   # delete entries by edit id
    python scripts/publish-blog.py --set-draft <file> [...]          # withdraw entry to draft
    python scripts/publish-blog.py --set-published <file> [...] [--published 2026-07-01T09:00:00+09:00]

Environment variables:
    DEVTO_API_KEY       - dev.to API key
    HATENA_USER_ID      - Hatena user ID (e.g. ma2k8)
    HATENA_BLOG_ID      - Hatena blog ID (e.g. ma2k8.hateblo.jp)
    HATENA_API_KEY      - Hatena Blog API key
"""

import json
import os
import re
import subprocess
import sys
import time
import urllib.request
import urllib.error
from base64 import b64encode
from pathlib import Path
from xml.etree import ElementTree as ET

REPO_ROOT = Path(__file__).resolve().parent.parent
MAPPING_FILE = REPO_ROOT / "docs" / ".blog-mapping.json"
DOCS_DIR = REPO_ROOT / "docs"

ATOM_NS = "http://www.w3.org/2005/Atom"
APP_NS = "http://www.w3.org/2007/app"

# --- Mapping ---

def load_mapping():
    if MAPPING_FILE.exists():
        return json.loads(MAPPING_FILE.read_text())
    return {}


def save_mapping(mapping):
    MAPPING_FILE.write_text(json.dumps(mapping, indent=2, ensure_ascii=False) + "\n")


# --- dev.to ---

def devto_headers():
    api_key = os.environ.get("DEVTO_API_KEY", "")
    if not api_key:
        raise RuntimeError("DEVTO_API_KEY not set")
    return {
        "api-key": api_key,
        "Content-Type": "application/json",
        "Accept": "application/json",
        "User-Agent": "EIS-Blog-Publisher/1.0",
    }


def devto_publish(filepath: Path, mapping: dict) -> dict:
    """Publish or update a dev.to article."""
    content = filepath.read_text()
    filename = filepath.name
    article_id = mapping.get(filename, {}).get("devto_id")

    if article_id:
        # Update existing article
        url = f"https://dev.to/api/articles/{article_id}"
        data = json.dumps({"article": {"body_markdown": content}}).encode()
        req = urllib.request.Request(url, data=data, headers=devto_headers(), method="PUT")
        print(f"  Updating dev.to article {article_id}...")
    else:
        # Create new article
        url = "https://dev.to/api/articles"
        data = json.dumps({"article": {"body_markdown": content}}).encode()
        req = urllib.request.Request(url, data=data, headers=devto_headers(), method="POST")
        print(f"  Creating new dev.to article...")

    # Retry with exponential backoff on 429 (rate-limit) and 5xx.
    max_attempts = 6
    for attempt in range(max_attempts):
        try:
            with urllib.request.urlopen(req) as resp:
                result = json.loads(resp.read())
                new_id = result["id"]
                article_url = result["url"]
                print(f"  OK: {article_url}")
                return {"devto_id": new_id, "devto_url": article_url}
        except urllib.error.HTTPError as e:
            body = e.read().decode()
            retriable = e.code == 429 or 500 <= e.code < 600
            if retriable and attempt < max_attempts - 1:
                retry_after = int(e.headers.get("Retry-After", 0) or 0)
                wait = retry_after if retry_after > 0 else 2 ** (attempt + 1)
                print(f"  {e.code} received; sleeping {wait}s then retrying (attempt {attempt + 2}/{max_attempts})")
                time.sleep(wait)
                req = urllib.request.Request(url, data=data, headers=devto_headers(), method=req.get_method())
                continue
            print(f"  ERROR ({e.code}): {body}", file=sys.stderr)
            raise


def devto_fetch_articles():
    """Fetch all published articles from dev.to to build mapping."""
    url = "https://dev.to/api/articles/me/published?per_page=100"
    req = urllib.request.Request(url, headers=devto_headers())
    with urllib.request.urlopen(req) as resp:
        return json.loads(resp.read())


def devto_list_all():
    """Fetch every article (published + drafts) so ids/state can be inspected.

    dev.to has no delete endpoint, so de-duping is done by unpublishing
    (published=false), which moves an article into this list as a draft.
    """
    out = []
    for status in ("published", "unpublished"):
        url = f"https://dev.to/api/articles/me/{status}?per_page=100"
        req = urllib.request.Request(url, headers=devto_headers())
        with urllib.request.urlopen(req) as resp:
            for a in json.loads(resp.read()):
                out.append({
                    "id": a["id"],
                    "state": status,
                    "published_at": (a.get("published_at") or "")[:10],
                    "title": a["title"],
                    "url": a.get("url", ""),
                })
    return out


def devto_set_published(article_id: str, published: bool) -> dict:
    """Publish (draft -> live) an existing dev.to article by id (no body change).

    NOTE: dev.to has no DELETE endpoint, and its API IGNORES published=false on an
    already-published article (verified: the post stays live). Unpublishing or
    deleting therefore has to be done by hand in the dev.to dashboard. published=true
    (draft -> live) does work, which is what the 7/1 launch uses.
    """
    if not published:
        print("  WARN: dev.to ignores published=false on a live post; unpublish in the dashboard.", file=sys.stderr)
    url = f"https://dev.to/api/articles/{article_id}"
    data = json.dumps({"article": {"published": published}}).encode()
    req = urllib.request.Request(url, data=data, headers=devto_headers(), method="PUT")
    action = "Publishing" if published else "Unpublishing"
    print(f"  {action} dev.to article {article_id}...")
    for attempt in range(6):
        try:
            with urllib.request.urlopen(req) as resp:
                result = json.loads(resp.read())
                print(f"  OK: published={result.get('published')} {result.get('url')}")
                return result
        except urllib.error.HTTPError as e:
            if (e.code == 429 or 500 <= e.code < 600) and attempt < 5:
                wait = int(e.headers.get("Retry-After", 0) or 0) or 2 ** (attempt + 1)
                print(f"  {e.code}; retry in {wait}s")
                time.sleep(wait)
                req = urllib.request.Request(url, data=data, headers=devto_headers(), method="PUT")
                continue
            print(f"  ERROR ({e.code}): {e.read().decode()}", file=sys.stderr)
            raise


# --- Hatena Blog ---

def hatena_auth_header():
    user_id = os.environ.get("HATENA_USER_ID", "")
    api_key = os.environ.get("HATENA_API_KEY", "")
    if not user_id or not api_key:
        raise RuntimeError("HATENA_USER_ID and HATENA_API_KEY must be set")
    credentials = b64encode(f"{user_id}:{api_key}".encode()).decode()
    return f"Basic {credentials}"


def hatena_base_url():
    user_id = os.environ.get("HATENA_USER_ID", "")
    blog_id = os.environ.get("HATENA_BLOG_ID", "")
    if not user_id or not blog_id:
        raise RuntimeError("HATENA_USER_ID and HATENA_BLOG_ID must be set")
    return f"https://blog.hatena.ne.jp/{user_id}/{blog_id}/atom/entry"


def hatena_build_xml(title: str, body: str, categories: list[str] = None, draft: bool = False, published: str = None) -> bytes:
    """Build AtomPub XML for Hatena Blog.

    published: optional RFC3339 datetime (e.g. "2026-07-01T09:00:00+09:00"). When
    set, Hatena dates the entry at that time; a future value schedules it (予約投稿).
    """
    entry = ET.Element("entry", xmlns=ATOM_NS)
    entry.set("xmlns:app", APP_NS)

    ET.SubElement(entry, "title").text = title

    if published:
        ET.SubElement(entry, "published").text = published

    content = ET.SubElement(entry, "content", type="text/x-markdown")
    content.text = body

    if categories:
        for cat in categories:
            ET.SubElement(entry, "category", term=cat)

    control = ET.SubElement(entry, f"{{{APP_NS}}}control")
    ET.SubElement(control, f"{{{APP_NS}}}draft").text = "yes" if draft else "no"

    return b'<?xml version="1.0" encoding="utf-8"?>\n' + ET.tostring(entry, encoding="unicode").encode("utf-8")


def hatena_parse_title_and_body(filepath: Path) -> tuple[str, str]:
    """Extract title from H1 heading and return (title, body_without_h1)."""
    content = filepath.read_text()
    lines = content.split("\n")

    # First line should be "# title"
    if lines and lines[0].startswith("# "):
        title = lines[0][2:].strip()
        # Remove the H1 and any following blank line
        body_lines = lines[1:]
        while body_lines and body_lines[0].strip() == "":
            body_lines = body_lines[1:]
        body = "\n".join(body_lines)
    else:
        title = filepath.stem
        body = content

    return title, body


def hatena_categories_for(filename: str) -> list[str]:
    """Return Hatena categories appropriate for the given filename."""
    if "psyos" in filename:
        return ["心理OS", "エンジニアリング", "組織論"]
    if "structure" in filename:
        return ["構造駆動", "エンジニアリング", "組織論"]
    # default: git考古学 series
    return ["git考古学", "engineering-impact-score", "エンジニアリング"]


def hatena_publish(filepath: Path, mapping: dict, draft: bool = False, published: str = None) -> dict:
    """Publish or update a Hatena Blog entry.

    draft=True withdraws the entry to a draft (下書き); draft=False (re)publishes it.
    published optionally pins the entry's publish datetime (see hatena_build_xml).
    """
    title, body = hatena_parse_title_and_body(filepath)
    filename = filepath.name
    entry_id = mapping.get(filename, {}).get("hatena_id")

    categories = hatena_categories_for(filename)
    xml_data = hatena_build_xml(title, body, categories=categories, draft=draft, published=published)

    headers = {
        "Authorization": hatena_auth_header(),
        "Content-Type": "application/xml; charset=utf-8",
    }

    if entry_id:
        url = f"{hatena_base_url()}/{entry_id}"
        req = urllib.request.Request(url, data=xml_data, headers=headers, method="PUT")
        print(f"  Updating Hatena entry {entry_id}...")
    else:
        url = hatena_base_url()
        req = urllib.request.Request(url, data=xml_data, headers=headers, method="POST")
        print(f"  Creating new Hatena entry...")

    try:
        with urllib.request.urlopen(req) as resp:
            response_xml = resp.read()
            root = ET.fromstring(response_xml)

            # Extract entry ID from <link rel="edit" href=".../{entry_id}"/>
            ns = {"atom": ATOM_NS}
            edit_link = root.find('.//atom:link[@rel="edit"]', ns)
            new_id = edit_link.get("href").rstrip("/").split("/")[-1] if edit_link is not None else entry_id

            # Extract URL
            alt_link = root.find('.//atom:link[@rel="alternate"]', ns)
            entry_url = alt_link.get("href") if alt_link is not None else ""

            print(f"  OK: {entry_url}")
            return {"hatena_id": new_id, "hatena_url": entry_url}
    except urllib.error.HTTPError as e:
        body = e.read().decode()
        print(f"  ERROR ({e.code}): {body}", file=sys.stderr)
        raise


def hatena_fetch_entries():
    """Fetch all entries from Hatena Blog to build mapping."""
    headers = {
        "Authorization": hatena_auth_header(),
    }
    url = hatena_base_url()
    entries = []

    while url:
        req = urllib.request.Request(url, headers=headers)
        with urllib.request.urlopen(req) as resp:
            root = ET.fromstring(resp.read())

        ns = {"atom": ATOM_NS}
        for entry in root.findall("atom:entry", ns):
            title_el = entry.find("atom:title", ns)
            edit_link = entry.find('.//atom:link[@rel="edit"]', ns)
            alt_link = entry.find('.//atom:link[@rel="alternate"]', ns)

            if title_el is not None and edit_link is not None:
                entry_id = edit_link.get("href").rstrip("/").split("/")[-1]
                entries.append({
                    "title": title_el.text,
                    "id": entry_id,
                    "url": alt_link.get("href") if alt_link is not None else "",
                })

        # Pagination: look for <link rel="next">
        next_link = root.find('.//atom:link[@rel="next"]', ns)
        url = next_link.get("href") if next_link is not None else None

    return entries


def hatena_delete(entry_id: str) -> None:
    """Delete a Hatena Blog entry by its AtomPub edit id."""
    headers = {"Authorization": hatena_auth_header()}
    url = f"{hatena_base_url()}/{entry_id}"
    req = urllib.request.Request(url, headers=headers, method="DELETE")
    print(f"  Deleting Hatena entry {entry_id}...")
    try:
        with urllib.request.urlopen(req) as resp:
            print(f"  OK: HTTP {resp.status}")
    except urllib.error.HTTPError as e:
        print(f"  ERROR ({e.code}): {e.read().decode()}", file=sys.stderr)
        raise


# --- File detection ---

def detect_platform(filepath: Path) -> str:
    """Detect target platform from filename."""
    name = filepath.name
    if "devto" in name:
        return "devto"
    elif "hatena" in name:
        return "hatena"
    elif "blog-ja-note-" in name or "blog-ja-x-" in name:
        # note (note.com) and X (twitter/x.com) are manually published; skip CI automation.
        return "skip"
    else:
        raise ValueError(f"Cannot detect platform from filename: {name}")


def get_changed_blog_files() -> list[Path]:
    """Get blog files changed in the latest commit."""
    result = subprocess.run(
        ["git", "diff", "--name-only", "HEAD~1", "HEAD", "--", "docs/blog-*.md"],
        capture_output=True, text=True, cwd=REPO_ROOT,
    )
    files = []
    for line in result.stdout.strip().split("\n"):
        if line:
            p = REPO_ROOT / line
            if p.exists():
                files.append(p)
    return files


# --- Init mode ---

def init_mapping():
    """Fetch existing articles from both platforms and build mapping."""
    mapping = load_mapping()
    blog_files = sorted(DOCS_DIR.glob("blog-*.md"))

    # Build title -> filename index for matching
    title_to_file = {}
    for f in blog_files:
        content = f.read_text()
        lines = content.split("\n")
        # Extract title
        if f.name.startswith("blog-en-devto"):
            # dev.to: title from frontmatter
            for line in lines:
                if line.startswith("title:"):
                    title = line.split(":", 1)[1].strip().strip('"').strip("'")
                    title_to_file[title.lower()] = f.name
                    break
        elif f.name.startswith("blog-ja-hatena"):
            # Hatena: title from H1
            if lines[0].startswith("# "):
                title = lines[0][2:].strip()
                title_to_file[title.lower()] = f.name

        # Also extract chapter number from filename for fallback matching
        ch_match = re.search(r"ch(\d+)", f.name)
        if ch_match:
            title_to_file[f"#{ ch_match.group(1) }"] = f.name

    # dev.to
    if os.environ.get("DEVTO_API_KEY"):
        print("Fetching dev.to articles...")
        try:
            articles = devto_fetch_articles()
            for article in articles:
                title = article["title"].lower()
                for key, filename in title_to_file.items():
                    if key in title or title in key:
                        if filename not in mapping:
                            mapping[filename] = {}
                        mapping[filename]["devto_id"] = article["id"]
                        mapping[filename]["devto_url"] = article["url"]
                        print(f"  Matched: {filename} -> {article['id']}")
                        break
        except Exception as e:
            print(f"  dev.to fetch failed: {e}", file=sys.stderr)

    # Hatena
    if os.environ.get("HATENA_API_KEY"):
        print("Fetching Hatena entries...")
        try:
            entries = hatena_fetch_entries()
            for entry in entries:
                title = entry["title"].lower()
                for key, filename in title_to_file.items():
                    if key in title or title in key:
                        if filename not in mapping:
                            mapping[filename] = {}
                        mapping[filename]["hatena_id"] = entry["id"]
                        mapping[filename]["hatena_url"] = entry["url"]
                        print(f"  Matched: {filename} -> {entry['id']}")
                        break
        except Exception as e:
            print(f"  Hatena fetch failed: {e}", file=sys.stderr)

    save_mapping(mapping)
    print(f"\nMapping saved to {MAPPING_FILE}")
    print(json.dumps(mapping, indent=2, ensure_ascii=False))


# --- Main ---

def publish_file(filepath: Path):
    """Publish a single blog file to its target platform."""
    mapping = load_mapping()
    platform = detect_platform(filepath)
    filename = filepath.name

    if platform == "skip":
        print(f"Skipping {filename} (manual-publish target)")
        return

    print(f"Publishing {filename} to {platform}...")

    if platform == "devto":
        result = devto_publish(filepath, mapping)
        if filename not in mapping:
            mapping[filename] = {}
        mapping[filename].update(result)
    elif platform == "hatena":
        result = hatena_publish(filepath, mapping)
        if filename not in mapping:
            mapping[filename] = {}
        mapping[filename].update(result)

    save_mapping(mapping)


def set_hatena_state(filepath: Path, draft: bool, published: str = None):
    """Re-PUT an existing Hatena entry to toggle its draft state (withdraw/republish).

    Requires the file to already be in the mapping (an existing hatena_id); refuses
    to create a new entry, so it can never accidentally duplicate a post.
    """
    mapping = load_mapping()
    filename = filepath.name
    if detect_platform(filepath) != "hatena":
        print(f"  {filename} is not a Hatena target; skipping", file=sys.stderr)
        return
    if not mapping.get(filename, {}).get("hatena_id"):
        print(f"  No hatena_id mapped for {filename}; refusing to create a new entry", file=sys.stderr)
        sys.exit(1)
    state = "draft (withdraw)" if draft else "published"
    print(f"Setting {filename} to {state}...")
    result = hatena_publish(filepath, mapping, draft=draft, published=published)
    mapping[filename].update(result)
    save_mapping(mapping)


def main():
    args = sys.argv[1:]

    if not args:
        print(__doc__)
        sys.exit(1)

    if args[0] == "--init":
        init_mapping()
        return

    if args[0] == "--devto-list":
        print(json.dumps(devto_list_all(), indent=2, ensure_ascii=False))
        return

    if args[0] in ("--devto-unpublish", "--devto-publish"):
        published = args[0] == "--devto-publish"
        for article_id in args[1:]:
            devto_set_published(article_id, published)
        return

    if args[0] == "--hatena-list":
        # Dump every Hatena entry (title/id/url) so edit-ids can be discovered.
        entries = hatena_fetch_entries()
        print(json.dumps(entries, indent=2, ensure_ascii=False))
        print(f"\n{len(entries)} entries.")
        return

    if args[0] == "--hatena-delete":
        # Delete one or more Hatena entries by AtomPub edit id.
        for entry_id in args[1:]:
            hatena_delete(entry_id)
        return

    if args[0] in ("--set-draft", "--set-published"):
        # Withdraw to / republish from draft. Remaining args are blog md paths.
        draft = args[0] == "--set-draft"
        published = None
        rest = args[1:]
        # Optional "--published <RFC3339>" pins the publish datetime on republish.
        if "--published" in rest:
            i = rest.index("--published")
            published = rest[i + 1]
            rest = rest[:i] + rest[i + 2:]
        for arg in rest:
            filepath = Path(arg)
            if not filepath.is_absolute():
                filepath = Path.cwd() / filepath
            if not filepath.exists():
                print(f"File not found: {filepath}", file=sys.stderr)
                continue
            try:
                set_hatena_state(filepath, draft=draft, published=published)
            except Exception as e:
                print(f"Failed to set state for {filepath.name}: {e}", file=sys.stderr)
        return

    if args[0] == "--changed":
        files = get_changed_blog_files()
        if not files:
            print("No blog files changed.")
            return
        for f in files:
            try:
                publish_file(f)
            except Exception as e:
                print(f"Failed to publish {f.name}: {e}", file=sys.stderr)
        return

    # Publish specific files
    for arg in args:
        filepath = Path(arg)
        if not filepath.is_absolute():
            filepath = Path.cwd() / filepath
        if not filepath.exists():
            print(f"File not found: {filepath}", file=sys.stderr)
            continue
        try:
            publish_file(filepath)
        except Exception as e:
            print(f"Failed to publish {filepath.name}: {e}", file=sys.stderr)


if __name__ == "__main__":
    main()
