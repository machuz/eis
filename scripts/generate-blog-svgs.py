#!/usr/bin/env python3
"""Generate Gruvbox-themed terminal SVG images for blog posts."""

import os
import html

# Gruvbox Dark palette
COLORS = {
    'bg': '#282828',
    'bg_dark': '#1d2021',
    'fg': '#ebdbb2',
    'fg_dim': '#928374',
    'red': '#fb4934',
    'green': '#b8bb26',
    'yellow': '#fabd2f',
    'blue': '#83a598',
    'purple': '#d3869b',
    'aqua': '#8ec07c',
    'orange': '#fe8019',
    'separator': '#504945',
}

FONT_SIZE = 12
LINE_HEIGHT = 18
CHAR_WIDTH = 7.2  # approximate for monospace at 12px
TITLE_BAR_HEIGHT = 32
PADDING_TOP = 50
PADDING_LEFT = 20
PADDING_BOTTOM = 20


def escape(text):
    return html.escape(text)


class TerminalSVG:
    """Builder for Gruvbox terminal SVG."""

    def __init__(self, title="Terminal", width=1000):
        self.title = title
        self.width = width
        self.lines = []  # list of (y_offset, elements)
        self.y = PADDING_TOP

    def add_blank(self, count=1):
        self.y += LINE_HEIGHT * count

    def add_text(self, text, color='fg', bold=False, x=None):
        """Add a single line of text."""
        if x is None:
            x = PADDING_LEFT
        weight = ' font-weight="700"' if bold else ''
        self.lines.append(
            f'  <text x="{x}" y="{self.y}" fill="{COLORS[color]}" font-size="{FONT_SIZE}"{weight}>{escape(text)}</text>'
        )
        self.y += LINE_HEIGHT

    def add_colored_spans(self, spans, y_override=None):
        """Add a line with mixed colors: [(text, color, bold), ...]"""
        y = y_override if y_override else self.y
        x = PADDING_LEFT
        parts = []
        for span in spans:
            text, color, bold = span if len(span) == 3 else (span[0], span[1], False)
            weight = ' font-weight="700"' if bold else ''
            parts.append(
                f'<tspan x="{x}" fill="{COLORS[color]}"{weight}>{escape(text)}</tspan>'
            )
            x += len(text) * CHAR_WIDTH
        self.lines.append(f'  <text y="{y}" font-size="{FONT_SIZE}">{"".join(parts)}</text>')
        if y_override is None:
            self.y += LINE_HEIGHT

    def add_separator(self, x1=None, x2=None):
        if x1 is None:
            x1 = PADDING_LEFT
        if x2 is None:
            x2 = self.width - PADDING_LEFT
        # Place line just below the previous text baseline (y was already incremented)
        line_y = self.y - LINE_HEIGHT + 5
        self.lines.append(
            f'  <line x1="{x1}" y1="{line_y}" x2="{x2}" y2="{line_y}" stroke="{COLORS["separator"]}" stroke-width="1"/>'
        )
        self.y += 4  # small gap after separator

    def add_table_row(self, cells, col_widths):
        """cells: [(text, color, bold), ...], col_widths: [int, ...]"""
        x = PADDING_LEFT
        parts = []
        for i, cell in enumerate(cells):
            text, color, bold = cell if len(cell) == 3 else (cell[0], cell[1], False)
            weight = ' font-weight="700"' if bold else ''
            parts.append(
                f'<tspan x="{x}" fill="{COLORS[color]}"{weight}>{escape(text)}</tspan>'
            )
            x += col_widths[i] if i < len(col_widths) else 60
        self.lines.append(f'  <text y="{self.y}" font-size="{FONT_SIZE}">{"".join(parts)}</text>')
        self.y += LINE_HEIGHT

    def add_command(self, cmd, prompt="❯"):
        """Add a command prompt line."""
        self.add_colored_spans([
            (prompt + " ", 'green'),
            (cmd, 'fg'),
        ])

    def render(self):
        height = self.y + PADDING_BOTTOM
        svg_parts = [
            f'<svg xmlns="http://www.w3.org/2000/svg" width="{self.width}" height="{height}" viewBox="0 0 {self.width} {height}">',
            '  <defs>',
            "    <style>",
            "      @import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&amp;display=swap');",
            "      text { font-family: 'JetBrains Mono', 'SF Mono', 'Menlo', monospace; }",
            '    </style>',
            '  </defs>',
            '',
            f'  <!-- Terminal background -->',
            f'  <rect width="{self.width}" height="{height}" rx="10" fill="{COLORS["bg"]}"/>',
            '',
            f'  <!-- Title bar -->',
            f'  <rect width="{self.width}" height="{TITLE_BAR_HEIGHT}" rx="10" fill="{COLORS["bg_dark"]}"/>',
            f'  <rect y="22" width="{self.width}" height="10" fill="{COLORS["bg_dark"]}"/>',
            f'  <circle cx="20" cy="16" r="6" fill="#cc241d"/>',
            f'  <circle cx="40" cy="16" r="6" fill="#d79921"/>',
            f'  <circle cx="60" cy="16" r="6" fill="#98971a"/>',
            f'  <text x="{self.width // 2}" y="20" text-anchor="middle" fill="{COLORS["fg_dim"]}" font-size="12">{escape(self.title)}</text>',
            '',
        ]
        svg_parts.extend(self.lines)
        svg_parts.append('</svg>')
        return '\n'.join(svg_parts)

    def save(self, path):
        with open(path, 'w') as f:
            f.write(self.render())
        print(f"  Generated: {path}")


def make_score_color(val, thresholds=None):
    """Return color based on score value."""
    if thresholds is None:
        thresholds = {'high': 80, 'mid': 60, 'senior': 40, 'low': 20}
    try:
        v = float(val)
    except (ValueError, TypeError):
        return 'fg_dim'
    if v >= thresholds['high']:
        return 'purple'
    elif v >= thresholds['mid']:
        return 'green'
    elif v >= thresholds['senior']:
        return 'yellow'
    elif v >= thresholds['low']:
        return 'fg'
    else:
        return 'fg_dim'


def total_color(val):
    try:
        v = float(val)
    except (ValueError, TypeError):
        return 'fg_dim'
    if v >= 80:
        return 'purple'
    elif v >= 60:
        return 'green'
    elif v >= 40:
        return 'yellow'
    else:
        return 'fg'


def role_color(role):
    if 'Architect' in role:
        return 'purple'
    elif 'Anchor' in role:
        return 'blue'
    elif 'Producer' in role:
        return 'yellow'
    elif 'Cleaner' in role:
        return 'aqua'
    return 'fg_dim'


def state_color(state):
    if 'Active' in state:
        return 'green'
    elif 'Growing' in state:
        return 'blue'
    elif 'Former' in state:
        return 'fg_dim'
    elif 'Fragile' in state:
        return 'red'
    elif 'Silent' in state:
        return 'fg_dim'
    return 'fg'


# ─── IMAGE OUTPUT DIRECTORY ───
IMG_DIR = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), 'docs', 'images', 'blog')
os.makedirs(IMG_DIR, exist_ok=True)


# ════════════════════════════════════════════════════════════
# CHAPTER 1 — Individual Scoring
# ════════════════════════════════════════════════════════════

