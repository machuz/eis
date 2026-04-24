package output

import (
	"encoding/json"
	"fmt"
	"io"
	"text/template"

	"github.com/machuz/eis/v2/internal/timeline"
)

// DomainTimelineData groups author timelines under a domain for HTML output.
type DomainTimelineData struct {
	DomainName string
	Span       string
	Timelines  []timeline.AuthorTimeline
}

// htmlAuthorPeriod is a JSON-serializable version of AuthorPeriod.
type htmlAuthorPeriod struct {
	Label      string  `json:"label"`
	Impact     float64 `json:"impact"`
	Production float64 `json:"production"`
	Quality    float64 `json:"quality"`
	Survival   float64 `json:"survival"`
	Design     float64 `json:"design"`
	Role       string  `json:"role"`
	RoleConf   float64 `json:"roleConf"`
	Style      string  `json:"style"`
	StyleConf  float64 `json:"styleConf"`
	State      string  `json:"state"`
	StateConf  float64 `json:"stateConf"`
	Active     bool    `json:"active"`
	Commits    int     `json:"commits"`
}

type htmlTransition struct {
	Axis     string `json:"axis"`
	From     string `json:"from"`
	To       string `json:"to"`
	AtPeriod string `json:"atPeriod"`
}

type htmlAuthorTimeline struct {
	Author      string             `json:"author"`
	Periods     []htmlAuthorPeriod  `json:"periods"`
	Transitions []htmlTransition   `json:"transitions"`
}

type htmlDomainData struct {
	DomainName string               `json:"domainName"`
	Span       string               `json:"span"`
	Timelines  []htmlAuthorTimeline  `json:"timelines"`
}

type htmlTeamPeriod struct {
	Label            string  `json:"label"`
	CoreMembers      int     `json:"coreMembers"`
	EffectiveMembers int     `json:"effectiveMembers"`
	TotalMembers     int     `json:"totalMembers"`
	AvgImpact        float64 `json:"avgImpact"`
	AvgProduction    float64 `json:"avgProduction"`
	AvgQuality       float64 `json:"avgQuality"`
	AvgSurvival      float64 `json:"avgSurvival"`
	AvgDesign        float64 `json:"avgDesign"`

	Complementarity     float64 `json:"complementarity"`
	GrowthPotential     float64 `json:"growthPotential"`
	Sustainability      float64 `json:"sustainability"`
	DebtBalance         float64 `json:"debtBalance"`
	ProductivityDensity float64 `json:"productivityDensity"`
	QualityConsistency  float64 `json:"qualityConsistency"`
	RiskRatio           float64 `json:"riskRatio"`

	Character string `json:"character"`
	Structure string `json:"structure"`
	Culture   string `json:"culture"`
	Phase     string `json:"phase"`
	Risk      string `json:"risk"`
}

type htmlTeamTimeline struct {
	TeamName    string             `json:"teamName"`
	Domain      string             `json:"domain"`
	Periods     []htmlTeamPeriod   `json:"periods"`
	Transitions []htmlTransition   `json:"transitions"`
}

type htmlTemplateData struct {
	JSONData string
}

type htmlJSONRoot struct {
	Domains []htmlDomainData   `json:"domains"`
	Teams   []htmlTeamTimeline `json:"teams"`
}

