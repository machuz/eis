#!/usr/bin/env python3
"""
Generate an HTML visualization of a project's team evolution over time.
Reads a timeline JSON (from eis timeline command) and outputs a self-contained HTML file.

Usage:
    python generate-evolution-html.py <timeline.json> <output.html> [--title "Project Name"]
"""

import json
import sys
import os
import html
from collections import defaultdict

# ---------------------------------------------------------------------------
# 1. Load data
# ---------------------------------------------------------------------------

def load_timeline(path: str) -> dict:
    """Load timeline JSON, handling trailing non-JSON lines."""
    raw = open(path, "r", encoding="utf-8").read()
    # Find the end of the first complete JSON object
    depth = 0
    in_string = False
    escape = False
    end = 0
    for i, ch in enumerate(raw):
        if escape:
            escape = False
            continue
        if ch == '\\' and in_string:
            escape = True
            continue
        if ch == '"' and not escape:
            in_string = not in_string
            continue
        if in_string:
            continue
        if ch == '{':
            depth += 1
        elif ch == '}':
            depth -= 1
            if depth == 0:
                end = i + 1
                break
    return json.loads(raw[:end])


# ---------------------------------------------------------------------------
# 2. Extract analytics
# ---------------------------------------------------------------------------

ROLES_ORDERED = ["Architect", "Anchor", "Producer", "Cleaner", "Specialist", "Other"]
ROLE_COLORS = {
    "Architect": "#f97583",
    "Anchor": "#79c0ff",
    "Producer": "#56d364",
    "Cleaner": "#d2a8ff",
    "Specialist": "#ffa657",
    "Other": "#484f58",
}

def normalize_role(r: str) -> str:
    if r in ("Architect", "Anchor", "Producer", "Cleaner", "Specialist"):
        return r
    return "Other"


# ---------------------------------------------------------------------------
# i18n — EIS taxonomy words (Architect/Anchor/gravity/peak/NEW) stay in English
# for consistency with the rest of the OSS Gravity Map pages; only descriptive
# prose is localized.
# ---------------------------------------------------------------------------

STRINGS = {
    "en": {
        "html_lang": "en",
        "title_suffix": "Generational Evolution",
        "subtitle": "How structural gravity transferred across {n} years",
        "lang_label": "日本語",
        "hm_title": "Generational Handover &mdash; who held a structural seat, year by year",
        "hm_sub": "Each row is one engineer who ever held a named role · colour = the role they held that year · shade by gravity",
        "tbl_name": "Name",
        "departed": "Departed: ",
        "spark_peak": "peak {v}",
        "t_architects": "Architects (structural leaders)",
        "t_anchors": "Anchors (gravity holders)",
        "t_producers": "Producers (output engines) — top 12 by peak gravity",
        "t_cleaners": "Cleaners (entropy fighters) — top 12 by peak gravity",
        "s1_title": "Role Composition Timeline",
        "s1_desc": "Active engineers per role in each period. Shows how the team's role balance evolved.",
        "s2_title": "Generational Handover",
        "s2_desc": "Every engineer who ever held a named structural seat — <strong>Architect</strong>, <strong>Anchor</strong>, <strong>Cleaner</strong>, <strong>Specialist</strong> — as one row, year by year. Ordered by when they first appear, so the baton passing between generations reads as a staircase: watch a seat empty in one era and someone new take it up in the next.",
        "s3_title": "Who Held Each Role, Year by Year",
        "s3_desc": "Each cell is that engineer's gravity in that year; blank = they did not hold the role then. Rows step in by first appearance, so you can read the handover from one generation to the next — who held the seat, and who took it over.",
        "s4_title": "Gravity Leadership Board",
        "s4_desc": 'Top 5 by gravity each period. <span style="color:#238636;font-weight:600">NEW</span> = just entered top 5. Departed names listed below.',
        "s5_title": "Individual Trajectories",
        "s5_desc": "Gravity over time for key figures (anyone ever in top 10). Dot color = role. Hover for details.",
    },
    "ja": {
        "html_lang": "ja",
        "title_suffix": "世代的進化",
        "subtitle": "構造的な重力が{n}年でどう移り変わったか",
        "lang_label": "EN",
        "hm_title": "世代交代 &mdash; 誰が構造的な席を、年ごとに担ったか",
        "hm_sub": "各行は名前付きロールを一度でも担ったエンジニア・色＝その年のロール・濃淡＝gravity",
        "tbl_name": "名前",
        "departed": "離脱: ",
        "spark_peak": "peak {v}",
        "t_architects": "Architect（構造リーダー）",
        "t_anchors": "Anchor（gravity の保持者）",
        "t_producers": "Producer（出力エンジン）— peak gravity 上位12",
        "t_cleaners": "Cleaner（entropy と戦う者）— peak gravity 上位12",
        "s1_title": "ロール構成の時系列",
        "s1_desc": "各期間のロール別アクティブエンジニア数。チームのロールバランスがどう移り変わったかを示す。",
        "s2_title": "世代交代",
        "s2_desc": "名前付きの構造的な席 — <strong>Architect</strong>・<strong>Anchor</strong>・<strong>Cleaner</strong>・<strong>Specialist</strong> — を一度でも持った全エンジニアを1行ずつ、年ごとに並べる。初出順に並ぶので、世代から世代へのバトンパスが階段状に読める。ある時代に席が空き、次の時代に新しい人がそれを引き継ぐ様子が見える。",
        "s3_title": "誰がどのロールを、年ごとに担ったか",
        "s3_desc": "各セルはその年のそのエンジニアの gravity。空欄＝その年はそのロールを担っていない。行は初出順に並ぶので、世代から次の世代への引き継ぎが読める — 誰が席を持ち、誰が引き継いだか。",
        "s4_title": "Gravity リーダーボード",
        "s4_desc": '各期間の gravity 上位5。<span style="color:#238636;font-weight:600">NEW</span> ＝上位5に新規参入。離脱者は下に記載。',
        "s5_title": "個人の軌跡",
        "s5_desc": "主要人物（一度でも上位10入り）の gravity の推移。点の色＝ロール。ホバーで詳細。",
    },
}