def ch1_backend_table():
    svg = TerminalSVG("Terminal — eis analyze --recursive ~/workspace", width=1180)
    svg.add_command("eis analyze --config eis.yaml --recursive ~/workspace")
    svg.add_blank()
    svg.add_text("═══ Backend ═══", color='red', bold=True)
    svg.add_text("Analyzed 12 repo(s), 10 engineers", color='fg_dim')
    svg.add_blank()

    #                #   Member Active Prod Qual Robust Dorm Design Brdth Debt Indisp Grav Total     Role       Style      State
    cols = [28, 120, 55, 46, 46, 58, 55, 58, 52, 46, 52, 48, 100, 140, 132, 110]
    headers = ['#', 'Member', 'Active', 'Prod', 'Qual', 'Robust', 'Dormnt', 'Design', 'Brdth', 'Debt', 'Indisp', 'Grav', 'Total', 'Role', 'Style', 'State']
    svg.add_table_row([(h, 'aqua', True) for h in headers], cols)
    svg.add_separator()

    rows = [
        ('1', 'machuz', '✓', '100', '57', '100', '100', '100', '74', '100', '43', '97', '90.3', 'Architect (1.00)', 'Builder (1.00)', 'Active'),
        ('2', 'Engineer F', '—', '69', '73', '12', '67', '81', '81', '11', '100', '52', '52.8', 'Architect (0.88)', '—', 'Former'),
        ('3', 'Engineer G', '✓', '17', '69', '50', '14', '48', '48', '88', '35', '44', '44.5', 'Anchor (0.96)', 'Balanced (0.30)', 'Active'),
        ('4', 'Engineer H', '✓', '27', '84', '30', '28', '52', '52', '71', '8', '41', '41.8', 'Anchor (0.98)', 'Balanced (0.30)', 'Active'),
        ('5', 'Engineer X', '—', '6', '79', '0', '4', '78', '78', '50', '0', '24', '24.9', '—', 'Spread (1.00)', 'Former'),
    ]

    for r in rows:
        rank, name, active, prod, qual, robust, dormant, design, breadth, debt, indisp, grav, total_str, role_str, style, state = r
        cells = [
            (rank, 'fg'),
            (name, 'yellow'),
            (active, 'green' if active == '✓' else 'fg_dim'),
            (prod, make_score_color(prod)),
            (qual, make_score_color(qual)),
            (robust, make_score_color(robust)),
            (dormant, make_score_color(dormant)),
            (design, make_score_color(design)),
            (breadth, make_score_color(breadth)),
            (debt, make_score_color(debt)),
            (indisp, make_score_color(indisp)),
            (grav, make_score_color(grav)),
            (total_str, total_color(total_str), True),
            (role_str, role_color(role_str)),
            (style, 'blue' if style != '—' else 'fg_dim'),
            (state, state_color(state)),
        ]
        svg.add_table_row(cells, cols)

    svg.save(os.path.join(IMG_DIR, 'ch1-backend-table.svg'))


def ch1_frontend_table():
    svg = TerminalSVG("Terminal — eis analyze (Frontend)", width=1180)
    svg.add_text("═══ Frontend ═══", color='red', bold=True)
    svg.add_text("Analyzed 8 repo(s), 8 engineers", color='fg_dim')
    svg.add_blank()

    cols = [25, 110, 50, 42, 42, 52, 50, 52, 48, 42, 48, 42, 50, 135, 125, 100]
    headers = ['#', 'Member', 'Active', 'Prod', 'Qual', 'Robust', 'Dormnt', 'Design', 'Brdth', 'Debt', 'Indisp', 'Grav', 'Total', 'Role', 'Style', 'State']
    svg.add_table_row([(h, 'aqua', True) for h in headers], cols)
    svg.add_separator()

    rows = [
        ('1', 'Engineer D', '✓', '100', '84', '100', '100', '100', '62', '39', '100', '84', '85.4', 'Architect (1.00)', 'Builder (1.00)', 'Active'),
        ('—', 'Engineer Y', '—', '24', '18', '0', '17', '38', '38', '0', '0', '12', '12.6', 'Producer (0.68)', 'Mass (0.81)', 'Former'),
    ]

    for r in rows:
        rank, name, active, prod, qual, robust, dormant, design, breadth, debt, indisp, grav, total_str, role_str, style, state = r
        cells = [
            (rank, 'fg'),
            (name, 'yellow'),
            (active, 'green' if active == '✓' else 'fg_dim'),
            (prod, make_score_color(prod)),
            (qual, make_score_color(qual)),
            (robust, make_score_color(robust)),
            (dormant, make_score_color(dormant)),
            (design, make_score_color(design)),
            (breadth, make_score_color(breadth)),
            (debt, make_score_color(debt)),
            (indisp, make_score_color(indisp)),
            (grav, make_score_color(grav)),
            (total_str, total_color(total_str), True),
            (role_str, role_color(role_str)),
            (style, 'blue' if style not in ('—', '') else 'fg_dim'),
            (state, state_color(state)),
        ]
        svg.add_table_row(cells, cols)

    svg.save(os.path.join(IMG_DIR, 'ch1-frontend-table.svg'))


# ════════════════════════════════════════════════════════════
# CHAPTER 2 — Team Health
# ════════════════════════════════════════════════════════════

def ch2_warnings():
    svg = TerminalSVG("Terminal — eis team --recursive ~/workspace", width=900)
    svg.add_command("eis team --recursive ~/workspace")
    svg.add_blank()
    svg.add_text("═══ Backend (4 core + 3 risk / 12 total, 13 repos) ═══", color='red', bold=True)
    svg.add_blank()
    svg.add_text("⚠ Warnings:", color='orange', bold=True)
    svg.add_text("  43% risk ratio — 3 of 7 effective members are Former/Silent/Fragile", color='orange')
    svg.add_text("  Top contributor (machuz) accounts for 46% of core production", color='orange')
    svg.add_text("  — ProdDensity drops to 39 without them", color='orange')
    svg.add_text("  2 Silent members — headcount says 16 but effective contributors are 4", color='orange')
    svg.add_text("  Fragile gravity — okatechnology (Grav 68) has high influence", color='orange')
    svg.add_text("    but low robust survival (8)", color='orange')
    svg.save(os.path.join(IMG_DIR, 'ch2-warnings.svg'))


# ════════════════════════════════════════════════════════════
# CHAPTER 3 — Archetypes
# ════════════════════════════════════════════════════════════

def ch3_engineer_profiles():
    svg = TerminalSVG("Terminal — Engineer Archetypes", width=900)

    profiles = [
        ("Engineer A — The Builder Architect", "purple",
         "Prod 100 | Qual 84 | Robust 100 | Dormant 100 | Design 100 | Grav 84 | Total 88.9",
         "Role: Architect (1.00) | Style: Builder (1.00)"),
        ("Engineer B — The Mass Anchor", "blue",
         "Prod 100 | Qual 87 | Robust 11 | Dormant 18 | Design 39 | Grav 32 | Total 44.6",
         "Role: Anchor (1.00) | Style: Mass (0.81)"),
        ("Engineer C — The Balanced Producer", "yellow",
         "Prod 42 | Qual 44 | Robust 39 | Dormant 39 | Design 9 | Grav 29 | Total 39.5",
         "Role: Producer (0.80) | Style: Balanced (0.30)"),
        ("Engineer D — The Emergent Producer", "yellow",
         "Prod 49 | Qual 66 | Robust 12 | Dormant 44 | Design 5 | Grav 68 | Indisp 100 | Total 39.0",
         "Role: Producer (0.96) | Style: Emergent (0.78)"),
        ("Engineer E — The Churn Producer", "orange",
         "Prod 37 | Qual 23 | Robust 10 | Dormant 2 | Design 2 | Grav 35 | Total 19.2",
         "Role: Producer (0.68) | Style: Churn (0.67)"),
    ]

    for name, color, scores, role in profiles:
        svg.add_text(name, color=color, bold=True)
        svg.add_text(f"  {scores}", color='fg')
        svg.add_text(f"  {role}", color='blue')
        svg.add_blank()

    svg.save(os.path.join(IMG_DIR, 'ch3-engineer-profiles.svg'))


def ch3_producer_warning():
    svg = TerminalSVG("Terminal — Team Warning", width=500)
    svg.add_text("Role Distribution:", color='aqua', bold=True)
    svg.add_colored_spans([
        ("  Producer     ", 'yellow'),
        ("██████████", 'yellow'),
        ("  5 (100%)", 'fg_dim'),
    ])
    svg.save(os.path.join(IMG_DIR, 'ch3-producer-warning.svg'))