// WriteTimelineHTML generates a self-contained HTML file with Chart.js interactive graphs.
func WriteTimelineHTML(w io.Writer, domainTimelines []DomainTimelineData, teamTimelines []timeline.TeamTimeline) error {
	root := htmlJSONRoot{}

	for _, dt := range domainTimelines {
		hd := htmlDomainData{
			DomainName: dt.DomainName,
			Span:       dt.Span,
		}
		for _, tl := range dt.Timelines {
			hat := htmlAuthorTimeline{
				Author: tl.Author,
			}
			for _, p := range tl.Periods {
				hat.Periods = append(hat.Periods, htmlAuthorPeriod{
					Label:      p.Label,
					Impact:     p.Impact,
					Production: p.Production,
					Quality:    p.Quality,
					Survival:   p.Survival,
					Design:     p.Design,
					Role:       p.Role,
					RoleConf:   p.RoleConf,
					Style:      p.Style,
					StyleConf:  p.StyleConf,
					State:      p.State,
					StateConf:  p.StateConf,
					Active:     p.Active,
					Commits:    p.TotalCommits,
				})
			}
			for _, tr := range tl.Transitions {
				hat.Transitions = append(hat.Transitions, htmlTransition{
					Axis:     tr.Axis,
					From:     tr.From,
					To:       tr.To,
					AtPeriod: tr.AtPeriod,
				})
			}
			hd.Timelines = append(hd.Timelines, hat)
		}
		root.Domains = append(root.Domains, hd)
	}

	for _, tl := range teamTimelines {
		htt := htmlTeamTimeline{
			TeamName: tl.TeamName,
			Domain:   tl.Domain,
		}
		for _, p := range tl.Periods {
			htt.Periods = append(htt.Periods, htmlTeamPeriod{
				Label:               p.Label,
				CoreMembers:         p.CoreMembers,
				EffectiveMembers:    p.EffectiveMembers,
				TotalMembers:        p.TotalMembers,
				AvgImpact:           p.AvgImpact,
				AvgProduction:       p.AvgProduction,
				AvgQuality:          p.AvgQuality,
				AvgSurvival:         p.AvgSurvival,
				AvgDesign:           p.AvgDesign,
				Complementarity:     p.Complementarity,
				GrowthPotential:     p.GrowthPotential,
				Sustainability:      p.Sustainability,
				DebtBalance:         p.DebtBalance,
				ProductivityDensity: p.ProductivityDensity,
				QualityConsistency:  p.QualityConsistency,
				RiskRatio:           p.RiskRatio,
				Character:           p.Character,
				Structure:           p.Structure,
				Culture:             p.Culture,
				Phase:               p.Phase,
				Risk:                p.Risk,
			})
		}
		for _, tr := range tl.Transitions {
			htt.Transitions = append(htt.Transitions, htmlTransition{
				Axis:     tr.Axis,
				From:     tr.From,
				To:       tr.To,
				AtPeriod: tr.AtPeriod,
			})
		}
		root.Teams = append(root.Teams, htt)
	}

	jsonBytes, err := json.Marshal(root)
	if err != nil {
		return fmt.Errorf("marshal timeline data: %w", err)
	}

	tmpl, err := template.New("timeline").Parse(timelineHTMLTemplate)
	if err != nil {
		return fmt.Errorf("parse HTML template: %w", err)
	}

	return tmpl.Execute(w, htmlTemplateData{
		JSONData: string(jsonBytes),
	})
}

const timelineHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>EIS Timeline Report</title>
<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
<style>
  :root {
    --bg: #282828;
    --bg-card: #3c3836;
    --bg-nav: #1d2021;
    --text: #ebdbb2;
    --text-muted: #928374;
    --accent: #83a598;
    --border: #504945;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, monospace;
    background: var(--bg);
    color: var(--text);
    display: flex;
    min-height: 100vh;
  }
  nav {
    width: 260px;
    min-width: 260px;
    background: var(--bg-nav);
    padding: 20px 0;
    overflow-y: auto;
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    z-index: 10;
  }
  nav h2 {
    padding: 0 16px 12px;
    font-size: 14px;
    color: var(--accent);
    text-transform: uppercase;
    letter-spacing: 1px;
  }
  nav a {
    display: block;
    padding: 6px 16px 6px 24px;
    color: var(--text-muted);
    text-decoration: none;
    font-size: 13px;
    transition: background 0.15s, color 0.15s;
  }
  nav a:hover { background: rgba(131,165,152,0.15); color: #ebdbb2; }
  nav .nav-section {
    padding: 10px 16px 4px;
    font-size: 12px;
    color: var(--accent);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    font-weight: 600;
  }
  main {
    margin-left: 260px;
    flex: 1;
    padding: 30px 40px;
    max-width: 1200px;
  }
  h1 {
    font-size: 24px;
    margin-bottom: 8px;
    color: #ebdbb2;
  }
  .subtitle {
    color: var(--text-muted);
    font-size: 13px;
    margin-bottom: 30px;
  }
  .domain-section { margin-bottom: 40px; }
  .domain-title {
    font-size: 18px;
    color: var(--accent);
    border-bottom: 1px solid var(--border);
    padding-bottom: 8px;
    margin-bottom: 20px;
  }
  .chart-card {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 20px;
    margin-bottom: 20px;
  }
  .chart-card h3 {
    font-size: 15px;
    margin-bottom: 12px;
    color: #ebdbb2;
  }
  .chart-container {
    position: relative;
    height: 300px;
  }
  .chart-container.bar-chart {
    height: 220px;
  }
  .transitions {
    margin-top: 10px;
    padding-top: 10px;
    border-top: 1px solid var(--border);
  }
  .transition-item {
    display: inline-block;
    font-size: 12px;
    padding: 3px 8px;
    margin: 2px 4px 2px 0;
    border-radius: 4px;
    background: rgba(131,165,152,0.15);
    color: var(--accent);
  }
  .classification-row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 10px;
  }
  .badge {
    display: inline-block;
    font-size: 11px;
    padding: 3px 10px;
    border-radius: 12px;
    font-weight: 600;
  }
  .badge-character { background: #8ec07c33; color: #8ec07c; }
  .badge-structure { background: #d3869b33; color: #d3869b; }
  .badge-culture   { background: #fe801933; color: #fe8019; }
  .badge-phase     { background: #83a59833; color: #83a598; }
  .badge-risk      { background: #fb493433; color: #fb4934; }
  .class-period {
    background: var(--bg);
    border-radius: 6px;
    padding: 8px 12px;
    margin-bottom: 6px;
  }
  .class-period-label {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-muted);
    margin-bottom: 4px;
  }
  @media (max-width: 768px) {
    nav { display: none; }
    main { margin-left: 0; padding: 16px; }
  }
</style>
</head>
<body>

<nav id="sidebar">
  <h2>EIS Timeline</h2>
</nav>

<main>
  <h1>EIS Timeline Report</h1>
  <p class="subtitle">Interactive timeline analysis of engineering impact scores</p>
  <div id="content"></div>
</main>

<script>
const DATA = {{.JSONData}};

const COLORS = {
  total:      '#83a598',
  production: '#b8bb26',
  quality:    '#fabd2f',
  survival:   '#d3869b',
  design:     '#8ec07c'
};

const HEALTH_COLORS = [
  '#83a598', '#b8bb26', '#fabd2f', '#d3869b',
  '#8ec07c', '#fb4934', '#fe8019'
];

const MEMBER_COLORS = {
  core:      '#83a598',
  effective: '#b8bb26',
  total:     '#fabd2f'
};

function makeId(str) {
  return str.replace(/[^a-zA-Z0-9]/g, '-').toLowerCase();
}

function buildNav() {
  const nav = document.getElementById('sidebar');
  let html = '<h2>EIS Timeline</h2>';

  if (DATA.domains && DATA.domains.length > 0) {
    html += '<div class="nav-section">Individual</div>';
    for (const d of DATA.domains) {
      html += '<a href="#domain-' + makeId(d.domainName) + '" style="font-weight:600;color:#ebdbb2;padding-left:16px">' + d.domainName + '</a>';
      for (const tl of d.timelines) {
        html += '<a href="#author-' + makeId(d.domainName + '-' + tl.author) + '">' + tl.author + '</a>';
      }
    }
  }

  if (DATA.teams && DATA.teams.length > 0) {
    html += '<div class="nav-section">Teams</div>';
    for (const t of DATA.teams) {
      html += '<a href="#team-' + makeId(t.teamName) + '">' + t.teamName + '</a>';
    }
  }

  nav.innerHTML = html;
}

function createScoreChart(canvasId, labels, periods, transitions) {
  const ctx = document.getElementById(canvasId).getContext('2d');

  const datasets = [
    { label: 'Impact',     data: periods.map(p => p.impact),     borderColor: COLORS.total,      backgroundColor: COLORS.total + '20' },
    { label: 'Production', data: periods.map(p => p.production), borderColor: COLORS.production, backgroundColor: COLORS.production + '20' },
    { label: 'Quality',    data: periods.map(p => p.quality),    borderColor: COLORS.quality,    backgroundColor: COLORS.quality + '20' },
    { label: 'Survival',   data: periods.map(p => p.survival),   borderColor: COLORS.survival,   backgroundColor: COLORS.survival + '20' },
    { label: 'Design',     data: periods.map(p => p.design),     borderColor: COLORS.design,     backgroundColor: COLORS.design + '20' },
  ];

  datasets.forEach(ds => {
    ds.tension = 0.3;
    ds.borderWidth = 2;
    ds.pointRadius = 4;
    ds.pointHoverRadius = 6;
    ds.fill = false;
  });

  // Build transition annotations
  const annotations = {};
  if (transitions) {
    transitions.forEach((tr, i) => {
      const idx = labels.indexOf(tr.atPeriod);
      if (idx >= 0) {
        annotations['tr' + i] = {
          type: 'line',
          xMin: idx,
          xMax: idx,
          borderColor: '#ebdbb244',
          borderWidth: 1,
          borderDash: [5, 5],
          label: {
            display: true,
            content: tr.axis + ': ' + tr.from + ' -> ' + tr.to,
            position: 'start',
            backgroundColor: '#3c3836',
            color: '#ebdbb2',
            font: { size: 10 }
          }
        };
      }
    });
  }

  const plugins = {};
  if (Object.keys(annotations).length > 0) {
    plugins.annotation = { annotations: annotations };
  }

  new Chart(ctx, {
    type: 'line',
    data: { labels: labels, datasets: datasets },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        y: {
          min: 0, max: 100,
          grid: { color: '#504945' },
          ticks: { color: '#928374' }
        },
        x: {
          grid: { color: '#504945' },
          ticks: { color: '#928374' }
        }
      },
      plugins: {
        legend: {
          labels: { color: '#ebdbb2', usePointStyle: true, pointStyle: 'circle' }
        },
        tooltip: {
          callbacks: {
            afterBody: function(context) {
              const idx = context[0].dataIndex;
              const p = periods[idx];
              const lines = [];
              if (p.role) lines.push('Role: ' + p.role + ' (' + (p.roleConf * 100).toFixed(0) + '%)');
              if (p.style) lines.push('Style: ' + p.style + ' (' + (p.styleConf * 100).toFixed(0) + '%)');
              if (p.state) lines.push('State: ' + p.state + ' (' + (p.stateConf * 100).toFixed(0) + '%)');
              if (p.commits) lines.push('Commits: ' + p.commits);
              lines.push('Active: ' + (p.active ? 'Yes' : 'No'));
              return lines;
            }
          }
        },
        ...plugins
      }
    }
  });
}

function createTeamScoreChart(canvasId, labels, periods) {
  const ctx = document.getElementById(canvasId).getContext('2d');
  const datasets = [
    { label: 'AvgImpact',     data: periods.map(p => p.avgImpact),     borderColor: COLORS.total },
    { label: 'AvgProduction', data: periods.map(p => p.avgProduction), borderColor: COLORS.production },
    { label: 'AvgQuality',    data: periods.map(p => p.avgQuality),    borderColor: COLORS.quality },
    { label: 'AvgSurvival',   data: periods.map(p => p.avgSurvival),   borderColor: COLORS.survival },
    { label: 'AvgDesign',     data: periods.map(p => p.avgDesign),     borderColor: COLORS.design },
  ];
  datasets.forEach(ds => {
    ds.tension = 0.3;
    ds.borderWidth = 2;
    ds.pointRadius = 4;
    ds.fill = false;
  });

  new Chart(ctx, {
    type: 'line',
    data: { labels, datasets },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        y: { min: 0, max: 100, grid: { color: '#504945' }, ticks: { color: '#928374' } },
        x: { grid: { color: '#504945' }, ticks: { color: '#928374' } }
      },
      plugins: {
        legend: { labels: { color: '#ebdbb2', usePointStyle: true, pointStyle: 'circle' } }
      }
    }
  });
}

