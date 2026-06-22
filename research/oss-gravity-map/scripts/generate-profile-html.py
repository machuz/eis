#!/usr/bin/env python3
"""Generate a per-engineer Gravity Profile page (bilingual) from EIS data.

Renders the same shape as the hand-built scala-xuwei-k profile, but driven by
data: static axes from data/results/<repo>.json and the per-period arc from
data/results/<repo>-timeline.json. EIS taxonomy (Gravity, Design, Catalysis…)
stays in English; only prose is localized. Gravity is presented as a
present-tense, time-dependent reading — never an eternal verdict.

Usage:
    python generate-profile-html.py <profile-key> <lang> <output.html>
    python generate-profile-html.py --all        # writes every profile, EN+JA
"""
import json, sys, os, html

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

def esc(s): return html.escape(str(s))

def first_object(path):
    raw = open(path, encoding='utf-8').read()
    depth=0; ins=False; esc_=False; end=0
    for i,c in enumerate(raw):
        if esc_: esc_=False; continue
        if c=='\\' and ins: esc_=True; continue
        if c=='"': ins=not ins; continue
        if ins: continue
        if c=='{': depth+=1
        elif c=='}':
            depth-=1
            if depth==0: end=i+1; break
    return json.loads(raw[:end])

# ---------------------------------------------------------------------------
# Profiles: famous developers of famous OSS, chosen so a reader who knows them
# can calibrate trust in the metric.
# ---------------------------------------------------------------------------
PROFILES = {
 'react-abramov': {
   'repo':'react','author':'Dan Abramov','repo_label':'facebook/react','color':'#61dafb',
   'tagline':{'en':'Famous Is Not the Same as Load-Bearing','ja':'有名であることと、荷重を負うことは違う'},
   'hero':{
     'en':"Dan Abramov is one of the most recognised names in front-end engineering — Redux, the React docs, years of teaching. He is a long, prolific presence in react itself. Yet his Gravity reads low today: much of his most-felt work was API shape, documentation and advocacy rather than react-core structure that others still press on. His timeline shows a real peak as a Cleaner around 2018, fading since. Gravity is a present-tense reading of structural load — not a measure of fame or influence.",
     'ja':"Dan Abramov はフロントエンド界で最も知られた名前の一人だ — Redux、React 公式ドキュメント、長年の発信。react 本体でも多作で息の長い存在である。それでも、いまの Gravity は低い。最も影響を残した仕事の多くが、react コアの構造そのものより、API の形・ドキュメント・啓発だったからだ。timeline は 2018 年前後に Cleaner として確かなピークを持ち、その後フェードしていく弧を描く。Gravity は構造的な荷重の現在形の観測であって、知名度や影響力の尺度ではない。"},
   'insight':{
     'en':"This is the clearest calibration case in the map. A globally famous engineer, deeply tied to React's story, reads as low current Gravity — because the metric asks one narrow question: whose code does the surviving structure still lean on, while others actively press on it? Abramov's peak was real and the timeline records it; today the load sits elsewhere. None of that diminishes the influence — it just is not what Gravity measures.",
     'ja':"これは地図の中で最も分かりやすい較正例だ。React の物語と深く結びついた世界的に著名なエンジニアが、いまは低い Gravity として読み取られる — この指標がただ一つの狭い問いを立てるからだ：他者が能動的に圧をかけ続けるなか、生き残った構造はいまも誰のコードに寄りかかっているか。Abramov のピークは本物で、timeline はそれを記録している。いま荷重は別の場所にある。それは影響力を貶めない — ただ Gravity が測っているものではない、というだけだ。"},
 },
 'rails-dhh': {
   'repo':'rails','author':'David Heinemeier Hansson','repo_label':'rails/rails','color':'#cc0000',
   'tagline':{'en':'The Founder’s Signature, the Faded Load','ja':'創設者の署名、薄れた荷重'},
   'hero':{
     'en':"David Heinemeier Hansson created Rails. His signature is still on the shape of the system — Design 100, Indispensability 100, the founder's marks. Yet his Gravity reads a quiet figure today, and the state is Silent. Rails is a dispersed, multi-architect civilisation; the founder's modules are revered but largely undisturbed, and the telescope does not confuse that stillness with the system leaning on him now.",
     'ja':"David Heinemeier Hansson は Rails を生んだ。システムの形にはいまも彼の署名が残る — Design 100、Indispensability 100、創設者の刻印だ。それでも、いまの Gravity は静かな数値で、状態は Silent。Rails は分散した多アーキテクトの文明であり、創設者のモジュールは敬われつつもほとんど揺らされない。望遠鏡は、その静けさを「いまシステムが彼に寄りかかっている」ことと取り違えない。"},
   'insight':{
     'en':"A founder's centrality can fade not by being displaced, but simply by no one needing to lean on it anymore. DHH's Design and Indispensability still read 100 — the structure he laid is recognisably his — yet relational Gravity stays low because the live pressure has moved to many other hands. That a creator's Gravity is modest is not a verdict on him; it is the instrument declining to read founding, or seniority, as present structural load.",
     'ja':"創設者の中心性は、押しのけられて薄れるのではなく、ただ誰もそれに寄りかかる必要がなくなって薄れていくこともある。DHH の Design と Indispensability はいまも 100 — 彼が据えた構造は紛れもなく彼のものだ — それでも relational Gravity が低いのは、生きた圧が他の多くの手に移ったからだ。創設者の Gravity が控えめであることは、彼への判定ではない。創設や年功を「いまの構造的荷重」として読むことを、計器が拒んでいるのだ。"},
 },
 'scala-odersky': {
   'repo':'scala','author':'Martin Odersky','repo_label':'scala/scala','color':'#8b5cf6',
   'tagline':{'en':'Caught at the Founding, Faded by Design','ja':'創成期に捉えられ、設計どおりに薄れる'},
   'hero':{
     'en':"Martin Odersky designed Scala. The timeline catches him exactly where you would expect: a near-maximal Architect Gravity in the founding era, decaying as the language matured and as Scala 3 drew the frontier elsewhere. His static Gravity today is modest but still architect-shaped (Design high, Architect role). The arc is the point — the instrument credited the creator at his peak, and faded him faithfully as others rebuilt on top.",
     'ja':"Martin Odersky は Scala を設計した。timeline は、まさに予想どおりの場所で彼を捉える — 創成期にほぼ最大の Architect Gravity を示し、言語が成熟し、Scala 3 が最前線を別の場所へ移すにつれて減衰していく。いまの静的 Gravity は控えめだが、なお architect の形を保つ(Design 高・Architect ロール)。弧こそが要点だ — 計器は創造者をピークで credit し、他者がその上に築き直すにつれ忠実にフェードさせた。"},
   'insight':{
     'en':"Founders are the hardest test of a relational metric: an honest lens must credit them when their code load-bears and let them fade when it is superseded — without ever erasing that they were once central. Odersky's profile does exactly this. The founding peak is unmistakable; the present-day decline is not a demotion but a reading of where the live structure now rests.",
     'ja':"創設者は relational な指標にとって最も厳しい試金石だ — 正直なレンズは、彼らのコードが荷重を負う時代には credit し、それが superseded された後にはフェードさせねばならない。しかも、かつて中心だった事実を消し去ることなく。Odersky のプロファイルはまさにそれを行う。創成期のピークは紛れもなく、現在の減衰は降格ではなく、生きた構造がいまどこに宿るかの観測である。"},
 },
 'grafana-torkel': {
   'repo':'grafana','author':'Torkel Ödegaard','repo_label':'grafana/grafana','color':'#f59e0b',
   'tagline':{'en':'The Founder the Timeline Remembers','ja':'timeline が記憶している創設者'},
   'hero':{
     'en':"Torkel Ödegaard created Grafana and, for years, was its structural centre. The timeline records a near-total Architect Gravity in the mid-2010s — and then a long, honest fade to a quiet figure today, with the state reading Former. He remains one of the most prolific committers in the project, yet the surviving load has moved to the organisation that grew around him.",
     'ja':"Torkel Ödegaard は Grafana を生み、長年その構造的中心だった。timeline は 2010 年代半ばにほぼ最大の Architect Gravity を記録し — そこから今日の静かな数値まで、長く正直にフェードしていく。状態は Former だ。彼はいまもプロジェクト屈指の多作なコミッターだが、生き残った荷重は、彼のまわりに育った組織へと移った。"},
   'insight':{
     'en':"Grafana is one of the cleanest demonstrations that Gravity is time-faithful. Independent ground-truth calls Torkel a founder; the static metric ranks him modestly today; and the per-period timeline shows precisely why — he was the architect, at near-maximum, in his era. The low number now is not a miss. It is the same needle, pointed at the present.",
     'ja':"Grafana は、Gravity が時間に忠実であることの最も分かりやすい実証の一つだ。独立した ground-truth は Torkel を創設者と呼び、静的な指標はいまの彼を控えめに位置づける — そして per-period の timeline が、その理由を正確に示す。彼はその時代に、ほぼ最大の architect だった。いまの低い数値は miss ではない。同じ針が、現在を指しているだけだ。"},
 },
 'envoy-klein': {
   'repo':'envoy','author':'Matt Klein','repo_label':'envoyproxy/envoy','color':'#ec4899',
   'tagline':{'en':'Anchored Early, Handed Over','ja':'初期に錨を下ろし、引き継いだ'},
   'hero':{
     'en':"Matt Klein started Envoy at Lyft and anchored its early structure. The timeline shows him as an Anchor at his peak in the mid-2010s, then fading as Envoy became a large, CNCF-governed, many-hands project. His static Gravity is modest today and the state is Silent — the founder's load distributed across the community that took the project on.",
     'ja':"Matt Klein は Lyft で Envoy を立ち上げ、その初期構造に錨を下ろした。timeline は 2010 年代半ばのピークで彼を Anchor として示し、Envoy が CNCF 統治の大規模・多人数プロジェクトになるにつれてフェードしていく。いまの静的 Gravity は控えめで、状態は Silent — 創設者の荷重は、プロジェクトを引き受けたコミュニティ全体に分散した。"},
   'insight':{
     'en':"A healthy handover looks like this from the outside: a founder who anchored the early structure, and a community that now carries the load. Gravity reads that transition honestly — a real early Anchor peak, a faded present. The metric is not asking who built it; it is asking who the structure leans on, now, under live pressure.",
     'ja':"健全な引き継ぎは、外から見るとこう映る — 初期構造に錨を下ろした創設者と、いま荷重を担うコミュニティ。Gravity はその移行を正直に読む — 初期の確かな Anchor ピークと、フェードした現在。この指標は「誰が作ったか」を問うていない。「生きた圧の下で、いま構造が誰に寄りかかっているか」を問うている。"},
 },
}