def extract_data(data: dict):
    periods = [p["label"] for p in data["periods"]]
    authors = data["authors"]

    # --- role composition per period ---
    role_counts = []  # list of dicts {role: count}
    for pi, plabel in enumerate(periods):
        counts = defaultdict(int)
        for a in authors:
            ap = a["periods"][pi]
            if ap["gravity"] > 0 or ap["commits"] > 0:
                counts[normalize_role(ap["role"])] += 1
        role_counts.append(dict(counts))

    # --- architects per period ---
    architects_per_period = []
    for pi, plabel in enumerate(periods):
        archs = []
        for a in authors:
            ap = a["periods"][pi]
            if ap["role"] == "Architect":
                archs.append((a["author"], ap["gravity"], ap["commits"]))
        archs.sort(key=lambda x: -x[1])
        architects_per_period.append(archs)

    # --- anchors per period ---
    anchors_per_period = []
    for pi, plabel in enumerate(periods):
        ancs = []
        for a in authors:
            ap = a["periods"][pi]
            if ap["role"] == "Anchor":
                ancs.append((a["author"], ap["gravity"], ap["commits"]))
        ancs.sort(key=lambda x: -x[1])
        anchors_per_period.append(ancs)

    # --- producers per period ---
    producers_per_period = []
    for pi, plabel in enumerate(periods):
        ps = []
        for a in authors:
            ap = a["periods"][pi]
            if ap["role"] == "Producer":
                ps.append((a["author"], ap["gravity"], ap["commits"]))
        ps.sort(key=lambda x: -x[1])
        producers_per_period.append(ps)

    # --- cleaners per period ---
    cleaners_per_period = []
    for pi, plabel in enumerate(periods):
        cs = []
        for a in authors:
            ap = a["periods"][pi]
            if ap["role"] == "Cleaner":
                cs.append((a["author"], ap["gravity"], ap["commits"]))
        cs.sort(key=lambda x: -x[1])
        cleaners_per_period.append(cs)

    # --- top 5 by gravity per period ---
    top5_per_period = []
    for pi, plabel in enumerate(periods):
        entries = []
        for a in authors:
            ap = a["periods"][pi]
            if ap["gravity"] > 0:
                entries.append((a["author"], ap["gravity"], normalize_role(ap["role"]), ap["commits"]))
        entries.sort(key=lambda x: -x[1])
        top5_per_period.append(entries[:5])

    # --- key figures: anyone ever in top 10 ---
    ever_top10 = set()
    for pi, plabel in enumerate(periods):
        entries = []
        for a in authors:
            ap = a["periods"][pi]
            if ap["gravity"] > 0:
                entries.append(a["author"])
        # sort by gravity
        scored = [(a["author"], a["periods"][pi]["gravity"]) for a in authors if a["periods"][pi]["gravity"] > 0]
        scored.sort(key=lambda x: -x[1])
        for name, _ in scored[:10]:
            ever_top10.add(name)

    key_figures = []
    for a in authors:
        if a["author"] in ever_top10:
            trajectory = []
            for pi, plabel in enumerate(periods):
                ap = a["periods"][pi]
                trajectory.append({
                    "label": plabel,
                    "gravity": ap["gravity"],
                    "role": normalize_role(ap["role"]),
                    "commits": ap["commits"],
                })
            # peak gravity
            peak = max(t["gravity"] for t in trajectory)
            key_figures.append({
                "name": a["author"],
                "trajectory": trajectory,
                "peak": peak,
            })
    key_figures.sort(key=lambda x: -x["peak"])

    # --- structural-role tenure per author (generational handover heatmap) ---
    # For everyone who EVER held a named structural role (not Producer/Other),
    # record which role they held each period. Sorting by first appearance makes
    # the baton-pass between generations read as a staircase.
    structural = ("Architect", "Anchor", "Cleaner", "Specialist")
    tenure = {}
    for a in authors:
        row, first, peak, held = [], None, 0.0, False
        for pi in range(len(periods)):
            ap = a["periods"][pi]
            r = ap["role"] if ap["role"] in structural else None
            if r:
                held = True
                if first is None:
                    first = pi
                peak = max(peak, ap["gravity"])
                row.append({"role": r, "gravity": ap["gravity"]})
            else:
                row.append(None)
        if held:
            tenure[a["author"]] = {"row": row, "first": first, "peak": peak}
    structural_tenure = sorted(
        tenure.items(), key=lambda kv: (kv[1]["first"], -kv[1]["peak"])
    )

    return {
        "periods": periods,
        "role_counts": role_counts,
        "architects_per_period": architects_per_period,
        "anchors_per_period": anchors_per_period,
        "producers_per_period": producers_per_period,
        "cleaners_per_period": cleaners_per_period,
        "top5_per_period": top5_per_period,
        "key_figures": key_figures,
        "structural_tenure": structural_tenure,
    }