# ════════════════════════════════════════════════════════════
# CHAPTER 4 — Backend Architect Concentration
# ════════════════════════════════════════════════════════════

def ch4_backend_team():
    svg = TerminalSVG("Terminal — eis analyze (Backend)", width=1250)
    svg.add_text("═══ Backend ═══", color='red', bold=True)
    svg.add_text("Analyzed 13 repo(s), 12 engineers", color='fg_dim')
    svg.add_blank()

    #                #  Member Active Prod Qual Robust Dorm Design Grav Total     Role       Style       State
    cols = [28, 120, 55, 46, 46, 58, 55, 58, 48, 100, 145, 140, 125]
    headers = ['#', 'Member', 'Active', 'Prod', 'Qual', 'Robust', 'Dormnt', 'Design', 'Grav', 'Total', 'Role', 'Style', 'State']
    svg.add_table_row([(h, 'aqua', True) for h in headers], cols)
    svg.add_separator()

    rows = [
        ('1', 'machuz', '✓', '100', '66', '100', '100', '100', '97', '92.4', 'Architect (1.00)', 'Builder (1.00)', 'Active (0.80)'),
        ('2', 'Engineer F', '—', '93', '75', '36', '21', '47', '76', '55.5', 'Anchor (0.87)', 'Resilient (0.66)', 'Former (0.73)'),
        ('3', 'Engineer G', '✓', '52', '78', '21', '32', '12', '26', '37.3', 'Anchor (0.96)', 'Balanced (0.30)', 'Active (0.80)'),
        ('4', 'Engineer H', '✓', '49', '90', '20', '25', '10', '31', '35.6', 'Anchor (0.98)', 'Balanced (0.30)', 'Active (0.80)'),
    ]

    for r in rows:
        rank, name, active, prod, qual, robust, dormant, design, grav, total_str, role_str, style, state = r
        cells = [
            (rank, 'fg'),
            (name, 'yellow'),
            (active, 'green' if active == '✓' else 'fg_dim'),
            (prod, make_score_color(prod)),
            (qual, make_score_color(qual)),
            (robust, make_score_color(robust)),
            (dormant, make_score_color(dormant)),
            (design, make_score_color(design)),
            (grav, make_score_color(grav)),
            (total_str, total_color(total_str), True),
            (role_str, role_color(role_str)),
            (style, 'blue'),
            (state, state_color(state)),
        ]
        svg.add_table_row(cells, cols)

    svg.save(os.path.join(IMG_DIR, 'ch4-backend-team.svg'))


def ch4_team_classification():
    svg = TerminalSVG("Terminal — eis team (Backend)", width=700)
    svg.add_text("═══ Backend (4 core + 3 risk / 12 total, 13 repos) ═══", color='red', bold=True)
    svg.add_text("  ★ Elite (1.00)", color='yellow', bold=True)
    svg.add_blank()
    svg.add_text("Classification:", color='aqua', bold=True)
    svg.add_text("  Structure: Emerging Architecture (0.66)", color='fg')
    svg.add_text("  Phase:     Legacy-Heavy (0.67)", color='fg')
    svg.add_text("  Risk:      Talent Drain (0.43)", color='orange')
    svg.add_blank()
    svg.add_text("Role Distribution:", color='aqua', bold=True)
    svg.add_colored_spans([("  Architect    ", 'purple'), ("█░░░░░░░░░", 'purple'), ("  1 (14%)", 'fg_dim')])
    svg.add_colored_spans([("  Anchor       ", 'blue'), ("████░░░░░░", 'blue'), ("  3 (43%)", 'fg_dim')])
    svg.add_colored_spans([("  —            ", 'fg_dim'), ("████░░░░░░", 'fg_dim'), ("  3 (43%)", 'fg_dim')])
    svg.save(os.path.join(IMG_DIR, 'ch4-team-classification.svg'))


def ch4_structure():
    svg = TerminalSVG("Backend Team Structure", width=400)
    svg.add_text("Star (Architect)", color='purple', bold=True)
    svg.add_text("    ↓", color='fg_dim')
    svg.add_text("  Planets (Anchors)", color='blue', bold=True)
    svg.add_text("    ↓", color='fg_dim')
    svg.add_text("  Vacuum (No Producers)", color='red', bold=True)
    svg.save(os.path.join(IMG_DIR, 'ch4-structure.svg'))


# ════════════════════════════════════════════════════════════
# CHAPTER 5 — Timeline
# ════════════════════════════════════════════════════════════

def _timeline_table(svg, name, domain, rows):
    """Helper for timeline table generation."""
    svg.add_text(f"--- {name} ({domain}) ---", color='yellow', bold=True)
    cols = [140, 55, 50, 50, 50, 60, 130, 130, 120]
    headers = ['Period', 'Total', 'Prod', 'Qual', 'Surv', 'Design', 'Role', 'Style', 'State']
    svg.add_table_row([(h, 'aqua', True) for h in headers], cols)
    svg.add_separator()

    for r in rows:
        period, total, prod, qual, surv, design = r[:6]
        role = r[6] if len(r) > 6 else ''
        style = r[7] if len(r) > 7 else ''
        state = r[8] if len(r) > 8 else ''
        cells = [
            (period, 'fg_dim'),
            (total, total_color(total), True),
            (prod, make_score_color(prod)),
            (qual, make_score_color(qual)),
            (surv, make_score_color(surv)),
            (design, make_score_color(design)),
            (role, role_color(role)),
            (style, 'blue' if style and style != '—' else 'fg_dim'),
            (state, state_color(state)),
        ]
        svg.add_table_row(cells, cols)


def ch5_engineer_f_timeline():
    svg = TerminalSVG("Terminal — eis timeline (Engineer F, Backend)", width=950)
    svg.add_command("eis timeline --author engineer-f --recursive ~/workspace")
    svg.add_blank()

    rows = [
        ('2024-Q1 (Jan)', '90.0', '100', '69', '100', '100', 'Architect', 'Builder', ''),
        ('2024-Q2 (Apr)', '94.4', '100', '71', '100', '87', 'Architect', 'Builder', ''),
        ('2024-Q3 (Jul)', '72.5', '59', '72', '100', '71', 'Producer', 'Balanced', ''),
        ('2024-Q4 (Oct)', '90.6', '100', '77', '100', '100', 'Architect', 'Builder', ''),
        ('2025-Q1 (Jan)', '79.2', '100', '82', '100', '28', 'Anchor', 'Balanced', ''),
        ('2025-Q2 (Apr)', '68.4', '36', '84', '100', '58', 'Anchor', 'Balanced', ''),
        ('2025-Q3 (Jul)', '49.1', '81', '77', '51', '4', 'Anchor', 'Balanced', ''),
        ('2025-Q4 (Oct)', '31.2', '18', '78', '23', '8', '—', 'Balanced', 'Fragile'),
        ('2026-Q1 (Jan)', '11.3', '0', '0', '34', '0', '—', '—', 'Former'),
    ]
    _timeline_table(svg, "Engineer F", "Backend", rows)
    svg.save(os.path.join(IMG_DIR, 'ch5-engineer-f-timeline.svg'))