ORDER = ['react-abramov','rails-dhh','scala-odersky','grafana-torkel','envoy-klein']

S = {
 'en':{'html_lang':'en','title_suffix':'Gravity Profile','scope_pre':'Scope:',
   'scope':lambda rl:f"this profile measures only commits to <code>{esc(rl)}</code>. Contributions to the wider ecosystem (other repositories, libraries and tooling) lie outside this measurement.",
   's_gravity':'Gravity','s_design':'Design','s_indisp':'Indispensability','s_commits':'Commits','s_span':'Active span',
   'tl_title':'Contribution Timeline','tl_desc':'Per-period signals across the project history. Gravity, Catalysis and Debt Cleanup on a 0–100 axis; commits as bars (volume).',
   'tl_callout_pre':'Read the arc:','radar_title':'Axis Profile','ctx_title':'Where the Gravity Sits',
   'ctx_desc':'This engineer against the project’s highest-Gravity contributors. A field, not a contest.',
   'method_title':'Methodology','you':'this engineer',
   'method':[
     "<strong>Gravity</strong> = catGate(Catalysis) × survGate(RobustSurvival) × shape(0.45·Design + 0.25·Breadth + 0.30·Indispensability). A measure of durable, load-bearing structure others build on — not commit count.",
     "<strong>Gravity is a present-tense reading.</strong> It reflects the relationship at a moment in time; the same needle moves a founder from a near-maximum peak down to a few points as the live load moves elsewhere. A low figure today is not a permanent verdict.",
     "<strong>Catalysis / RobustSurvival</strong> are the two gates. RobustSurvival counts only lines that endure in modules where <em>other</em> authors keep committing; Catalysis asks whether others built on the surviving foundation. If either is near zero, the product collapses — ownership or founding alone cannot mint Gravity.",
   ],
   'commits':'Commits','gravity':'Gravity','catalysis':'Catalysis','debt':'Debt Cleanup','metric':'Metric',
 },
 'ja':{'html_lang':'ja','title_suffix':'Gravity Profile','scope_pre':'対象範囲:',
   'scope':lambda rl:f"このプロファイルは <code>{esc(rl)}</code> へのコミットのみを測定する。より広いエコシステム(他リポジトリ・ライブラリ・ツール)への貢献は、この測定の範囲外である。",
   's_gravity':'Gravity','s_design':'Design','s_indisp':'Indispensability','s_commits':'コミット','s_span':'活動期間',
   'tl_title':'貢献の Timeline','tl_desc':'プロジェクト史を通じた期間ごとのシグナル。Gravity・Catalysis・Debt Cleanup は 0〜100 軸、コミットはバー(量)。',
   'tl_callout_pre':'弧を読む:','radar_title':'軸プロファイル','ctx_title':'Gravity がどこに宿るか',
   'ctx_desc':'このエンジニアを、プロジェクトで最も Gravity の高い貢献者と並べて見る。競争ではなく、場として。',
   'method_title':'Methodology','you':'このエンジニア',
   'method':[
     "<strong>Gravity</strong> = catGate(Catalysis) × survGate(RobustSurvival) × shape(0.45·Design + 0.25·Breadth + 0.30·Indispensability)。他者が上に築く、耐久的で荷重を負う構造の指標 — コミット数ではない。",
     "<strong>Gravity は現在形の観測だ。</strong>その時々の関係を映す。同じ針が、生きた荷重が他へ移るにつれ、創設者をほぼ最大のピークから数ポイントへと動かす。いまの低い数値は永続的な評価ではない。",
     "<strong>Catalysis / RobustSurvival</strong> が二つのゲートだ。RobustSurvival は、<em>他者</em>がコミットし続けるモジュールで残った行だけを数える。Catalysis は、他者がその生き残った基盤の上に築いたかを問う。どちらかが 0 に近ければ積は崩れる — 所有や創設だけでは Gravity は鋳造できない。",
   ],
   'commits':'Commits','gravity':'Gravity','catalysis':'Catalysis','debt':'Debt Cleanup','metric':'Metric',
 },
}