function createHealthChart(canvasId, labels, periods) {
  const ctx = document.getElementById(canvasId).getContext('2d');
  const metrics = [
    { key: 'complementarity',     label: 'Complementarity' },
    { key: 'growthPotential',     label: 'GrowthPotential' },
    { key: 'sustainability',      label: 'Sustainability' },
    { key: 'debtBalance',         label: 'DebtBalance' },
    { key: 'productivityDensity', label: 'ProductivityDensity' },
    { key: 'qualityConsistency',  label: 'QualityConsistency' },
    { key: 'riskRatio',           label: 'RiskRatio' },
  ];
  const datasets = metrics.map((m, i) => ({
    label: m.label,
    data: periods.map(p => p[m.key]),
    borderColor: HEALTH_COLORS[i],
    tension: 0.3,
    borderWidth: 2,
    pointRadius: 3,
    fill: false,
  }));

  new Chart(ctx, {
    type: 'line',
    data: { labels, datasets },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        y: { min: 0, grid: { color: '#504945' }, ticks: { color: '#928374' } },
        x: { grid: { color: '#504945' }, ticks: { color: '#928374' } }
      },
      plugins: {
        legend: { labels: { color: '#ebdbb2', usePointStyle: true, pointStyle: 'circle' } }
      }
    }
  });
}

function createMembershipChart(canvasId, labels, periods) {
  const ctx = document.getElementById(canvasId).getContext('2d');
  new Chart(ctx, {
    type: 'bar',
    data: {
      labels,
      datasets: [
        { label: 'Core',      data: periods.map(p => p.coreMembers),      backgroundColor: MEMBER_COLORS.core },
        { label: 'Effective', data: periods.map(p => p.effectiveMembers), backgroundColor: MEMBER_COLORS.effective },
        { label: 'Total',     data: periods.map(p => p.totalMembers),     backgroundColor: MEMBER_COLORS.total },
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        y: { beginAtZero: true, grid: { color: '#504945' }, ticks: { color: '#928374', stepSize: 1 } },
        x: { grid: { color: '#504945' }, ticks: { color: '#928374' } }
      },
      plugins: {
        legend: { labels: { color: '#ebdbb2', usePointStyle: true, pointStyle: 'rect' } }
      }
    }
  });
}