# ---------------------------------------------------------------------------
# 3. HTML generation
# ---------------------------------------------------------------------------

def esc(s):
    return html.escape(str(s))

def fmt_gravity(g):
    if g >= 1000:
        return f"{g/1000:.1f}k"
    return f"{g:.0f}"


def generate_stacked_bar_svg(periods, role_counts, width=900, height=340):
    """Generate stacked bar chart SVG for role composition."""
    n = len(periods)
    bar_w = min(60, (width - 80) // n - 8)
    gap = 8
    total_bar_area = n * (bar_w + gap) - gap
    x_offset = (width - total_bar_area) // 2

    # find max total per period
    max_total = max(sum(rc.get(r, 0) for r in ROLES_ORDERED) for rc in role_counts)
    if max_total == 0:
        max_total = 1
    chart_h = height - 80
    y_top = 30

    lines = [f'<svg viewBox="0 0 {width} {height}" xmlns="http://www.w3.org/2000/svg" style="width:100%;max-width:{width}px">']

    # bars
    for i, (plabel, rc) in enumerate(zip(periods, role_counts)):
        bx = x_offset + i * (bar_w + gap)
        by = y_top
        total = sum(rc.get(r, 0) for r in ROLES_ORDERED)

        # draw stacked from bottom
        cy = y_top + chart_h
        for role in ROLES_ORDERED:
            count = rc.get(role, 0)
            if count == 0:
                continue
            seg_h = (count / max_total) * chart_h
            cy -= seg_h
            color = ROLE_COLORS[role]
            lines.append(f'  <rect x="{bx}" y="{cy:.1f}" width="{bar_w}" height="{seg_h:.1f}" fill="{color}" rx="2">'
                         f'<title>{role}: {count}</title></rect>')

        # total label on top
        top_y = y_top + chart_h - (total / max_total) * chart_h - 6
        lines.append(f'  <text x="{bx + bar_w/2}" y="{top_y:.1f}" text-anchor="middle" fill="#8b949e" font-size="11">{total}</text>')

        # period label
        lines.append(f'  <text x="{bx + bar_w/2}" y="{y_top + chart_h + 18}" text-anchor="middle" fill="#8b949e" font-size="12" font-weight="600">{plabel}</text>')

    # legend
    lx = x_offset
    ly = height - 14
    for role in ROLES_ORDERED:
        color = ROLE_COLORS[role]
        lines.append(f'  <rect x="{lx}" y="{ly - 9}" width="10" height="10" fill="{color}" rx="2"/>')
        lines.append(f'  <text x="{lx + 14}" y="{ly}" fill="#8b949e" font-size="11">{role}</text>')
        lx += len(role) * 7 + 30

    lines.append('</svg>')
    return '\n'.join(lines)


def generate_handover_heatmap_svg(periods, structural_tenure, width=900, max_rows=24, lang="en"):
    """Generational-handover heatmap: one row per engineer who ever held a named
    structural role (Architect / Anchor / Cleaner / Specialist), columns are
    years, each cell coloured by the role they held that year (blank if none).
    Rows are ordered by first appearance, so the baton passing from one
    generation to the next reads as a staircase down the page — and you can see
    who held the Anchor/Architect seat in each era and who took it over."""
    rows = structural_tenure[:max_rows]
    n = len(periods)
    label_w = 168
    pad = 18
    cell_w = max(18, (width - label_w - pad) // n)
    cell_h = 20
    top = 58
    height = top + len(rows) * cell_h + 30

    # opacity scales with gravity against the whole panel's peak, with a floor
    # so a low-gravity Cleaner seat is still legible.
    gpeak = max((c["gravity"] for _, meta in rows for c in meta["row"] if c), default=1) or 1

    S = STRINGS[lang]
    lines = [f'<svg viewBox="0 0 {width} {height}" xmlns="http://www.w3.org/2000/svg" style="width:100%;max-width:{width}px">']
    lines.append(f'<text x="12" y="22" fill="#f0f6fc" font-size="14" font-weight="600">{S["hm_title"]}</text>')
    lines.append(f'<text x="12" y="40" fill="#8b949e" font-size="11">{S["hm_sub"]}</text>')

    for ri, (name, meta) in enumerate(rows):
        ry = top + ri * cell_h
        nm = name if len(name) <= 24 else name[:23] + "…"
        lines.append(f'<text x="{label_w - 8}" y="{ry + cell_h/2 + 4:.0f}" text-anchor="end" fill="#c9d1d9" font-size="11">{esc(nm)}</text>')
        for ci, cell in enumerate(meta["row"]):
            cx = label_w + ci * cell_w
            if cell is None:
                lines.append(f'  <rect x="{cx}" y="{ry}" width="{cell_w-2}" height="{cell_h-2}" rx="2" fill="#161b22"/>')
                continue
            color = ROLE_COLORS.get(cell["role"], "#484f58")
            op = 0.40 + 0.60 * (cell["gravity"] / gpeak)
            lines.append(f'  <rect x="{cx}" y="{ry}" width="{cell_w-2}" height="{cell_h-2}" rx="2" fill="{color}" fill-opacity="{op:.2f}">'
                         f'<title>{esc(name)} · {esc(str(periods[ci]))}: {cell["role"]} (gravity {cell["gravity"]:.0f})</title></rect>')

    # year labels
    yl = top + len(rows) * cell_h + 14
    step = max(1, n // 16)  # avoid label crowding on long histories
    for ci, p in enumerate(periods):
        if ci % step != 0 and ci != n - 1:
            continue
        cx = label_w + ci * cell_w + (cell_w - 2) / 2
        lines.append(f'  <text x="{cx:.0f}" y="{yl}" text-anchor="middle" fill="#8b949e" font-size="10">{esc(str(p))}</text>')

    # legend
    lx = 12
    ly = height - 6
    for role in ["Architect", "Anchor", "Cleaner", "Specialist"]:
        color = ROLE_COLORS[role]
        lines.append(f'  <rect x="{lx}" y="{ly - 9}" width="10" height="10" rx="2" fill="{color}"/>')
        lines.append(f'  <text x="{lx + 14}" y="{ly}" fill="#8b949e" font-size="11">{role}</text>')
        lx += len(role) * 7 + 28

    lines.append('</svg>')
    return '\n'.join(lines)


def generate_sparkline_svg(trajectory, w=260, h=40):
    """Small gravity sparkline for a key figure."""
    gravities = [t["gravity"] for t in trajectory]
    max_g = max(gravities) if max(gravities) > 0 else 1
    n = len(gravities)
    pad_x = 2
    pad_y = 4
    usable_w = w - 2 * pad_x
    usable_h = h - 2 * pad_y

    points = []
    dots = []
    for i, g in enumerate(gravities):
        x = pad_x + (i / (n - 1)) * usable_w if n > 1 else pad_x + usable_w / 2
        y = pad_y + usable_h - (g / max_g) * usable_h
        points.append(f"{x:.1f},{y:.1f}")
        role = trajectory[i]["role"]
        color = ROLE_COLORS.get(role, "#484f58")
        dots.append(f'<circle cx="{x:.1f}" cy="{y:.1f}" r="3" fill="{color}"><title>{trajectory[i]["label"]}: {fmt_gravity(g)} ({role})</title></circle>')

    svg = [f'<svg viewBox="0 0 {w} {h}" xmlns="http://www.w3.org/2000/svg" style="width:{w}px;height:{h}px">']
    if any(g > 0 for g in gravities):
        svg.append(f'  <polyline points="{" ".join(points)}" fill="none" stroke="#30363d" stroke-width="1.5"/>')
        svg.extend(f'  {d}' for d in dots)
    svg.append('</svg>')
    return '\n'.join(svg)


def generate_html(analytics, project_title="React", lang="en", en_href="", ja_href=""):
    S = STRINGS[lang]
    periods = analytics["periods"]
    role_counts = analytics["role_counts"]
    architects = analytics["architects_per_period"]
    anchors = analytics["anchors_per_period"]
    producers = analytics["producers_per_period"]
    cleaners = analytics["cleaners_per_period"]
    top5 = analytics["top5_per_period"]
    key_figures = analytics["key_figures"]

    bar_svg = generate_stacked_bar_svg(periods, role_counts)
    heatmap_svg = generate_handover_heatmap_svg(periods, analytics["structural_tenure"], lang=lang)

    # language toggle (rendered only when both counterpart hrefs are known)
    if en_href and ja_href:
        def _link(href, label, active):
            cls = ' class="active"' if active else ''
            return f'<a href="{esc(href)}"{cls}>{label}</a>'
        lang_switch = (
            '<div class="lang-switch">'
            + _link(en_href, "EN", lang == "en")
            + '<span class="sep">·</span>'
            + _link(ja_href, "日本語", lang == "ja")
            + '</div>'
        )
    else:
        lang_switch = ""

    # Order names so the generational handover reads top-to-bottom: pick the
    # strongest holders by peak gravity (cap keeps populous roles legible), then
    # sort the chosen names by first appearance so each new generation steps in
    # below the one before.
    def sorted_role_names(per_period, cap=None):
        meta = {}  # name -> [first_period, peak_gravity]
        for pi, entries in enumerate(per_period):
            for n, g, c in entries:
                m = meta.setdefault(n, [pi, 0.0])
                m[0] = min(m[0], pi)
                m[1] = max(m[1], g)
        names = list(meta)
        if cap is not None and len(names) > cap:
            names = sorted(names, key=lambda n: -meta[n][1])[:cap]
        return sorted(names, key=lambda n: (meta[n][0], -meta[n][1]))

    sorted_architect_names = sorted_role_names(architects)
    sorted_anchor_names = sorted_role_names(anchors)
    sorted_producer_names = sorted_role_names(producers, cap=12)
    sorted_cleaner_names = sorted_role_names(cleaners, cap=12)

    def role_table_html(title, role_name, sorted_names, per_period_data):
        rows = []
        for name in sorted_names:
            cells = []
            for pi, entries in enumerate(per_period_data):
                match = None
                for n, g, c in entries:
                    if n == name:
                        match = (g, c)
                        break
                if match:
                    g, c = match
                    cells.append(f'<td class="cell-active" title="{esc(name)} — gravity {g:.0f}, {c} commits">{fmt_gravity(g)}</td>')
                else:
                    cells.append('<td class="cell-empty"></td>')
            rows.append(f'<tr><td class="name-cell">{esc(name)}</td>{"".join(cells)}</tr>')

        header_cells = ''.join(f'<th>{esc(p)}</th>' for p in periods)
        return f'''
        <h3 class="section-subtitle">{esc(title)}</h3>
        <div class="table-scroll">
        <table class="role-table">
          <thead><tr><th>{esc(S["tbl_name"])}</th>{header_cells}</tr></thead>
          <tbody>{"".join(rows)}</tbody>
        </table>
        </div>'''

    architects_html = role_table_html(
        S["t_architects"],
        "Architect", sorted_architect_names, architects)
    anchors_html = role_table_html(
        S["t_anchors"],
        "Anchor", sorted_anchor_names, anchors)
    producers_html = role_table_html(
        S["t_producers"],
        "Producer", sorted_producer_names, producers)
    cleaners_html = role_table_html(
        S["t_cleaners"],
        "Cleaner", sorted_cleaner_names, cleaners)

    # --- top 5 leaderboard ---
    leaderboard_rows = []
    # track who was in previous top5
    prev_names = set()
    for pi, entries in enumerate(top5):
        cur_names = set(e[0] for e in entries)
        new_names = cur_names - prev_names
        gone_names = prev_names - cur_names

        period_cards = []
        for rank, (name, grav, role, commits) in enumerate(entries, 1):
            is_new = name in new_names and pi > 0
            role_color = ROLE_COLORS.get(role, "#484f58")
            new_badge = ' <span class="badge-new">NEW</span>' if is_new else ''
            period_cards.append(
                f'<div class="lb-card">'
                f'<span class="lb-rank">#{rank}</span>'
                f'<span class="lb-name" style="color:{role_color}">{esc(name)}{new_badge}</span>'
                f'<span class="lb-grav">{fmt_gravity(grav)}</span>'
                f'<span class="lb-role-tag" style="border-color:{role_color};color:{role_color}">{esc(role)}</span>'
                f'</div>'
            )

        departed = ""
        if gone_names and pi > 0:
            dep_list = ", ".join(sorted(gone_names))
            departed = f'<div class="lb-departed">{esc(S["departed"])}{esc(dep_list)}</div>'

        leaderboard_rows.append(
            f'<div class="lb-period">'
            f'<div class="lb-period-label">{esc(periods[pi])}</div>'
            f'{"".join(period_cards)}'
            f'{departed}'
            f'</div>'
        )
        prev_names = cur_names

    # --- key figures sparklines ---
    sparkline_cards = []
    for kf in key_figures[:30]:  # limit to top 30
        spark_svg = generate_sparkline_svg(kf["trajectory"])
        # find active periods
        active = [t for t in kf["trajectory"] if t["gravity"] > 0]
        if not active:
            continue
        first_year = active[0]["label"]
        last_year = active[-1]["label"]
        peak_role = max(kf["trajectory"], key=lambda t: t["gravity"])["role"]
        role_color = ROLE_COLORS.get(peak_role, "#484f58")

        sparkline_cards.append(
            f'<div class="spark-card">'
            f'<div class="spark-header">'
            f'<span class="spark-name" style="color:{role_color}">{esc(kf["name"])}</span>'
            f'<span class="spark-peak">{S["spark_peak"].format(v=fmt_gravity(kf["peak"]))}</span>'
            f'</div>'
            f'<div class="spark-meta">{first_year}–{last_year}</div>'
            f'{spark_svg}'
            f'</div>'
        )

    return f'''<!DOCTYPE html>
<html lang="{S["html_lang"]}">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{esc(project_title)} — {esc(S["title_suffix"])}</title>
<style>
  *,*::before,*::after{{box-sizing:border-box;margin:0;padding:0}}
  body{{
    background:#0d1117;color:#c9d1d9;
    font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;
    line-height:1.6;padding:0 16px 80px;
  }}
  ::selection{{background:#388bfd44;color:#fff}}

  header{{max-width:960px;margin:0 auto;padding:60px 0 24px;text-align:center}}
  h1{{font-size:clamp(1.6rem,4vw,2.6rem);font-weight:700;color:#f0f6fc;letter-spacing:-.02em;margin-bottom:8px}}
  h1 span{{color:#58a6ff}}
  .subtitle{{font-size:clamp(.85rem,1.8vw,1.05rem);color:#8b949e;margin-bottom:0}}

  .container{{max-width:960px;margin:0 auto}}

  .section{{margin-top:56px}}
  .section-title{{
    font-size:1.3rem;font-weight:700;color:#f0f6fc;margin-bottom:8px;
    padding-bottom:8px;border-bottom:1px solid #21262d;
  }}
  .section-desc{{font-size:.9rem;color:#8b949e;margin-bottom:20px}}
  .section-subtitle{{
    font-size:1.05rem;font-weight:600;color:#c9d1d9;margin:28px 0 12px;
  }}

  /* Stacked bar chart */
  .chart-wrap{{
    background:#161b22;border:1px solid #30363d;border-radius:12px;
    padding:24px;overflow-x:auto;
  }}

  /* Role tables */
  .table-scroll{{overflow-x:auto;margin-bottom:8px}}
  .role-table{{
    border-collapse:collapse;width:100%;font-size:.85rem;
  }}
  .role-table th{{
    padding:8px 10px;text-align:center;color:#8b949e;font-weight:600;
    border-bottom:1px solid #30363d;white-space:nowrap;
  }}
  .role-table th:first-child{{text-align:left;min-width:180px}}
  .role-table td{{padding:6px 10px;text-align:center;border-bottom:1px solid #161b22}}
  .role-table .name-cell{{
    text-align:left;font-weight:500;color:#e6edf3;white-space:nowrap;
    max-width:200px;overflow:hidden;text-overflow:ellipsis;
  }}
  .cell-active{{
    background:#1f6feb22;color:#79c0ff;font-weight:600;font-variant-numeric:tabular-nums;
  }}
  .cell-empty{{color:#21262d}}

  /* Leaderboard */
  .lb-grid{{display:grid;grid-template-columns:repeat(auto-fill,minmax(260px,1fr));gap:16px}}
  .lb-period{{
    background:#161b22;border:1px solid #30363d;border-radius:10px;padding:16px;
  }}
  .lb-period-label{{
    font-size:1.1rem;font-weight:700;color:#f0f6fc;margin-bottom:12px;
    padding-bottom:8px;border-bottom:1px solid #21262d;
  }}
  .lb-card{{
    display:flex;align-items:center;gap:8px;padding:5px 0;
  }}
  .lb-rank{{font-size:.85rem;font-weight:700;color:#484f58;width:28px;font-variant-numeric:tabular-nums}}
  .lb-name{{font-weight:600;font-size:.9rem;flex:1;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}}
  .lb-grav{{font-size:.8rem;color:#8b949e;font-variant-numeric:tabular-nums;margin-right:4px}}
  .lb-role-tag{{
    font-size:.65rem;font-weight:600;padding:1px 6px;border:1px solid;border-radius:8px;
    text-transform:uppercase;letter-spacing:.03em;white-space:nowrap;
  }}
  .badge-new{{
    display:inline-block;font-size:.6rem;font-weight:700;
    background:#238636;color:#fff;padding:1px 5px;border-radius:6px;
    margin-left:4px;vertical-align:middle;
  }}
  .lb-departed{{
    margin-top:8px;padding-top:8px;border-top:1px solid #21262d;
    font-size:.75rem;color:#f8514966;font-style:italic;
  }}

  /* Sparklines */
  .spark-grid{{display:grid;grid-template-columns:repeat(auto-fill,minmax(280px,1fr));gap:12px}}
  .spark-card{{
    background:#161b22;border:1px solid #30363d;border-radius:10px;padding:14px 16px;
  }}
  .spark-header{{display:flex;justify-content:space-between;align-items:baseline}}
  .spark-name{{font-weight:600;font-size:.95rem}}
  .spark-peak{{font-size:.75rem;color:#8b949e;font-variant-numeric:tabular-nums}}
  .spark-meta{{font-size:.72rem;color:#484f58;margin:2px 0 8px}}

  /* Language switch */
  .lang-switch{{
    position:absolute;top:18px;right:20px;font-size:.82rem;
  }}
  .lang-switch a{{color:#8b949e;text-decoration:none;padding:2px 4px}}
  .lang-switch a.active{{color:#58a6ff;font-weight:700}}
  .lang-switch a:hover{{color:#c9d1d9}}
  .lang-switch .sep{{color:#30363d;margin:0 2px}}
</style>
</head>
<body>

{lang_switch}
<header>
  <h1><span>{esc(project_title)}</span> — {esc(S["title_suffix"])}</h1>
  <p class="subtitle">{esc(S["subtitle"].format(n=len(periods)))}</p>
</header>

<div class="container">

  <!-- Section 1: Role Composition -->
  <div class="section">
    <h2 class="section-title">{esc(S["s1_title"])}</h2>
    <p class="section-desc">{S["s1_desc"]}</p>
    <div class="chart-wrap">
      {bar_svg}
    </div>
  </div>

  <!-- Section 1b: Generational Handover Heatmap -->
  <div class="section">
    <h2 class="section-title">{esc(S["s2_title"])}</h2>
    <p class="section-desc">{S["s2_desc"]}</p>
    <div class="chart-wrap">
      {heatmap_svg}
    </div>
  </div>

  <!-- Section 2: Architects & Anchors -->
  <div class="section">
    <h2 class="section-title">{esc(S["s3_title"])}</h2>
    <p class="section-desc">{S["s3_desc"]}</p>
    {architects_html}
    {anchors_html}
    {producers_html}
    {cleaners_html}
  </div>

  <!-- Section 3: Gravity Leadership Board -->
  <div class="section">
    <h2 class="section-title">{esc(S["s4_title"])}</h2>
    <p class="section-desc">{S["s4_desc"]}</p>
    <div class="lb-grid">
      {"".join(leaderboard_rows)}
    </div>
  </div>

  <!-- Section 4: Individual Trajectories -->
  <div class="section">
    <h2 class="section-title">{esc(S["s5_title"])}</h2>
    <p class="section-desc">{S["s5_desc"]}</p>
    <div class="spark-grid">
      {"".join(sparkline_cards)}
    </div>
  </div>

</div>

</body>
</html>'''


# ---------------------------------------------------------------------------
# 4. Main
# ---------------------------------------------------------------------------

def main():
    if len(sys.argv) < 3:
        print(f"Usage: {sys.argv[0]} <timeline.json> <output.html> [--title 'Name']")
        sys.exit(1)

    json_path = sys.argv[1]
    output_path = sys.argv[2]
    title = "React"
    if "--title" in sys.argv:
        idx = sys.argv.index("--title")
        if idx + 1 < len(sys.argv):
            title = sys.argv[idx + 1]
    lang = "en"
    if "--lang" in sys.argv:
        idx = sys.argv.index("--lang")
        if idx + 1 < len(sys.argv):
            lang = sys.argv[idx + 1]
    if lang not in STRINGS:
        print(f"Unknown --lang '{lang}'; use one of {list(STRINGS)}")
        sys.exit(1)

    # counterpart hrefs for the language toggle, derived from the output basename:
    # foo.html <-> foo-ja.html (same directory, relative link)
    base = os.path.basename(output_path)
    if base.endswith("-ja.html"):
        en_base, ja_base = base[:-8] + ".html", base
    elif base.endswith(".html"):
        en_base, ja_base = base, base[:-5] + "-ja.html"
    else:
        en_base = ja_base = base

    print(f"Loading {json_path}...")
    data = load_timeline(json_path)
    print(f"  {len(data['authors'])} authors, {len(data['periods'])} periods")

    print("Extracting analytics...")
    analytics = extract_data(data)
    print(f"  Key figures: {len(analytics['key_figures'])}")

    print(f"Generating HTML ({lang}) -> {output_path}")
    html_content = generate_html(analytics, project_title=title, lang=lang,
                                 en_href=en_base, ja_href=ja_base)

    os.makedirs(os.path.dirname(output_path) or ".", exist_ok=True)
    with open(output_path, "w", encoding="utf-8") as f:
        f.write(html_content)
    print(f"Done. {len(html_content):,} bytes written.")


if __name__ == "__main__":
    main()
