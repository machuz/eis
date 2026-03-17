---
title: "コードの宇宙を観測する"
---


![Cover](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/hatena/cover-ch7.png)

*優秀なエンジニアとは、単にコードを書く人ではない。コードベースの重力を曲げる人なのかもしれない。*

### 前章までのあらすじ

第6章ではチームタイムラインを読み解き、個人のRole/Styleの変化がチームのCharacter/Structureの変化として表面化する法則を見た。

しかしあれはまだ「分析結果を読む」話だった。

今回は少し違う話をする。EISを作り続けた先に辿り着いた感覚——**コードベースが持つ宇宙的な構造**と、それを観測可能にすることの意味について。

---

## HTML Dashboard：データを眺める体験

まず実用的な話から。`eis timeline --format html` が出力するインタラクティブダッシュボードが使えるようになった。

![Timeline HTML Dashboard](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/timeline-html-output.png?v=0.11.0)

```bash
❯ eis timeline --format html --output timeline.html --recursive ~/workspace
```

Chart.jsベースの折れ線グラフで、個人・チームのスコア推移、Health指標、メンバーシップ構成、Classification変遷が一覧できる。ツールチップにはRole/Style/State/Confidenceが表示され、Transitionマーカーが変化のタイミングを示す。

ターミナルでサッと確認したいなら `--format ascii` もある。

![Timeline ASCII Output](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/timeline-ascii-output.png?v=0.11.0)

何が良いかというと、**AIと一緒にこの画面を眺められる**ことだ。

HTMLをブラウザで開き、`eis timeline --format json` の出力をAIに渡して「このチームの2024-H2に何が起きた？」と聞く。AIはグラフの数値変化、Role遷移、Health指標の動きを読み取り、仮説と解釈を返してくれる。これは`eis analyze`のターミナル出力では難しかった体験だ。

特にチームのHealth Metricsが面白い。Complementarity（補完性）、Growth Potential（成長余地）、Sustainability（持続性）、Debt Balance（負債バランス）——これらが期間ごとにどう変化しているかが一目でわかる。

![Team Health Metrics](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/team-health-metrics-output.png?v=0.11.0)

---

## コードの宇宙

EISを作っていて、最後に辿り着いた感覚はとてもシンプルだった。

**コードベースは宇宙のような構造を持っている。**

そこには重力がある。

あるコードの周りに他のコードが集まり始める。抽象化がそこを通る。設計がそこを中心に組み上がる。

すると**構造の中心**が生まれる。

EISの用語で言えば、これは **Architect** だ。Design軸が高く、Survival軸が高い。つまりそのエンジニアが書いたコードを中心に、他のコードが組み上がり、そして**その構造が生き残っている**。

これがコードベースの重力だ。

コードベースには目に見えない**構造の重力場**がある。新しいコードは既存の重力に引かれて配置される。新しいエンジニアは重力場に沿って設計を理解する。重力場が強いほど、構造は安定する。弱いほど、カオスに向かう。

---

## 重力が一つしかない宇宙

多くのチームでは、この重力は一つしかない。

一人のArchitectがコードベースの中心構造を作る。APIの設計方針を決める。抽象化の粒度を定める。ディレクトリ構造を固める。

これは強い構造だ。一貫性がある。迷いがない。

しかし同時に**とても脆い構造**でもある。

EISの指標で見ると、こういうチームには特徴がある。

- **Risk: Bus Factor** — 一人の離脱がチーム構造を崩壊させる
- **Structure: Architectural Team** ではなく **Maintenance Team** — 設計者が一人なので、残りのメンバーは既存構造を維持するだけ
- **Anchor Density: 低い** — 品質の安定化を担うメンバーが少ない

もしその重力が消えたら——そのArchitectが異動した、退職した、別プロジェクトに移った——コードベースは一気に拡散する。設計の一貫性は崩れ、構造は弱くなる。

第5章で見たY.Y.の退場は、この危機が起こり得た瞬間だった。しかし実際には崩壊しなかった。machuzが同じタイミングでArchitect Builderに到達し、**重力の世代交代**が起きたからだ。これは幸運でもあり、構造的な必然でもあった。だが、もしmachuzがいなかったら？ そのときチームは「重力を失った宇宙」になっていた可能性がある。

---

## 複数の重力が存在する宇宙

本当に強いチームでは、別の現象が起きる。

**新しい重力が生まれる。**

既存のArchitectが構造を成立させている。その周辺で**創発型Architect**が新しい重力を作り始める。

そして時間が経つと、重力は洗練される。抽象化が安定し、コードが生き残り、依存が集まる。こうして生まれるのが**「練度の高い良い重力」**だ。

設計は衝突する。抽象化の粒度が議論される。実装方針がぶつかる。