ROLE_COLORS={'Architect':'#f97583','Anchor':'#79c0ff','Producer':'#56d364','Cleaner':'#d2a8ff','Specialist':'#ffa657'}

def fmt(v): return f"{v:.0f}" if abs(v-round(v))<0.05 else f"{v:.1f}"

def get_static(repo, author):
    doms=json.load(open(os.path.join(ROOT,f'data/results/{repo}.json')))['domains']
    ms=[m for d in doms for m in d['members']]
    me=next((m for m in ms if (m.get('member') or '')==author), None)
    ms_sorted=sorted(ms,key=lambda m:-(m.get('gravity',0)or 0))
    rank=ms_sorted.index(me)+1 if me in ms_sorted else None
    return me, ms_sorted, rank, len(ms)

def get_timeline(repo, author):
    p=os.path.join(ROOT,f'data/results/{repo}-timeline.json')
    if not os.path.exists(p) or os.path.getsize(p)==0: return None
    tl=first_object(p); labels=[x['label'] for x in tl['periods']]
    a=next((x for x in tl['authors'] if x['author']==author), None)
    if not a: return None
    per=a['periods']
    # trim to active range (first..last period with a commit)
    act=[i for i,pp in enumerate(per) if (pp.get('commits',0) or 0)>0]
    if not act: return None
    lo,hi=act[0],act[-1]
    sl=slice(lo,hi+1)
    return {'labels':labels[sl],
            'commits':[(per[i].get('commits',0)or 0) for i in range(lo,hi+1)],
            'gravity':[(per[i].get('gravity',0)or 0) for i in range(lo,hi+1)],
            'catalysis':[(per[i].get('catalysis',0)or 0) for i in range(lo,hi+1)],
            'debt':[(per[i].get('debt_cleanup',0)or 0) for i in range(lo,hi+1)],
            'roles':[per[i].get('role') for i in range(lo,hi+1)]}