function renderClassification(periods) {
  let html = '';
  for (const p of periods) {
    html += '<div class="class-period">';
    html += '<div class="class-period-label">' + p.label + '</div>';
    html += '<div class="classification-row">';
    if (p.character) html += '<span class="badge badge-character">Character: ' + p.character + '</span>';
    if (p.structure) html += '<span class="badge badge-structure">Structure: ' + p.structure + '</span>';
    if (p.culture)   html += '<span class="badge badge-culture">Culture: ' + p.culture + '</span>';
    if (p.phase)     html += '<span class="badge badge-phase">Phase: ' + p.phase + '</span>';
    if (p.risk)      html += '<span class="badge badge-risk">Risk: ' + p.risk + '</span>';
    html += '</div></div>';
  }
  return html;
}

function renderTransitions(transitions) {
  if (!transitions || transitions.length === 0) return '';
  let html = '<div class="transitions">';
  for (const t of transitions) {
    html += '<span class="transition-item">[' + t.atPeriod + '] ' + t.axis + ': ' + t.from + ' &rarr; ' + t.to + '</span>';
  }
  html += '</div>';
  return html;
}

function render() {
  buildNav();
  const content = document.getElementById('content');
  let html = '';
  let chartInits = [];
  let chartIdx = 0;

  // Individual timelines
  if (DATA.domains) {
    for (const d of DATA.domains) {
      const domainId = 'domain-' + makeId(d.domainName);
      html += '<div class="domain-section" id="' + domainId + '">';
      html += '<h2 class="domain-title">' + d.domainName + ' Timeline (' + d.span + ' spans)</h2>';

      for (const tl of d.timelines) {
        const authorId = 'author-' + makeId(d.domainName + '-' + tl.author);
        const canvasId = 'chart-' + chartIdx++;
        html += '<div class="chart-card" id="' + authorId + '">';
        html += '<h3>' + tl.author + '</h3>';
        html += '<div class="chart-container"><canvas id="' + canvasId + '"></canvas></div>';
        html += renderTransitions(tl.transitions);
        html += '</div>';

        const labels = tl.periods.map(p => p.label);
        chartInits.push(() => createScoreChart(canvasId, labels, tl.periods, tl.transitions));
      }

      html += '</div>';
    }
  }

  // Team timelines
  if (DATA.teams && DATA.teams.length > 0) {
    for (const t of DATA.teams) {
      const teamId = 'team-' + makeId(t.teamName);
      html += '<div class="domain-section" id="' + teamId + '">';
      html += '<h2 class="domain-title">' + t.teamName + ' / ' + t.domain + ' &mdash; Team Timeline</h2>';

      const labels = t.periods.map(p => p.label);

      // Score averages
      const scoreId = 'chart-' + chartIdx++;
      html += '<div class="chart-card"><h3>Score Averages</h3>';
      html += '<div class="chart-container"><canvas id="' + scoreId + '"></canvas></div></div>';
      chartInits.push(() => createTeamScoreChart(scoreId, labels, t.periods));

      // Health metrics
      const healthId = 'chart-' + chartIdx++;
      html += '<div class="chart-card"><h3>Health Metrics</h3>';
      html += '<div class="chart-container"><canvas id="' + healthId + '"></canvas></div></div>';
      chartInits.push(() => createHealthChart(healthId, labels, t.periods));

      // Membership
      const memberId = 'chart-' + chartIdx++;
      html += '<div class="chart-card"><h3>Membership</h3>';
      html += '<div class="chart-container bar-chart"><canvas id="' + memberId + '"></canvas></div></div>';
      chartInits.push(() => createMembershipChart(memberId, labels, t.periods));

      // Classification
      html += '<div class="chart-card"><h3>Classification</h3>';
      html += renderClassification(t.periods);
      html += renderTransitions(t.transitions);
      html += '</div>';

      html += '</div>';
    }
  }

  content.innerHTML = html;

  // Initialize all charts after DOM is updated
  requestAnimationFrame(() => {
    chartInits.forEach(fn => fn());
  });
}

document.addEventListener('DOMContentLoaded', render);
</script>
</body>
</html>
`