def ch5_engineer_j_timeline():
    svg = TerminalSVG("Terminal — eis timeline (Engineer J, Frontend)", width=950)
    rows = [
        ('2024-Q1 (Jan)', '28.1', '26', '73', '33', '2', 'Anchor', '—', 'Growing'),
        ('2024-Q2 (Apr)', '15.5', '8', '100', '16', '0', '—', '—', 'Growing'),
        ('2024-Q3 (Jul)', '61.9', '52', '72', '38', '100', 'Architect', 'Balanced', ''),
        ('2024-Q4 (Oct)', '91.7', '100', '74', '96', '100', 'Architect', 'Builder', ''),
        ('2025-Q1 (Jan)', '63.9', '90', '85', '15', '61', 'Anchor', 'Emergent', ''),
        ('2025-Q2 (Apr)', '63.8', '48', '73', '76', '81', 'Architect', 'Balanced', ''),
        ('2025-Q3 (Jul)', '44.7', '62', '70', '18', '18', 'Producer', 'Emergent', ''),
        ('2025-Q4 (Oct)', '39.4', '62', '60', '50', '0', 'Producer', 'Balanced', 'Former'),
        ('2026-Q1 (Jan)', '54.2', '43', '61', '100', '1', 'Producer', 'Balanced', 'Active'),
    ]
    _timeline_table(svg, "Engineer J", "Frontend", rows)
    svg.save(os.path.join(IMG_DIR, 'ch5-engineer-j-timeline.svg'))


def ch5_engineer_i_timeline():
    svg = TerminalSVG("Terminal — eis timeline (Engineer I, Frontend)", width=950)
    rows = [
        ('2024-Q3 (Jul)', '56.1', '100', '97', '60', '2', 'Anchor', 'Balanced', ''),
        ('2024-Q4 (Oct)', '75.7', '59', '84', '100', '78', 'Architect', 'Balanced', ''),
        ('2025-Q1 (Jan)', '87.5', '100', '93', '100', '100', 'Architect', 'Builder', ''),
        ('2025-Q2 (Apr)', '73.2', '67', '91', '100', '100', 'Architect', 'Builder', ''),
        ('2025-Q3 (Jul)', '72.4', '73', '97', '100', '73', 'Anchor', 'Balanced', ''),
        ('2025-Q4 (Oct)', '81.7', '100', '68', '100', '100', 'Architect', 'Balanced', ''),
        ('2026-Q1 (Jan)', '78.1', '100', '84', '83', '100', 'Anchor', 'Builder', 'Active'),
    ]
    _timeline_table(svg, "Engineer I", "Frontend", rows)
    svg.save(os.path.join(IMG_DIR, 'ch5-engineer-i-timeline.svg'))


def ch5_per_repo_commits():
    svg = TerminalSVG("Per-Repository Commit Distribution — Engineer I", width=700)
    cols = [120, 160, 160, 160]
    svg.add_table_row([('Quarter', 'aqua', True), ('Repo A (existing)', 'aqua', True), ('Repo B (existing)', 'aqua', True), ('Repo C (new)', 'aqua', True)], cols)
    svg.add_separator()
    svg.add_table_row([('2025-Q2', 'fg_dim'), ('135', 'fg'), ('44', 'fg'), ('—', 'fg_dim')], cols)
    svg.add_table_row([('2025-Q3', 'fg_dim'), ('201', 'fg'), ('274', 'fg'), ('—', 'fg_dim')], cols)
    svg.add_table_row([('2025-Q4', 'fg_dim'), ('5', 'fg_dim'), ('5', 'fg_dim'), ('1,352', 'purple', True)], cols)
    svg.add_table_row([('2026-Q1', 'fg_dim'), ('2', 'fg_dim'), ('2', 'fg_dim'), ('1,333', 'purple', True)], cols)
    svg.save(os.path.join(IMG_DIR, 'ch5-per-repo-commits.svg'))


def ch5_transitions():
    svg = TerminalSVG("Terminal — Notable Transitions", width=800)
    svg.add_text("Notable transitions:", color='aqua', bold=True)
    transitions_i = [
        ("  * Engineer I: Role ", "Anchor", "→", "Architect", " (2024-Q4)", ""),
        ("  * Engineer I: Style ", "Balanced", "→", "Builder", " (2025-Q1)", ""),
        ("  * Engineer I: Role ", "Architect", "→", "Anchor", " (2025-Q3)", "  ← friction"),
        ("  * Engineer I: Style ", "Builder", "→", "Balanced", " (2025-Q3)", "  ← hesitation"),
        ("  * Engineer I: Role ", "Anchor", "→", "Architect", " (2025-Q4)", "  ← return"),
    ]
    for t in transitions_i:
        svg.add_colored_spans([
            (t[0], 'fg'),
            (t[1], 'blue'),
            (t[2], 'fg_dim'),
            (t[3], role_color(t[3])),
            (t[4], 'fg_dim'),
            (t[5], 'red') if t[5] else (t[5], 'fg'),
        ])

    svg.add_blank()
    transitions_j = [
        ("  * Engineer J: Style ", "Balanced", "→", "Builder", " (2024-Q4)", "  ← building phase"),
        ("  * Engineer J: Role ", "Architect", "→", "Producer", " (2025-Q3)", "  ← structure complete"),
        ("  * Engineer J: State ", "Former", "→", "Active", " (2026-Q1)", "  ← return"),
    ]
    for t in transitions_j:
        svg.add_colored_spans([
            (t[0], 'fg'),
            (t[1], 'blue'),
            (t[2], 'fg_dim'),
            (t[3], role_color(t[3]) if 'Role' in t[0] else state_color(t[3]) if 'State' in t[0] else 'blue'),
            (t[4], 'fg_dim'),
            (t[5], 'green') if 'return' in t[5] else (t[5], 'aqua'),
        ])

    svg.save(os.path.join(IMG_DIR, 'ch5-transitions.svg'))


def ch5_comparison_table():
    svg = TerminalSVG("Timeline Comparison — Engineer F vs machuz (Backend)", width=900)

    cols = [120, 60, 140, 30, 60, 140]
    svg.add_table_row([
        ('', 'fg'), ('', 'fg'),
        ('Engineer F (BE)', 'yellow', True),
        ('', 'fg'), ('', 'fg'),
        ('machuz (BE)', 'yellow', True),
    ], cols)
    svg.add_separator()

    rows = [
        ('2024-Q1', '90.0', 'Architect Builder', '', '--', ''),
        ('2024-Q2', '94.4', 'Architect Builder', '', '31.5', 'Anchor Balanced'),
        ('2024-Q3', '72.5', 'Producer Balanced', '', '73.8', 'Anchor Builder'),
        ('2024-Q4', '90.6', 'Architect Builder', '', '64.1', 'Anchor Builder'),
        ('2025-Q1', '79.2', 'Anchor Balanced', '', '61.7', 'Anchor Builder'),
        ('2025-Q2', '68.4', 'Anchor Balanced', '', '49.2', 'Anchor Balanced'),
        ('2025-Q3', '49.1', 'Anchor Balanced', '', '93.2', 'Architect Builder'),
        ('2025-Q4', '31.2', '— Fragile', '', '87.7', 'Architect Builder'),
        ('2026-Q1', '11.3', '— Former', '', '92.4', 'Architect Builder'),
    ]
    for r in rows:
        period, f_total, f_role, _, m_total, m_role = r
        cells = [
            (period, 'fg_dim'),
            (f_total, total_color(f_total), True),
            (f_role, role_color(f_role) if 'Architect' in f_role or 'Anchor' in f_role or 'Producer' in f_role else 'red' if 'Fragile' in f_role or 'Former' in f_role else 'fg_dim'),
            ('  ', 'fg'),
            (m_total if m_total != '--' else '—', total_color(m_total) if m_total != '--' else 'fg_dim', True),
            (m_role, role_color(m_role) if m_role else 'fg_dim'),
        ]
        svg.add_table_row(cells, cols)

    svg.save(os.path.join(IMG_DIR, 'ch5-comparison-table.svg'))


