# git考古学 #7 —— コードの宇宙を観測する

*優秀なエンジニアとは、単にコードを書く人ではない。コードベースの重力を曲げる人なのかもしれない。*

### 前章までのあらすじ

[第6章](https://ma2k8.hateblo.jp/entry/2026/03/14/180329)ではチームタイムラインを読み解き、個人のRole/Styleの変化がチームのCharacter/Structureの変化として表面化する法則を見た。

しかしあれはまだ「分析結果を読む」話だった。

今回は少し違う話をする。EISを作り続けた先に辿り着いた感覚——**コードベースが持つ宇宙的な構造**と、それを観測可能にすることの意味について。

---

## HTML Dashboard：データを眺める体験

まず実用的な話から。`eis timeline --format html` が出力するインタラクティブダッシュボードが、かなり使えるようになった。

![Timeline HTML Dashboard](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/timeline-html-output.png?v=0.11.0)

```bash
eis timeline --format html --output timeline.html --recursive ~/workspace
```

Chart.jsベースの折れ線グラフで、個人・チームのスコア推移、Health指標、メンバーシップ構成、Classification変遷が一覧できる。ツールチップにはRole/Style/State/Confidenceが表示され、Transitionマーカーが変化のタイミングを示す。

何が良いかというと、**AIと一緒にこの画面を眺められる**ことだ。

HTMLをブラウザで開いて、AIに「このチームの2024-H2に何が起きた？」と聞く。数値の変化、Role遷移、Health指標の動き——データを見ながら仮説を立て、AIが解釈を返す。これは`eis analyze`のターミナル出力では難しかった体験だ。

---

## コードの宇宙

EISを作っていて、最後に辿り着いた感覚はとてもシンプルだった。

**コードベースは宇宙のような構造を持っている。**

そこには重力がある。

あるコードの周りに他のコードが集まり始める。抽象化がそこを通る。設計がそこを中心に組み上がる。

すると**構造の中心**が生まれる。

EISの用語で言えば、これは **Architect** だ。Design軸が高く、Survival軸が高い。つまりそのエンジニアが書いたコードを中心に、他のコードが組み上がり、そして**その構造が生き残っている**。

これがコードベースの重力だ。

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

第5章で見たEngineer Fの退場が、まさにこれだった。

---

## 複数の重力が存在する宇宙

本当に強いチームでは、別の現象が起きる。

**新しい重力が生まれる。**

既存のArchitectが構造を成立させている。その周辺で**創発型Architect**が新しい重力を作り始める。

設計は衝突する。抽象化の粒度が議論される。実装方針がぶつかる。

一見すると**揉めている**ように見えるかもしれない。

しかし構造的に見ると、これは**コード宇宙の進化**である。

EISのチームタイムラインで見ると、この進化は具体的な指標で追える。

```
Classification:
  Period         2024-H1            2024-H2             2025-H1
  Character      Elite              Guardian            Elite
  Structure      Architectural Team Maintenance Team    Architectural Engine
  Risk           Bus Factor         Design Vacuum       Healthy
```

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

### シリーズ

- [第1章：Git履歴だけでエンジニアの「本当の影響力」を測る](https://ma2k8.hateblo.jp/entry/2026/03/11/153212)
- [第2章：チーム分析——個人スコアからチームの健康状態が見える](https://ma2k8.hateblo.jp/entry/2026/03/13/023837)
- [第3章：アーキタイプ——数値の裏に人間が見える](https://ma2k8.hateblo.jp/entry/2026/03/13/024048)
- [第4章：正規化の罠——「あるリポジトリのトップ」は何も意味しない](https://ma2k8.hateblo.jp/entry/2026/03/13/024230)
- [第5章：タイムライン——スコアは嘘をつかない、そして「遠慮」も捉える](https://ma2k8.hateblo.jp/entry/2026/03/14/180329)
- [第6章：チームは進化する——タイムラインが暴く組織の法則](https://ma2k8.hateblo.jp/entry/2026/03/14/180329)
- **第7章：コードの宇宙を観測する**（本記事）

**GitHub**: [engineering-impact-score](https://github.com/machuz/engineering-impact-score)