def y100(v,base=290,h=260): return base-(v/100.0)*h

def timeline_svg(tl, lang):
    L=S[lang]; labels=tl['labels']; n=len(labels)
    W=900; x0=50; x1=880
    xs=[x0+(x1-x0)*(i/(n-1) if n>1 else 0.5) for i in range(n)]
    out=[f'<svg width="{W}" height="340" viewBox="0 0 {W} 340">',
         f'<rect width="{W}" height="340" fill="#161b22" rx="8"/>']
    for gv,yy in [(100,30),(75,95),(50,160),(25,225),(0,290)]:
        out.append(f'<line x1="{x0}" y1="{yy}" x2="{x1}" y2="{yy}" stroke="#21262d" stroke-width="1"/>')
        out.append(f'<text x="42" y="{yy}" fill="#484f58" font-size="10" text-anchor="end" dominant-baseline="middle">{gv}</text>')
    # x labels (thin if many)
    step=max(1,n//14)
    for i,lb in enumerate(labels):
        if i%step==0 or i==n-1:
            out.append(f'<text x="{xs[i]:.1f}" y="330" fill="#484f58" font-size="10" text-anchor="middle">{esc(lb)}</text>')
    # commit bars (normalized)
    cmax=max(tl['commits']) or 1
    for i,c in enumerate(tl['commits']):
        bh=(c/cmax)*78
        out.append(f'<rect x="{xs[i]-8:.1f}" y="{290-bh:.1f}" width="16" height="{bh:.1f}" fill="#42a5f5" fill-opacity="0.15" rx="2"/>')
        if c>0 and (i%step==0 or i==n-1 or c==cmax):
            out.append(f'<text x="{xs[i]:.1f}" y="{290-bh-4:.1f}" fill="#42a5f5" font-size="9" text-anchor="middle" fill-opacity="0.7">{c}</text>')
    # metric lines (Gravity / Catalysis / Debt on the 0-100 axis)
    for key,col in [('gravity','#ffd700'),('catalysis','#00e676'),('debt','#ab47bc')]:
        d='M '+' L '.join(f"{xs[i]:.1f} {y100(v):.1f}" for i,v in enumerate(tl[key]))
        out.append(f'<path d="{d}" fill="none" stroke="{col}" stroke-width="2" stroke-opacity="0.85"/>')
        for i,v in enumerate(tl[key]):
            out.append(f'<circle cx="{xs[i]:.1f}" cy="{y100(v):.1f}" r="3.2" fill="{col}"><title>{esc(labels[i])} {key}: {fmt(v)}</title></circle>')
    # legend
    leg=[('#42a5f5',L['commits']),('#ffd700',L['gravity']),('#00e676',L['catalysis']),('#ab47bc',L['debt'])]
    ly=40
    for col,lab in leg:
        out.append(f'<rect x="60" y="{ly-5}" width="12" height="3" fill="{col}" rx="1"/>')
        out.append(f'<text x="76" y="{ly}" fill="{col}" font-size="10" dominant-baseline="middle">{esc(lab)}</text>')
        ly+=16
    out.append('</svg>')
    # table
    th=''.join(f'<th>{esc(l)}</th>' for l in labels)
    def row(lab,vals,fmtfn=lambda v:fmt(v)):
        tds=''.join(f'<td>{fmtfn(v)}</td>' for v in vals)
        return f'<tr><td style="text-align:left;color:#8b949e;font-weight:600">{esc(lab)}</td>{tds}</tr>'
    tbl=(f'<table class="timeline-table"><thead><tr><th>{esc(L["metric"])}</th>{th}</tr></thead><tbody>'
         + row(L['commits'],tl['commits'],lambda v:str(v))
         + row(L['gravity'],tl['gravity'])
         + row(L['catalysis'],tl['catalysis'])
         + row(L['debt'],tl['debt'])
         + '</tbody></table>')
    return '\n'.join(out), tbl

def radar_svg(me, color):
    axes=[('Design','design'),('Indisp.','indispensability'),('Catalysis','catalysis'),
          ('RobustSurv','robust_survival'),('Breadth','breadth'),('Debt','debt_cleanup')]
    import math
    cx=cy=110; R=80; N=len(axes)
    def pt(i,r):
        ang=-math.pi/2+2*math.pi*i/N
        return cx+r*math.cos(ang)*0.0+ (r)*math.cos(ang), cy+(r)*math.sin(ang)
    out=[f'<svg width="220" height="220" viewBox="0 0 220 220">']
    for ring in (0.25,0.5,0.75,1.0):
        poly=' '.join(f"{cx+R*ring*math.cos(-math.pi/2+2*math.pi*i/N):.1f},{cy+R*ring*math.sin(-math.pi/2+2*math.pi*i/N):.1f}" for i in range(N))
        out.append(f'<polygon points="{poly}" fill="none" stroke="#30363d" stroke-width="0.5"/>')
    for i,(lab,key) in enumerate(axes):
        ex=cx+R*math.cos(-math.pi/2+2*math.pi*i/N); ey=cy+R*math.sin(-math.pi/2+2*math.pi*i/N)
        out.append(f'<line x1="{cx}" y1="{cy}" x2="{ex:.1f}" y2="{ey:.1f}" stroke="#30363d" stroke-width="0.5"/>')
        lx=cx+(R+16)*math.cos(-math.pi/2+2*math.pi*i/N); ly=cy+(R+16)*math.sin(-math.pi/2+2*math.pi*i/N)
        anchor='middle' if abs(lx-cx)<8 else ('start' if lx>cx else 'end')
        out.append(f'<text x="{lx:.1f}" y="{ly:.1f}" fill="#8b949e" font-size="9" text-anchor="{anchor}" dominant-baseline="middle">{lab}</text>')
    pts=[]
    for i,(lab,key) in enumerate(axes):
        v=(me.get(key,0) or 0)/100.0
        pts.append(f"{cx+R*v*math.cos(-math.pi/2+2*math.pi*i/N):.1f},{cy+R*v*math.sin(-math.pi/2+2*math.pi*i/N):.1f}")
    out.append(f'<polygon points="{" ".join(pts)}" fill="{color}33" stroke="{color}" stroke-width="2"/>')
    out.append('</svg>')
    return '\n'.join(out)

def context_html(me, ms_sorted, author, lang, color):
    top=ms_sorted[:8]
    if me not in top: top=top+[me]
    gmax=max((m.get('gravity',0)or 0) for m in top) or 1
    rows=[]
    for m in top:
        g=m.get('gravity',0)or 0; is_me=(m is me)
        w=max(1.5,g/gmax*100)
        nm=esc(m.get('member') or '')
        star=' ★' if is_me else ''
        cls=' style="background:rgba(88,166,255,.10)"' if is_me else ''
        bar_col=color if is_me else '#42a5f5'
        rows.append(f'<div class="ctx-row"{cls}><div class="ctx-name"{" style=\"color:#f0f6fc;font-weight:700\"" if is_me else ""}>{nm}{star}</div>'
                    f'<div class="ctx-track"><div class="ctx-fill" style="width:{w:.1f}%;background:{bar_col}"></div><span class="ctx-val">{fmt(g)}</span></div></div>')
    return '\n'.join(rows)

CSS = """*{margin:0;padding:0;box-sizing:border-box}
body{background:#0d1117;color:#c9d1d9;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif}
.container{max-width:1100px;margin:0 auto;padding:20px}
h2{color:#f0f6fc;font-size:1.3em;margin:36px 0 14px;border-bottom:1px solid #30363d;padding-bottom:8px}
.badge{display:inline-block;padding:3px 10px;border-radius:12px;font-size:.75em;font-weight:600;color:#fff}
.hero{border-radius:16px;padding:40px;margin-bottom:24px;position:relative}
.hero-name{font-size:2.6em;font-weight:800;color:#f0f6fc;margin-bottom:6px}
.hero-tagline{font-size:1.1em;font-weight:600;margin-bottom:14px}
.hero-desc{color:#c9d1d9;font-size:.95em;line-height:1.75;max-width:760px}
.hero-badges{display:flex;gap:8px;flex-wrap:wrap;margin-top:18px}
.note{background:#161b22;border:1px solid #30363d;border-left:3px solid #58a6ff;border-radius:4px;padding:12px 16px;margin:0 0 24px;color:#8b949e;font-size:.88em;line-height:1.6}
.big-stats{display:flex;gap:16px;flex-wrap:wrap;margin:20px 0}
.big-stat{background:#161b22;border:1px solid #30363d;border-radius:8px;padding:18px 24px;text-align:center;flex:1;min-width:120px}
.big-num{font-size:2em;font-weight:800}.big-label{color:#8b949e;font-size:.78em;margin-top:4px}
.timeline-chart{background:#161b22;border:1px solid #30363d;border-radius:12px;padding:24px;margin:14px 0;overflow-x:auto}
.tl-callout{background:rgba(88,166,255,.08);border:1px solid #30363d;border-radius:8px;padding:12px 16px;margin-bottom:16px;font-size:.92em;line-height:1.6;color:#c9d1d9}
.timeline-table{width:100%;border-collapse:collapse;margin-top:16px;font-size:.8em}
.timeline-table th{padding:7px 8px;border-bottom:2px solid #30363d;color:#8b949e;font-weight:600;text-align:center;white-space:nowrap}
.timeline-table td{padding:7px 8px;border-bottom:1px solid #21262d;text-align:center}
.cols{display:grid;grid-template-columns:240px 1fr;gap:24px;align-items:center;background:#161b22;border:1px solid #30363d;border-radius:12px;padding:24px;margin:14px 0}
@media(max-width:720px){.cols{grid-template-columns:1fr}}
.ctx-row{display:flex;align-items:center;gap:12px;padding:5px 8px;border-radius:6px}
.ctx-name{width:200px;flex-shrink:0;font-size:.9em;color:#c9d1d9;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.ctx-track{position:relative;flex:1;height:18px;background:#21262d;border-radius:9px;overflow:hidden}
.ctx-fill{height:100%;border-radius:9px}
.ctx-val{position:absolute;right:8px;top:50%;transform:translateY(-50%);font-size:.75em;font-weight:700;color:#f0f6fc}
.insight{background:rgba(88,166,255,.06);border-left:3px solid #58a6ff;border-radius:4px;padding:16px 20px;margin:20px 0;font-size:.95em;line-height:1.75}
.insight strong{color:#f0f6fc}
.method{color:#8b949e;font-size:.9em;line-height:1.8}.method p{margin-top:8px}.method strong{color:#c9d1d9}
.footer{margin-top:48px;padding-top:16px;border-top:1px solid #30363d;color:#484f58;font-size:.8em;text-align:center}
.footer a{color:#58a6ff;text-decoration:none}
.lang-switch{position:absolute;top:18px;right:20px;font-size:.82em;z-index:50}
.lang-switch a{color:#8b949e;text-decoration:none;padding:2px 4px}.lang-switch a.active{color:#58a6ff;font-weight:700}
.lang-switch a:hover{color:#c9d1d9}.lang-switch .sep{color:#30363d;margin:0 2px}"""

def render(key, lang):
    p=PROFILES[key]; L=S[lang]; repo=p['repo']; author=p['author']; color=p['color']
    me, ms_sorted, rank, total = get_static(repo, author)
    if me is None: raise SystemExit(f"{author} not found in {repo}")
    tl=get_timeline(repo, author)
    g=me.get('gravity',0)or 0
    slug=key
    en_href=f"{slug}.html"; ja_href=f"{slug}-ja.html"
    def link(href,lab,active): return f'<a href="{href}"{" class=\"active\"" if active else ""}>{lab}</a>'
    toggle=f'<div class="lang-switch">{link(en_href,"EN",lang=="en")}<span class="sep">·</span>{link(ja_href,"日本語",lang=="ja")}</div>'
    # stats
    span=f"{tl['labels'][0]}–{tl['labels'][-1]}" if tl else "—"
    stats=[(L['s_gravity'],fmt(g),color),(L['s_design'],fmt(me.get('design',0)or 0),'#ffd700'),
           (L['s_indisp'],fmt(me.get('indispensability',0)or 0),'#79c0ff'),
           (L['s_commits'],f"{me.get('commits',0):,}",'#56d364'),(L['s_span'],span,'#8b949e')]
    bigstats=''.join(f'<div class="big-stat"><div class="big-num" style="color:{c}">{v}</div><div class="big-label">{esc(l)}</div></div>' for l,v,c in stats)
    # timeline
    if tl:
        svg,tbl=timeline_svg(tl,lang)
        pk=max(range(len(tl['gravity'])),key=lambda i:tl['gravity'][i])
        pkg=tl['gravity'][pk]; pky=tl['labels'][pk]
        if lang=='en':
            arc=f"{L['tl_callout_pre']} Gravity peaks at <strong>{fmt(pkg)} in {pky}</strong>, then settles toward today’s <strong>{fmt(g)}</strong> — a present-tense reading, not a fixed label."
        else:
            arc=f"{L['tl_callout_pre']} Gravity は <strong>{pky} に {fmt(pkg)}</strong> でピークを打ち、いまの <strong>{fmt(g)}</strong> へと落ち着く — 固定ラベルではなく、現在形の観測だ。"
        timeline_block=(f'<h2>{esc(L["tl_title"])}</h2><p style="color:#8b949e;font-size:.9em;margin-bottom:12px">{L["tl_desc"]}</p>'
                        f'<div class="timeline-chart"><div class="tl-callout">{arc}</div>{svg}{tbl}</div>')
    else:
        timeline_block=f'<h2>{esc(L["tl_title"])}</h2><p style="color:#8b949e">—</p>'
    radar=radar_svg(me,color)
    ctx=context_html(me,ms_sorted,author,lang,color)
    method=''.join(f'<p>{m}</p>' for m in L['method'])
    rank_txt=(f"#{rank} / {total}" )  # internal index only; shown small
    return f"""<!DOCTYPE html>
<html lang="{L['html_lang']}">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{esc(author)} — {esc(L['title_suffix'])} | OSS Gravity Map by EIS</title>
<style>
{CSS}
.hero{{background:linear-gradient(135deg,{color}22 0%,#0d1117 70%);border:2px solid {color}}}
.hero-tagline{{color:{color}}}
</style>
</head>
<body>
{toggle}
<div class="container">
<div class="hero">
  <div class="hero-name">{esc(author)}</div>
  <div class="hero-tagline">{esc(p['tagline'][lang])}</div>
  <div class="hero-desc">{p['hero'][lang]}</div>
  <div class="hero-badges">
    <span class="badge" style="background:{color};color:#000">{esc(p['repo_label'])}</span>
    <span class="badge" style="background:#30363d">Gravity {fmt(g)}</span>
    <span class="badge" style="background:#30363d">{esc(L['s_design'])} {fmt(me.get('design',0)or 0)}</span>
    {f'<span class="badge" style="background:#30363d">{esc(str(me.get("role")))}</span>' if me.get('role') and me.get('role')!='—' else ''}
    {f'<span class="badge" style="background:#30363d">{esc(str(me.get("state")))}</span>' if me.get('state') and me.get('state')!='—' else ''}
  </div>
</div>
<div class="note"><strong style="color:#c9d1d9">{esc(L['scope_pre'])}</strong> {L['scope'](p['repo_label'])}</div>
<div class="big-stats">{bigstats}</div>
{timeline_block}
<div class="cols">
  <div style="display:flex;justify-content:center">{radar}</div>
  <div>
    <h3 style="color:#f0f6fc;margin-bottom:6px;font-size:1.05em">{esc(L['ctx_title'])}</h3>
    <p style="color:#8b949e;font-size:.85em;margin-bottom:12px">{L['ctx_desc']}</p>
    {ctx}
  </div>
</div>
<h2>{esc(L['method_title'])}</h2>
<div class="method">{method}</div>
<div class="insight">{p['insight'][lang]}</div>
<div class="footer">
  Generated by <a href="https://github.com/machuz/eis">EIS (Engineering Impact Signal)</a> — OSS Gravity Map Project<br>
  {esc(p['repo_label'])} &middot; {total:,} engineers observed
</div>
</div>
</body>
</html>"""

def main():
    outdir=os.path.join(ROOT,'analysis')
    if len(sys.argv)>=2 and sys.argv[1]=='--all':
        done=[]
        for k in ORDER:
            for lang in ('en','ja'):
                fn=f"{k}.html" if lang=='en' else f"{k}-ja.html"
                try:
                    s=render(k,lang)
                except SystemExit as e:
                    print(f"  SKIP {fn}: {e}"); continue
                open(os.path.join(outdir,fn),'w',encoding='utf-8').write(s)
                done.append(fn)
        print("written:", ', '.join(done))
        return
    key,lang,out=sys.argv[1],sys.argv[2],sys.argv[3]
    open(out,'w',encoding='utf-8').write(render(key,lang))
    print("wrote",out)

if __name__=='__main__':
    main()