def ch5_team_timeline():
    svg = TerminalSVG("Terminal — eis timeline --team (Backend)", width=800)
    svg.add_text("=== Backend / Backend -- Team Timeline ===", color='red', bold=True)
    svg.add_blank()
    svg.add_text("Classification:", color='aqua', bold=True)

    cols = [140, 160, 160, 170]
    svg.add_table_row([('Period', 'fg_dim'), ('2024-Q4', 'fg_dim'), ('2025-Q4', 'fg_dim'), ('2026-Q1', 'fg_dim')], cols)
    svg.add_table_row([('Character', 'fg'), ('Guardian', 'blue'), ('Balanced', 'green'), ('Elite', 'yellow', True)], cols)
    svg.add_table_row([('Structure', 'fg'), ('Maintenance', 'fg_dim'), ('Unstructured', 'red'), ('Architectural Engine', 'purple', True)], cols)
    svg.add_table_row([('Phase', 'fg'), ('Declining', 'red'), ('Declining', 'red'), ('Mature', 'green')], cols)
    svg.add_table_row([('Risk', 'fg'), ('Quality Drift', 'orange'), ('Design Vacuum', 'red'), ('Healthy', 'green')], cols)
    svg.save(os.path.join(IMG_DIR, 'ch5-team-timeline.svg'))


# ════════════════════════════════════════════════════════════
# CHAPTER 6 — Team Evolution
# ════════════════════════════════════════════════════════════

def ch6_backend_team_timeline():
    svg = TerminalSVG("Terminal — Backend Team Timeline", width=700)
    svg.add_text("═══ Backend — Team Timeline ═══", color='red', bold=True)
    svg.add_blank()
    svg.add_text("Classification:", color='aqua', bold=True)
    cols = [140, 180, 200]
    svg.add_table_row([('Period', 'fg_dim'), ('2024-H2', 'fg_dim'), ('2026-H1', 'fg_dim')], cols)
    svg.add_table_row([('Character', 'fg'), ('Balanced', 'green'), ('Elite', 'yellow', True)], cols)
    svg.add_table_row([('Structure', 'fg'), ('Unstructured', 'red'), ('Architectural Engine', 'purple', True)], cols)
    svg.add_table_row([('Culture', 'fg'), ('Stability', 'blue'), ('Builder', 'green')], cols)
    svg.add_table_row([('Phase', 'fg'), ('Declining', 'red'), ('Mature', 'green')], cols)
    svg.add_table_row([('Risk', 'fg'), ('Design Vacuum', 'red'), ('Healthy', 'green')], cols)
    svg.save(os.path.join(IMG_DIR, 'ch6-backend-team-timeline.svg'))


def ch6_backend_score_averages():
    svg = TerminalSVG("Backend Score Averages", width=500)
    svg.add_text("Score Averages:", color='aqua', bold=True)
    cols = [140, 130, 130]
    svg.add_table_row([('Period', 'fg_dim'), ('2024-H2', 'fg_dim'), ('2026-H1', 'fg_dim')], cols)
    svg.add_separator()
    svg.add_table_row([('Production', 'fg'), ('0.0', 'fg_dim'), ('57.7', 'green')], cols)
    svg.add_table_row([('Quality', 'fg'), ('0.0', 'fg_dim'), ('64.6', 'green')], cols)
    svg.add_table_row([('Survival', 'fg'), ('0.0', 'fg_dim'), ('39.2', 'yellow')], cols)
    svg.add_table_row([('Design', 'fg'), ('0.0', 'fg_dim'), ('36.4', 'yellow')], cols)
    svg.add_table_row([('Total', 'fg', True), ('0.0', 'fg_dim'), ('48.3', 'yellow', True)], cols)
    svg.save(os.path.join(IMG_DIR, 'ch6-backend-scores.svg'))


def ch6_frontend_team_timeline():
    svg = TerminalSVG("Terminal — Frontend Team Timeline", width=900)
    svg.add_text("═══ Frontend — Team Timeline ═══", color='red', bold=True)
    svg.add_blank()
    svg.add_text("Classification:", color='aqua', bold=True)
    cols = [120, 150, 140, 150, 160]
    svg.add_table_row([('Period', 'fg_dim'), ('2024-H2', 'fg_dim'), ('2025-H1', 'fg_dim'), ('2025-H2', 'fg_dim'), ('2026-H1', 'fg_dim')], cols)
    svg.add_table_row([('Character', 'fg'), ('Guardian', 'blue'), ('Factory', 'yellow'), ('Guardian', 'blue'), ('Balanced', 'green')], cols)
    svg.add_table_row([('Structure', 'fg'), ('Maintenance', 'fg_dim'), ('Delivery', 'yellow'), ('Maintenance', 'fg_dim'), ('Maintenance', 'fg_dim')], cols)
    svg.add_table_row([('Culture', 'fg'), ('Stability', 'blue'), ('Stability', 'blue'), ('Stability', 'blue'), ('Builder', 'green')], cols)
    svg.add_table_row([('Phase', 'fg'), ('Declining', 'red'), ('Declining', 'red'), ('Declining', 'red'), ('Mature', 'green')], cols)
    svg.add_table_row([('Risk', 'fg'), ('Quality Drift', 'orange'), ('Quality Drift', 'orange'), ('Quality Drift', 'orange'), ('Design Vacuum', 'red')], cols)

    svg.add_blank()
    svg.add_text("Transitions:", color='aqua', bold=True)
    svg.add_colored_spans([("  [2025-H1] Character: ", 'fg'), ("Guardian", 'blue'), (" → ", 'fg_dim'), ("Factory", 'yellow')])
    svg.add_colored_spans([("  [2025-H1] Structure: ", 'fg'), ("Maintenance", 'fg_dim'), (" → ", 'fg_dim'), ("Delivery", 'yellow')])
    svg.add_colored_spans([("  [2025-H2] Character: ", 'fg'), ("Factory", 'yellow'), (" → ", 'fg_dim'), ("Guardian", 'blue')])
    svg.add_colored_spans([("  [2025-H2] Structure: ", 'fg'), ("Delivery", 'yellow'), (" → ", 'fg_dim'), ("Maintenance", 'fg_dim')])
    svg.save(os.path.join(IMG_DIR, 'ch6-frontend-team-timeline.svg'))


def ch6_infra_firmware():
    svg = TerminalSVG("Terminal — Infra & Firmware Classification", width=600)
    svg.add_text("═══ Infra (2026-H1) ═══", color='red', bold=True)
    svg.add_text("  Character:  Explorer", color='aqua')
    svg.add_text("  Structure:  Balanced", color='green')
    svg.add_text("  Culture:    Exploration", color='aqua')
    svg.add_text("  Phase:      Emerging", color='blue')
    svg.add_text("  Risk:       Design Vacuum", color='red')
    svg.add_blank()
    svg.add_text("═══ Firmware (2026-H1) ═══", color='red', bold=True)
    svg.add_text("  Character:  Firefighting", color='orange')
    svg.add_text("  Structure:  Maintenance Team", color='fg_dim')
    svg.add_text("  Culture:    Firefighting", color='orange')
    svg.add_text("  Phase:      Declining", color='red')
    svg.add_text("  Risk:       Design Vacuum", color='red')
    svg.save(os.path.join(IMG_DIR, 'ch6-infra-firmware.svg'))


def ch6_machuz_timeline():
    svg = TerminalSVG("machuz Backend Timeline", width=700)
    svg.add_text("--- machuz (Backend) ---", color='yellow', bold=True)
    rows = [
        ('2024-H1', '27.6', 'Anchor', '—', 'Growing'),
        ('2024-H2', '76.4', 'Anchor', 'Builder', ''),
        ('2025-H1', '58.4', 'Producer', 'Balanced', ''),
        ('2025-H2', '92.5', 'Architect', 'Builder', ''),
        ('2026-H1', '92.4', 'Architect', 'Builder', 'Active'),
    ]
    cols = [120, 60, 120, 120, 120]
    for r in rows:
        period, total, role, style, state = r
        cells = [
            (period, 'fg_dim'),
            (total, total_color(total), True),
            (role, role_color(role)),
            (style, 'blue' if style and style != '—' else 'fg_dim'),
            (state, state_color(state)),
        ]
        svg.add_table_row(cells, cols)
    svg.save(os.path.join(IMG_DIR, 'ch6-machuz-timeline.svg'))