一見すると**揉めている**ように見えるかもしれない。

しかし構造的に見ると、これは**コード宇宙の進化**である。

EISのチームタイムラインで見ると、この進化は具体的な指標で追える。

![Team Classification](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch7-team-classification.svg)

**Architectural Team → Maintenance Team → Architectural Engine**

最初は一人のArchitectが構造を支えていた（Architectural Team）。その人の影響力が薄れると構造は維持モードに入る（Maintenance Team）。しかし複数のメンバーがDesign軸を伸ばし始めると、チーム全体が設計力を持つ構造に進化する（Architectural Engine）。

これが「複数の重力が存在する宇宙」だ。

---

## 四つの力

重力が複数存在するチームでは、四つの力が同時に働く。

| 力 | 担い手 | EISでの観測 |
|---|---|---|
| **守る力** | Architect Builder | 高Design + 高Survival。既存構造を維持し、一貫性を守る |
| **壊す力** | 創発型Architect | 高Design + 高Production。新しい抽象化を導入し、既存構造を更新する |
| **安定させる力** | Anchor | 高Quality + 高Survival。バグを直し、テストを書き、品質を底上げする |
| **作る力** | Producer | 高Production。機能を前進させ、ユーザー価値を生み出す |

この四つが揃ったとき、チームのClassificationは **Elite / Architectural Engine / Builder / Mature / Healthy** になる。

逆に言えば、どれか一つでも欠けると、それがRiskとして表面化する。

- 守る力が欠ける → **Design Vacuum**（設計の空白）
- 安定させる力が欠ける → **Quality Drift**（品質の漂流）
- 壊す力が欠ける → **Stagnation**（停滞）
- 作る力が欠ける → **Declining**（衰退）

---

## 声の大きさではなく、構造への影響

ここまで書いてきて、エンジニア評価の文脈でも面白い意味を持つことに気づく。

ソフトウェアの世界には、**腕は立つが声が小さいエンジニア**がいる。

良い設計を理解している。良いコードを書く。しかし会議ではあまり話さない。ドキュメントも最低限。プレゼンは苦手。

逆に、**声は大きいがコードには重力が残らないエンジニア**もいる。

方針は語る。設計レビューでは意見を言う。しかし実際のコードベースを見ると、そのエンジニアのコードは他のコードの中心にはなっていない。Survival軸が低い。Design軸が低い。

どちらが**コードベースを動かしているのか**。

それは議論ではなくコードに残る。Git履歴には、誰がコードを書いたかだけではなく、**誰がコードベースの重力を作ったか**も残っている。

EISがやろうとしているのは、**声の大きさではなく構造への影響**でエンジニアを見ることだ。

---

## チームの実力を証明する——採用という文脈

この「観測可能性」には、もう一つ実用的な使い道がある。**採用**だ。

エンジニア採用の場面で「うちのチームは技術力が高い」と言うのは簡単だ。しかしそれを**データで示せる**チームはほとんどない。

EISのチームタイムラインがあれば、こういうことができる。

- **Classification: Elite / Architectural Engine / Mature** ——「うちのチームは設計力が分散していて、特定個人に依存しない構造です」
- **Health: Complementarity 0.85** ——「メンバーのスキルが相互補完的で、偏りが少ないです」
- **Risk: Healthy** ——「Bus Factor、Design Vacuum、Quality Driftのいずれのリスクもありません」
- **Phase: Mature → Rebuilding ではなく Mature を維持** ——「安定しつつも停滞していません」

候補者に対して「技術的に面白いチームです」と口で言うのではなく、**グラフを見せる**。スコアの推移を見せる。チームがどう進化してきたかをデータで語る。

これは逆方向にも効く。候補者がチームを選ぶ判断材料になる。「このチームは本当に成長しているのか？」「設計文化があるのか？」——EISのダッシュボードはそれに答えられる。

採用は「マッチング」だ。チームの実力を正直に、定量的に見せられることは、双方にとって価値がある。

---

## 観測可能にするということ

物理学の歴史は「観測」の歴史でもある。

惑星が太陽の周りを回っていることは、望遠鏡ができる前から事実だった。しかし**観測可能になった**ことで、初めてその構造を理解し、予測し、活用できるようになった。

コードベースの構造も同じだ。

誰がArchitectなのか。どこに重力があるのか。チームが進化しているのか衰退しているのか。

これらはGit履歴の中に**既に存在している**。ただ観測できなかっただけだ。

EISは、コードの宇宙を少しだけ**観測可能にする**試みである。

---

## Great engineers don't just write code. They bend the gravity of codebases.

優秀なエンジニアとは、単にコードを書く人ではない。

**コードベースの重力を曲げる人**なのかもしれない。

---
