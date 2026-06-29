#!/usr/bin/env python3
"""Generate the OrbitLens Ace blog cover SVGs and render them to PNG.

The covers reproduce the hand-made terminal-card rasters, swapping only the
right-side logo for the current gold orbit-ring Ace mark (logo-ace-mark.svg,
itself a copy of lp/logo-ace.svg). Layout/copy/colours were measured pixel-wise
from the original PNGs; text widths are pinned with textLength so the rendering
is font-independent (rsvg falls back to a system monospace).
"""
import os
import re
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(HERE)
IMG = os.path.join(ROOT, "docs", "images")
BLOG = os.path.join(IMG, "blog")

MONO = "'JetBrains Mono','SFMono-Regular',Menlo,Consolas,'Liberation Mono',monospace"

# gruvbox palette sampled from the originals
BG = "#1d2021"
GRAY = "#928374"      # comment
GOLD = "#fabd2f"      # title
TEAL = "#83a598"      # observe/interpret lines
MAGENTA = "#d3869b"   # section heading
DGRAY = "#7c6f64"     # $ lines + topbar
ORANGE = "#fe8019"    # url / arrow
RED = "#fb4934"
GREEN = "#b8bb26"


def logo_fragment():
    """Return the inner markup of logo-ace-mark.svg wrapped in a nested <svg>
    positioned to match the original cover (centre 812,232; r 114)."""
    with open(os.path.join(IMG, "logo-ace-mark.svg"), encoding="utf-8") as f:
        svg = f.read()
    inner = re.sub(r"^.*?<svg[^>]*>", "", svg, count=1, flags=re.S)
    inner = re.sub(r"</svg>\s*$", "", inner, flags=re.S)
    return (
        '  <svg x="698" y="118" width="228" height="228" '
        'viewBox="48 6 340 340">\n' + inner + "  </svg>\n"
    )


def line(x, y, text, fill, size, length, anchor="start", weight="normal"):
    return (
        f'  <text x="{x}" y="{y}" fill="{fill}" font-size="{size}" '
        f'font-family="{MONO}" font-weight="{weight}" '
        f'text-anchor="{anchor}" textLength="{length}" '
        f'lengthAdjust="spacingAndGlyphs" xml:space="preserve">{text}</text>\n'
    )


def chrome():
    """traffic-light dots + the shared bottom $ install line."""
    s = ""
    for cx, col in ((22, RED), (47, GOLD), (72, GREEN)):
        s += f'  <circle cx="{cx}" cy="20" r="7" fill="{col}"/>\n'
    s += line(45, 373, "$ brew install machuz/tap/eis", DGRAY, 15, 229)
    s += line(331, 373, "&#8594; ace.orbitlens.io", ORANGE, 15, 141)
    return s


def cover(topbar, comment, title, l1, l2, heading, d1, d2):
    body = (
        f'<svg xmlns="http://www.w3.org/2000/svg" width="1000" height="420" '
        f'viewBox="0 0 1000 420">\n'
        f'  <rect width="1000" height="420" fill="{BG}"/>\n'
    )
    body += chrome()
    body += line(500, 25, topbar, DGRAY, 14, len(topbar) * 8.4, anchor="middle")
    body += line(44, 92, comment, GRAY, 17, comment_w(comment))
    body += line(44, 150, title, GOLD, 46, title_w(title))
    body += line(44, 196, l1, TEAL, 20, l1_w)
    body += line(44, 224, l2, TEAL, 20, l2_w)
    body += line(44, 275, heading, MAGENTA, 15, heading_w)
    body += line(45, 305, d1, DGRAY, 15, d_w)
    body += line(45, 329, d2, DGRAY, 15, d_w)
    body += logo_fragment()
    body += "</svg>\n"
    return body


def render(svg_path, png_path, dims=(1000, 420)):
    try:
        subprocess.run(
            ["rsvg-convert", "-w", str(dims[0]), "-h", str(dims[1]),
             svg_path, "-o", png_path],
            check=True,
        )
    except FileNotFoundError:
        sys.stderr.write(
            "error: rsvg-convert not found on PATH. Install librsvg "
            "(e.g. `brew install librsvg` or `apt-get install librsvg2-bin`).\n"
        )
        sys.exit(1)


# per-cover measured widths are passed explicitly below
comment_w = lambda t: comment_w.v
title_w = lambda t: title_w.v


def build(name, **kw):
    global l1_w, l2_w, heading_w, d_w
    comment_w.v = kw.pop("comment_w")
    title_w.v = kw.pop("title_w")
    l1_w = kw.pop("l1_w")
    l2_w = kw.pop("l2_w")
    heading_w = kw.pop("heading_w")
    d_w = kw.pop("d_w")
    svg = cover(**kw)
    svg_path = os.path.join(BLOG, f"cover-ace-{name}.svg")
    with open(svg_path, "w", encoding="utf-8") as f:
        f.write(svg)
    for sub in ("png", "hatena"):
        os.makedirs(os.path.join(BLOG, sub), exist_ok=True)
        render(svg_path, os.path.join(BLOG, sub, f"cover-ace-{name}.png"))
    print("wrote", svg_path)


build(
    "launch",
    topbar="OrbitLens Ace &#8212; Launch",
    comment="// the telescope now has an observatory",
    comment_w=350,
    title="OrbitLens Ace",
    title_w=334,
    l1="Observe with the open telescope.",
    l1_w=347,
    l2="Interpret with the observatory.",
    l2_w=336,
    heading="What becomes observable",
    heading_w=182,
    d1="$ structural summary &#183; conway &#183; collapse risk",
    d2="$ organizational chronicle &#8212; a record, not a score",
    d_w=397,
)

build(
    "features",
    topbar="OrbitLens Ace &#8212; Feature Tour",
    comment="// a tour of the observatory",
    comment_w=251,
    title="Inside Ace",
    title_w=254,
    l1="Every screen answers one question:",
    l1_w=367,
    l2="what becomes observable?",
    l2_w=262,
    heading="The screens",
    heading_w=83,
    d1="$ dashboard &#183; star detail &#183; module topology",
    d2="$ chronicle &#183; ambient &#183; gravity certificate",
    d_w=341,
)
print("done")