def ch6_be_architects():
    svg = TerminalSVG("Backend Architect Concentration", width=900)
    svg.add_text("Backend — Architect Seats Over Time", color='aqua', bold=True)
    svg.add_blank()
    cols = [100, 60, 170, 30, 60, 170]
    rows = [
        ('2024-H1', '93.5', 'Engineer F  Architect Builder', '', '27.6', 'machuz  Anchor'),
        ('2024-H2', '84.1', 'Engineer F  Architect Builder', '', '76.4', 'machuz  Anchor Builder'),
        ('2025-H1', '72.7', 'Engineer F  Anchor Balanced', '', '58.4', 'machuz  Producer'),
        ('2025-H2', '37.5', 'Engineer F  Anchor', '', '92.5', 'machuz  Architect Builder'),
    ]
    for r in rows:
        period, f_score, f_info, _, m_score, m_info = r
        f_color = 'purple' if 'Architect' in f_info else 'blue' if 'Anchor' in f_info else 'fg'
        m_color = 'purple' if 'Architect' in m_info else 'blue' if 'Anchor' in m_info else 'yellow'
        cells = [
            (period, 'fg_dim'),
            (f_score, total_color(f_score), True),
            (f_info, f_color),
            ('', 'fg'),
            (m_score, total_color(m_score), True),
            (m_info, m_color),
        ]
        svg.add_table_row(cells, cols)
    svg.save(os.path.join(IMG_DIR, 'ch6-be-architects.svg'))


def ch6_fe_architects():
    svg = TerminalSVG("Frontend Architect Flow", width=900)
    svg.add_text("Frontend — Architect Seats Over Time", color='aqua', bold=True)
    svg.add_blank()
    cols = [100, 60, 170, 30, 60, 170]
    rows = [
        ('2024-H2', '72.7', 'Engineer I  Anchor', '', '74.9', 'Engineer J  Architect Builder'),
        ('2025-H1', '83.8', 'Engineer I  Architect', '', '54.3', 'Engineer J  Anchor'),
        ('2025-H2', '85.1', 'Engineer I  Architect', '', '38.6', 'Engineer J  Anchor'),
    ]
    for r in rows:
        period, i_score, i_info, _, j_score, j_info = r
        i_color = 'purple' if 'Architect' in i_info else 'blue'
        j_color = 'purple' if 'Architect' in j_info else 'blue'
        cells = [
            (period, 'fg_dim'),
            (i_score, total_color(i_score), True),
            (i_info, i_color),
            ('', 'fg'),
            (j_score, total_color(j_score), True),
            (j_info, j_color),
        ]
        svg.add_table_row(cells, cols)
    svg.save(os.path.join(IMG_DIR, 'ch6-fe-architects.svg'))


def ch6_engineer_k():
    svg = TerminalSVG("Engineer K — Frontend Lifecycle", width=600)
    svg.add_text("--- Engineer K (Frontend) ---", color='yellow', bold=True)
    rows = [
        ('2024-H1', '87.8', 'Architect', 'Builder', ''),
        ('2024-H2', '14.6', '—', '—', ''),
        ('2025-H1', '7.1', '—', '—', 'Silent'),
        ('2025-H2', '3.2', '—', '—', ''),
        ('2026-H1', '3.2', '—', '—', ''),
    ]
    cols = [120, 60, 120, 120, 120]
    for r in rows:
        period, total, role, style, state = r
        cells = [
            (period, 'fg_dim'),
            (total, total_color(total), True),
            (role, role_color(role)),
            (style, 'blue' if style and style != '—' else 'fg_dim'),
            (state, state_color(state)),
        ]
        svg.add_table_row(cells, cols)
    svg.save(os.path.join(IMG_DIR, 'ch6-engineer-k.svg'))


def ch6_gravity_transfer():
    svg = TerminalSVG("Gravity Transfer — Frontend", width=700)
    svg.add_text("Score Transfer:", color='aqua', bold=True)
    svg.add_blank()
    svg.add_colored_spans([("Engineer K:  ", 'yellow'), ("87.8", 'purple', True), (" → ", 'fg_dim'), ("14.6", 'fg'), (" → ", 'fg_dim'), ("7.1", 'fg_dim'), (" → ", 'fg_dim'), ("3.2", 'fg_dim'), (" → ", 'fg_dim'), ("3.2", 'fg_dim')])
    svg.add_colored_spans([("Engineer I:  ", 'yellow'), ("  — ", 'fg_dim'), (" → ", 'fg_dim'), ("72.7", 'green'), (" → ", 'fg_dim'), ("83.8", 'purple', True), (" → ", 'fg_dim'), ("85.1", 'purple', True), (" → ", 'fg_dim'), ("78.1", 'green')])
    svg.add_colored_spans([("Engineer J:  ", 'yellow'), ("25.9", 'fg'), (" → ", 'fg_dim'), ("74.9", 'green'), (" → ", 'fg_dim'), ("54.3", 'green'), (" → ", 'fg_dim'), ("38.6", 'yellow'), (" → ", 'fg_dim'), ("54.2", 'green')])
    svg.save(os.path.join(IMG_DIR, 'ch6-gravity-transfer.svg'))


def ch6_evolution_paths():
    svg = TerminalSVG("Evolution Paths", width=800)
    svg.add_text("Evolution Paths:", color='aqua', bold=True)
    svg.add_blank()
    svg.add_colored_spans([("machuz:     ", 'yellow'), ("Anchor", 'blue'), (" → ", 'fg_dim'), ("Anchor Builder", 'blue'), (" → ", 'fg_dim'), ("Producer Balanced", 'yellow'), (" → ", 'fg_dim'), ("Architect Builder", 'purple', True)])
    svg.add_colored_spans([("Engineer I: ", 'yellow'), ("Anchor Balanced", 'blue'), (" → ", 'fg_dim'), ("Architect Balanced", 'purple'), (" → ", 'fg_dim'), ("Architect Builder", 'purple', True)])
    svg.add_colored_spans([("Engineer J: ", 'yellow'), ("Anchor Growing", 'blue'), (" → ", 'fg_dim'), ("Architect Balanced", 'purple'), (" → ", 'fg_dim'), ("Architect Builder", 'purple', True)])
    svg.add_colored_spans([("Engineer F: ", 'yellow'), ("(first appearance) ", 'fg_dim'), ("Architect Builder", 'purple', True)])
    svg.save(os.path.join(IMG_DIR, 'ch6-evolution-paths.svg'))


def ch6_evolution_model():
    svg = TerminalSVG("Evolution Model Overview", width=750)
    lines = [
        ("┌──────────────────────────────────────────────────────────┐", 'fg_dim'),
        ("│              Evolution Model Overview                    │", 'aqua'),
        ("├──────────────────────────────────────────────────────────┤", 'fg_dim'),
        ("│                                                          │", 'fg_dim'),
    ]
    for text, color in lines:
        svg.add_text(text, color=color)

    svg.add_colored_spans([("│  ", 'fg_dim'), ("[Growing]", 'blue'), (" → ", 'fg_dim'), ("[Anchor]", 'blue'), (" → ", 'fg_dim'), ("[Producer]", 'yellow'), (" → ", 'fg_dim'), ("[Architect]", 'purple'), ("        │", 'fg_dim')])
    svg.add_text("│                  ↑            │            │             │", color='fg_dim')
    svg.add_text("│                  │            │   structure │             │", color='fg_dim')
    svg.add_text("│                  │            ←─────────────┘             │", color='fg_dim')
    svg.add_text("│                  │       (metabolism: back to Producer)   │", color='fg_dim')
    svg.add_text("│                                                          │", color='fg_dim')

    svg.add_colored_spans([("│  * Permeation: ", 'fg_dim'), ("[Anchor]", 'blue'), (" → ", 'fg_dim'), ("[Producer]", 'yellow'), (" → ", 'fg_dim'), ("[Architect]", 'purple'), ("        │", 'fg_dim')])
    svg.add_colored_spans([("│  * Immediate:  ", 'fg_dim'), ("[Anchor]", 'blue'), (" → ", 'fg_dim'), ("[Architect]", 'purple'), (" direct               │", 'fg_dim')])
    svg.add_colored_spans([("│  * Founding:   ", 'fg_dim'), ("[Architect]", 'purple'), (" → score decline (= success)   │", 'fg_dim')])
    svg.add_text("│                                                          │", color='fg_dim')
    svg.add_text("├──────────────────────────────────────────────────────────┤", color='fg_dim')
    svg.add_colored_spans([("│  BE: Architect seats tend to ", 'fg'), ("concentrate", 'red'), (" (observed: 1)    │", 'fg')])
    svg.add_colored_spans([("│  FE: Architect seats are ", 'fg'), ("fluid", 'green'), (" (observed: 1–2)           │", 'fg')])
    svg.add_text("├──────────────────────────────────────────────────────────┤", color='fg_dim')
    svg.add_text("│  Builder prerequisite: Builder experience → Architect    │", color='fg')
    svg.add_text("│  Producer fuel: using structure deeply powers design     │", color='fg')
    svg.add_colored_spans([("│  Producer Vacuum: no producers = ", 'fg'), ("structure sits idle", 'red'), ("      │", 'fg')])
    svg.add_text("└──────────────────────────────────────────────────────────┘", color='fg_dim')
    svg.save(os.path.join(IMG_DIR, 'ch6-evolution-model.svg'))


# ════════════════════════════════════════════════════════════
# CHAPTER 7 — Universe of Code
# ════════════════════════════════════════════════════════════

def ch7_team_classification():
    svg = TerminalSVG("Terminal — Team Timeline Classification", width=800)
    svg.add_text("Classification:", color='aqua', bold=True)
    cols = [120, 180, 180, 190]
    svg.add_table_row([('Period', 'fg_dim'), ('2024-H1', 'fg_dim'), ('2024-H2', 'fg_dim'), ('2025-H1', 'fg_dim')], cols)
    svg.add_table_row([('Character', 'fg'), ('Elite', 'yellow', True), ('Guardian', 'blue'), ('Elite', 'yellow', True)], cols)
    svg.add_table_row([('Structure', 'fg'), ('Architectural Team', 'purple'), ('Maintenance Team', 'fg_dim'), ('Architectural Engine', 'purple', True)], cols)
    svg.add_table_row([('Risk', 'fg'), ('Bus Factor', 'orange'), ('Design Vacuum', 'red'), ('Healthy', 'green')], cols)
    svg.save(os.path.join(IMG_DIR, 'ch7-team-classification.svg'))


# ════════════════════════════════════════════════════════════
# CHAPTER 8 — Engineering Relativity
# ════════════════════════════════════════════════════════════

def ch8_repo_scores():
    svg = TerminalSVG("Same Engineer, Different Universes", width=600)
    svg.add_colored_spans([("Repo A (Backend API)           Total: ", 'fg'), ("35", 'yellow', True)])
    svg.add_colored_spans([("Repo B (New microservice)      Total: ", 'fg'), ("60", 'green', True)])
    svg.save(os.path.join(IMG_DIR, 'ch8-repo-scores.svg'))


def ch8_structure_comparison():
    svg = TerminalSVG("Gravitational Field Strength", width=800)
    svg.add_colored_spans([("Structure: ", 'fg'), ("Architectural Engine", 'purple', True), ("  →  Strong gravitational field", 'fg_dim')])
    svg.add_text("                                  (scores are hard-earned)", color='fg_dim')
    svg.add_blank()
    svg.add_colored_spans([("Structure: ", 'fg'), ("Unstructured", 'red', True), ("          →  Weak gravitational field", 'fg_dim')])
    svg.add_text("                                  (scores come easily)", color='fg_dim')
    svg.save(os.path.join(IMG_DIR, 'ch8-structure-comparison.svg'))


def ch8_per_repo_breakdown():
    svg = TerminalSVG("Terminal — eis analyze --recursive --per-repo", width=800)
    svg.add_command("eis analyze --recursive --per-repo ~/workspace")
    svg.add_blank()
    svg.add_text("─── Backend Per-Repository Breakdown ───", color='red', bold=True)
    svg.add_blank()

    cols = [100, 160, 150, 150, 160]
    svg.add_table_row([('Author', 'aqua', True), ('api-manage', 'aqua', True), ('api', 'aqua', True), ('worker', 'aqua', True), ('Pattern', 'aqua', True)], cols)
    svg.add_separator()
    svg.add_table_row([('machuz', 'yellow'), ('Architect 94', 'purple', True), ('Architect 73', 'purple'), ('Architect 76', 'purple'), ('Reproducible', 'green', True)], cols)
    svg.add_table_row([('alice', 'yellow'), ('Producer 34', 'yellow'), ('Architect 71', 'purple'), ('30', 'fg'), ('Context-dependent', 'blue')], cols)
    svg.add_table_row([('bob', 'yellow'), ('Anchor 41', 'blue'), ('30', 'fg'), ('Cleaner 34', 'aqua'), ('Variable', 'fg_dim')], cols)
    svg.save(os.path.join(IMG_DIR, 'ch8-per-repo-breakdown.svg'))


# ════════════════════════════════════════════════════════════
# COVER IMAGES
# ════════════════════════════════════════════════════════════

def cover_image(chapter, title, subtitle, visual_fn):
    """Create a chapter cover image."""
    svg = TerminalSVG(f"Engineering Impact Score — Chapter {chapter}", width=1200)
    svg.y = 70
    svg.add_text(f"Chapter {chapter}", color='fg_dim')
    svg.add_blank()
    svg.add_text(title, color='yellow', bold=True)
    # override font size for title
    svg.lines[-1] = svg.lines[-1].replace(f'font-size="{FONT_SIZE}"', 'font-size="24"')
    svg.add_blank()
    svg.add_text(subtitle, color='fg_dim')
    svg.lines[-1] = svg.lines[-1].replace(f'font-size="{FONT_SIZE}"', 'font-size="14"')
    svg.add_blank(2)

    visual_fn(svg)

    svg.save(os.path.join(IMG_DIR, f'cover-ch{chapter}.svg'))


def cover_ch1_visual(svg):
    svg.add_text("  7-Axis Scoring Model", color='aqua', bold=True)
    svg.add_blank()
    bars = [
        ('Production', 75, 'green'),
        ('Quality', 64, 'green'),
        ('Survival', 100, 'purple'),
        ('Design', 85, 'purple'),
        ('Breadth', 60, 'green'),
        ('Debt', 78, 'green'),
        ('Indispensability', 43, 'yellow'),
    ]
    for name, val, color in bars:
        bar = '█' * (val // 5) + '░' * (20 - val // 5)
        svg.add_colored_spans([
            (f"  {name:<20}", 'fg'),
            (bar, color),
            (f"  {val}", color, True),
        ])


def cover_ch2_visual(svg):
    svg.add_text("  Team Health Radar", color='aqua', bold=True)
    svg.add_blank()
    axes = [
        ('Complementarity', 72, 'green'),
        ('Growth Potential', 58, 'green'),
        ('Sustainability', 45, 'yellow'),
        ('Productivity', 67, 'green'),
        ('Quality', 81, 'purple'),
        ('Bus Factor', 35, 'red'),
        ('Gravity Health', 62, 'green'),
    ]
    for name, val, color in axes:
        bar = '█' * (val // 5) + '░' * (20 - val // 5)
        svg.add_colored_spans([
            (f"  {name:<20}", 'fg'),
            (bar, color),
            (f"  {val}", color, True),
        ])


def cover_ch3_visual(svg):
    svg.add_text("  Archetypes", color='aqua', bold=True)
    svg.add_blank()
    svg.add_colored_spans([("  ◆ ", 'purple'), ("Architect", 'purple', True), ("  — Creates gravity, shapes codebase structure", 'fg_dim')])
    svg.add_colored_spans([("  ◆ ", 'blue'), ("Anchor", 'blue', True), ("     — Maintains orbit, stabilizes code", 'fg_dim')])
    svg.add_colored_spans([("  ◆ ", 'yellow'), ("Producer", 'yellow', True), ("   — Generates output, uses structure", 'fg_dim')])
    svg.add_colored_spans([("  ◆ ", 'aqua'), ("Cleaner", 'aqua', True), ("    — Reduces entropy, improves quality", 'fg_dim')])
    svg.add_blank()
    svg.add_colored_spans([("  Builder → Mass → Balanced → Spread → Churn → Rescue", 'fg_dim')])


def cover_ch4_visual(svg):
    svg.add_text("  Backend Architect Concentration", color='aqua', bold=True)
    svg.add_blank()
    svg.add_colored_spans([("  2024-H1  ", 'fg_dim'), ("★", 'purple'), (" Engineer F  ●", 'blue'), (" machuz", 'fg_dim')])
    svg.add_colored_spans([("  2024-H2  ", 'fg_dim'), ("★", 'purple'), (" Engineer F  ●", 'blue'), (" machuz (Builder)", 'blue')])
    svg.add_colored_spans([("  2025-H1  ", 'fg_dim'), ("●", 'blue'), (" Engineer F  ●", 'yellow'), (" machuz", 'fg_dim')])
    svg.add_colored_spans([("  2025-H2  ", 'fg_dim'), ("●", 'fg_dim'), (" Engineer F  ", 'fg_dim'), ("★", 'purple'), (" machuz (Architect)", 'purple', True)])
    svg.add_blank()
    svg.add_text("  The seat converges to one. Always.", color='fg_dim')


def cover_ch5_visual(svg):
    svg.add_text("  Timeline — Scores Don't Lie", color='aqua', bold=True)
    svg.add_blank()
    svg.add_colored_spans([("  Q1  ", 'fg_dim'), ("████████████████████", 'purple'), ("  90.0  Architect", 'purple')])
    svg.add_colored_spans([("  Q2  ", 'fg_dim'), ("██████████████████", 'green'), ("    72.5  Producer", 'yellow')])
    svg.add_colored_spans([("  Q3  ", 'fg_dim'), ("█████████████", 'yellow'), ("         49.1  Anchor", 'blue')])
    svg.add_colored_spans([("  Q4  ", 'fg_dim'), ("███████", 'red'), ("               31.2  — Fragile", 'red')])
    svg.add_colored_spans([("  Q1  ", 'fg_dim'), ("███", 'fg_dim'), ("                   11.3  — Former", 'fg_dim')])
    svg.add_blank()
    svg.add_text("  Hesitation leaves traces in the data.", color='fg_dim')


def cover_ch6_visual(svg):
    svg.add_text("  Team Evolution Models", color='aqua', bold=True)
    svg.add_blank()
    svg.add_colored_spans([("  Growing", 'blue'), (" → ", 'fg_dim'), ("Anchor", 'blue'), (" → ", 'fg_dim'), ("Producer", 'yellow'), (" → ", 'fg_dim'), ("Architect", 'purple', True)])
    svg.add_blank()
    svg.add_colored_spans([("  Permeation:  ", 'fg'), ("respect → produce → design", 'aqua')])
    svg.add_colored_spans([("  Immediate:   ", 'fg'), ("Anchor → Architect (rare)", 'purple')])
    svg.add_colored_spans([("  Founding:    ", 'fg'), ("Architect → decline = success", 'green')])
    svg.add_blank()
    svg.add_text("  Organizations have laws. Timelines reveal them.", color='fg_dim')


def cover_ch7_visual(svg):
    svg.add_text("  The Universe of Code", color='aqua', bold=True)
    svg.add_blank()
    svg.add_colored_spans([("  ◉ ", 'purple'), ("Gravity     ", 'fg'), ("— Architects create gravitational centers", 'fg_dim')])
    svg.add_colored_spans([("  ◉ ", 'green'), ("Strong Force", 'fg'), (" — Domain coupling holds modules together", 'fg_dim')])
    svg.add_colored_spans([("  ◉ ", 'blue'), ("Weak Force  ", 'fg'), (" — Cross-domain interactions and decay", 'fg_dim')])
    svg.add_colored_spans([("  ◉ ", 'yellow'), ("EM Force    ", 'fg'), (" — Communication patterns between members", 'fg_dim')])
    svg.add_blank()
    svg.add_text("  Codebases are universes. Engineers are celestial bodies.", color='fg_dim')


def cover_ch8_visual(svg):
    svg.add_text("  Engineering Relativity", color='aqua', bold=True)
    svg.add_blank()
    svg.add_colored_spans([("  Same engineer, different universes:", 'fg')])
    svg.add_blank()
    svg.add_colored_spans([("    Repo A  ", 'fg_dim'), ("Architectural Engine", 'purple'), ("  →  Total: ", 'fg'), ("35", 'yellow', True), ("  (hard-earned)", 'fg_dim')])
    svg.add_colored_spans([("    Repo B  ", 'fg_dim'), ("Unstructured", 'red'), ("          →  Total: ", 'fg'), ("60", 'green', True), ("  (easy)", 'fg_dim')])
    svg.add_blank()
    svg.add_text("  The gravity depends on the space it exists in.", color='fg_dim')


# ════════════════════════════════════════════════════════════
# MAIN
# ════════════════════════════════════════════════════════════

if __name__ == '__main__':
    print("Generating blog SVGs...")

    # Chapter 1
    ch1_backend_table()
    ch1_frontend_table()

    # Chapter 2
    ch2_warnings()

    # Chapter 3
    ch3_engineer_profiles()
    ch3_producer_warning()

    # Chapter 4
    ch4_backend_team()
    ch4_team_classification()
    ch4_structure()

    # Chapter 5
    ch5_engineer_f_timeline()
    ch5_engineer_j_timeline()
    ch5_engineer_i_timeline()
    ch5_per_repo_commits()
    ch5_transitions()
    ch5_comparison_table()
    ch5_team_timeline()

    # Chapter 6
    ch6_backend_team_timeline()
    ch6_backend_score_averages()
    ch6_frontend_team_timeline()
    ch6_infra_firmware()
    ch6_machuz_timeline()
    ch6_be_architects()
    ch6_fe_architects()
    ch6_engineer_k()
    ch6_gravity_transfer()
    ch6_evolution_paths()
    ch6_evolution_model()

    # Chapter 7
    ch7_team_classification()

    # Chapter 8
    ch8_repo_scores()
    ch8_structure_comparison()
    ch8_per_repo_breakdown()

    # Cover images
    cover_image(1, "Measuring Engineering Impact", "from Git History Alone", cover_ch1_visual)
    cover_image(2, "Beyond Individual Scores", "Measuring Team Health from Git History", cover_ch2_visual)
    cover_image(3, "Two Paths to Architect", "How Engineers Evolve Differently", cover_ch3_visual)
    cover_image(4, "Backend Architects Converge", "The Sacred Work of Laying Souls to Rest", cover_ch4_visual)
    cover_image(5, "Timeline: Scores Don't Lie", "And They Capture Hesitation Too", cover_ch5_visual)
    cover_image(6, "Teams Evolve", "The Laws of Organization Revealed by Timelines", cover_ch6_visual)
    cover_image(7, "Observing the Universe of Code", "Four Forces, Gravity, and Seasoned Design", cover_ch7_visual)
    cover_image(8, "Engineering Relativity", "Why the Same Engineer Gets Different Scores", cover_ch8_visual)

    print(f"\nDone! Generated SVGs in {IMG_DIR}")
