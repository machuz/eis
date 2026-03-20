# Engineering Impact Signal: Git履歴からソフトウェアエンジニアリング貢献を定量化する

**Version 0.11.0** — 2026年3月

**著者:** machuz ([@machuz](https://github.com/machuz))

---

## 概要

Engineering Impact Signal (EIS) は、Git履歴データのみを用いて、個人およびチームレベルのソフトウェアエンジニアリング貢献を定量化するオープンソースフレームワークである。コミット数、コード行数、PR処理数といった代理指標に依存する既存手法とは異なり、EISはエンジニアが *何を* 貢献しているか、*どのように* 貢献しているか、*ライフサイクルのどこにいるか* を捉える多軸観測モデルを構築する。コミットベースの生産性メトリクスと `git blame` ベースの生存分析、そして独自の変更圧力分解（Change-Pressure Decomposition）を組み合わせ、活発に開発されているモジュールで耐え抜くコードと、単に放置されたモジュールに残存するコードを区別する。

チームレベルでは、個人シグナルを5軸分類システムに集約し、チームの構造、文化、ライフサイクルフェーズ、リスクプロファイル、総合的な性格を特徴づける。タイムライン分析モードは、設定可能な期間にわたってこれらのメトリクスを追跡し、エンジニアリング組織の経年観測を可能にする。

本稿では、EISの数学的基盤、分類アルゴリズム、設計の根拠を、限界と想定ユースケースとともに提示する。

**キーワード:** ソフトウェアエンジニアリングメトリクス、git分析、コード生存率、チームヘルス、エンジニア評価、開発者生産性

---

## 1. はじめに

### 1.1 問題

ソフトウェアエンジニアリングの貢献を測定することは、長年の課題である。従来のメトリクスは予測可能な形で失敗する：

- **コミット数** はコミットの粒度に報い、影響力には報いない
- **コード行数** はリファクタリングを罰し、冗長さに報いる
- **PRスループット** はプロセス遵守を測り、エンジニアリング品質は測らない
- **ストーリーポイント** は主観的でチーム固有である

これらのメトリクスに共通する欠陥は、*活動* を測定しているのであって *影響* を測定していないことだ。次の四半期に削除される10,000行を書いたエンジニアと、システムアーキテクチャの基盤となる500行を書いたエンジニアは、活動指標では区別できない。

### 1.2 核心的洞察

Gitリポジトリにはコミット数以上の情報が含まれている。具体的には、`git blame` はどのコードが *生存した* かを明らかにする——どの著者が書いたどの行が、現在のコードベースにまだ存在しているかを。コミットメタデータ（タイムスタンプ、ファイルパス、メッセージパターン）と組み合わせることで、貢献の性質と耐久性に関するリッチなシグナルが生まれる。

EISはこの洞察を3つのメカニズムで活用する：

1. **生存分析**: `git blame` を使って、エンジニアのコードがどれだけ持続しているかを指数関数的時間減衰重み付きで測定
2. **変更圧力分解**: 生存を *堅牢* （頻繁に変更されるモジュールで生存）と *休眠* （ほとんど触られないモジュールで生存）に分割
3. **アーキテクチャパターン検出**: 構造的に重要なファイル（インターフェース、ルーター、依存性注入、ドメインサービス）への貢献を識別

### 1.3 設計原則

#### 基礎原則

1. **局所宇宙（Local Universes）**: すべてのコードベースはそれぞれ固有の宇宙である。インパクトはそのローカルコンテキストの中で理解されなければならない——正規化、アーキタイプ、チーム分類はすべて観測対象のリポジトリ内で相対的に計算される。
2. **観測可能な重力（Observable Gravity）**: 影響力はコードにおける重力として現れる：生存、再利用、構造的な引力。EISはこの重力を測定するのであり、活動を測定するのではない。
3. **進化的軌跡（Evolutionary Trajectories）**: ソフトウェアは時間とともに進化する。エンジニアリングの価値はその進化の軌跡に現れる——単一のスナップショットにではなく。

**観測者の原理（The Principle of Observers）。** EISは価値を定義しない。コードベースにすでに存在する構造を明らかにする観測装置として機能する。望遠鏡のように、宇宙を変えることはない。ただその重力を可視化するだけである。

#### 技術原則

1. **Git-only**: プロジェクト管理ツール、CIシステム、コードレビュープラットフォームとの統合不要。分析に必要なのはGitリポジトリのみ。
2. **多軸**: 単一のシグナルではエンジニアリング貢献を捉えられない。EISは7つの個人軸と5つのチームレベル分類軸を生成する。
3. **相対＋絶対**: 一部のメトリクス（Production）はチーム間比較のために絶対参照を使用し、他（Design、Survival）はチーム内正規化を使用する。
4. **観測的、規範的でない**: EISはコードベースで何が起きたかを記述する。何が起きる *べき* かは定義しない。

---

## 2. 関連研究

### 2.1 DORAメトリクス

DevOps Research and Assessment (DORA) フレームワークは4つの主要メトリクスを測定する：デプロイ頻度、変更のリードタイム、変更失敗率、サービス復旧時間。DORAはチーム/組織レベルで動作し、個人の貢献パターンではなくデリバリーパイプラインのパフォーマンスに焦点を当てる。EISはDORAが意図的に避けている個人レベルの解像度を提供することで補完する。

### 2.2 CodeScene

CodeSceneは行動的コード分析を行い、ホットスポット、コードヘルス、組織パターンを識別する。プロプライエタリな観測モデルを使用し、エンジニアレベルの分類ではなくコードレベルのヘルスに焦点を当てる。EISは明示的な多軸分類システムと、オープンソースで再現可能な方法論という点で異なる。

### 2.3 git-fameとgit-quick-stats

これらのツールは基本的な帰属統計（著者ごとのコード行数、コミット数）を提供する。EISは同じ生データに基づくが、生存分析、アーキテクチャパターン検出、これらのツールにはない多軸分類を追加する。

### 2.4 学術研究

Nagappanら (2008) は、バージョン管理から導出された組織メトリクスがコードメトリクス単体よりもソフトウェア欠陥を良く予測することを実証した。Birdら (2011) は、コード所有権パターンがソフトウェア品質と相関することを示した。EISはこれらの知見を実用的な自動化フレームワークに形式化する。

---

## 3. 個人プロファイリングモデル

### 3.1 概要

EISはリポジトリ内の各コントリビュータに対して7つの軸を計算する：

| 軸 | シグナル | スケール | ソース |
|----|----------|----------|--------|
| Production | コード変更量 | 絶対 | コミット |
| Quality | 修正比率の逆数 | 絶対 | コミットメッセージ |
| Survival | コードの耐久性（時間減衰付き） | 相対 | `git blame` |
| Design | アーキテクチャファイルへの変更 | 相対 | コミット＋パターン |
| Breadth | 触れたモジュールの多様性 | 相対 | コミット |
| Debt Cleanup | 他者のコード修正 vs 負債生成の比率 | 絶対 | `git blame` ＋ コミット |
| Indispensability | バスファクター——モジュールの独占的所有 | 相対 | `git blame` |

各軸は [0, 100] のシグナルを生成する。

### 3.2 Production（生産性）

Productionはコード変更の生の量を測定する。

**計算式:**

$$\text{Production}_a = \sum_{c \in \text{commits}(a)} \sum_{f \in \text{files}(c)} (\text{insertions}_f + \text{deletions}_f)$$

**正規化（絶対）:**

$$\text{Score}_a = \min\left(\frac{\text{Production}_a / \text{activeDays}_a}{\text{ProductionDailyRef}} \times 100,\; 100\right)$$

`ProductionDailyRef`（デフォルト: 1000行/日）がチーム間比較のための固定ベースラインを提供する。除外パターンに一致するファイル（ロックファイル、生成コード、swaggerドキュメント）は除外される。

**根拠:** Productionは意図的に絶対メトリクスとして維持される。2人のチームと20人のチームは、一人当たりの出力で比較可能であるべきだ。

### 3.3 Quality（品質）

Qualityはエンジニアの修正比率の逆数を測定する。

**計算式:**

$$\text{FixRatio}_a = \frac{|\{c \in \text{commits}(a) : \text{isFix}(c)\}|}{|\text{commits}(a) \setminus \text{merges}(a)|}$$

$$\text{Quality}_a = 100 - \text{FixRatio}_a \times 100$$

**修正検出:** コミットメッセージが以下にマッチする場合、修正として分類される：

```
(?i)^[^\w]*(?:\[?\s*(?:fix|revert|hotfix)\s*\]?[:/\s])
```

または日本語の「修正」を含む場合。マージコミットは修正カウントに寄与するが、総カウントには寄与しない。これにより、マージの多いワークフローで品質シグナルが膨張するのを防ぐ。

**根拠:** コミットの大部分が修正であるエンジニアは、ミスの修正に時間を費やしている（自分自身のまたは他者の）。Qualityは絶対スケールであり、チーム構成に依存しない。

### 3.4 Survival（生存率）

SurvivalはEISの中心的イノベーションである。エンジニアのコードが現在のコードベースにどれだけ持続しているかを、時間重み付きで測定する。

**Raw Survival:**

$$\text{RawSurvival}_a = |\{l \in \text{blame} : \text{author}(l) = a\}|$$

**時間減衰Survival:**

$$\text{Survival}_a = \sum_{l \in \text{blame}(a)} e^{-d_l / \tau}$$

$d_l$ は行 $l$ の経過日数、$\tau$ は減衰定数（デフォルト: 180日）。この重み付けにより、最近書かれた生存コードが、慣性だけで残存する古いコードよりも高く評価される。

#### 3.4.1 変更圧力分解（Change-Pressure Decomposition）

すべての生存コードが等価ではない。四半期に50コミットを受けるモジュール内のコードは *協働によってテストされている* ——他のエンジニアがその周辺で作業し、隣接コードを修正し、それでもなお生存している。四半期に1コミットしか受けないモジュール内のコードは、単に誰も見ていないから生存しているかもしれない。

EISはこれらを区別するために **変更圧力** を導入する：

$$\text{Pressure}_m = \frac{|\text{モジュール } m \text{ に触れるコミット数}|}{|\text{モジュール } m \text{ のblame行数}|}$$

中央値を閾値として：

- **Robust Survival**: $\text{Pressure}_m \geq \text{median}$ のモジュール内の行
- **Dormant Survival**: $\text{Pressure}_m < \text{median}$ のモジュール内の行

両者とも同じ指数減衰式を使用するが、独立に計算される。

**根拠:** この分解は「忘れられたコード」問題を防ぐ——エンジニアが誰も修正しないモジュールにコードを書いたというだけで生存シグナルが高くなることを。Robust Survivalは特に *活発な開発の中で耐え抜いた* コードを測定する。

### 3.5 Design（設計）

Designはアーキテクチャ的に重要なファイルへの貢献を測定する。

**計算式:**

$$\text{Design}_a = \sum_{c \in \text{commits}(a)} \sum_{f \in \text{archFiles}(c)} (\text{insertions}_f + \text{deletions}_f)$$

**アーキテクチャ検出:** 設定可能なglobパターンに一致するファイルがアーキテクチャファイルとして分類される：

```
*/repository/*interface*    # リポジトリインターフェース
*/domainservice/            # ドメインサービス
*/router.go                 # ルーティング定義
*/middleware/               # ミドルウェア層
di/*.go                     # 依存性注入
*/core/                     # コアモジュール
*/stores/                   # 状態管理（フロントエンド）
*/hooks/                    # Reactフック（フロントエンド）
*/types/                    # 型定義
```

**正規化:** 相対（最大値ベース）。チーム内で最もDesignに貢献したメンバーが100を獲得する。

**根拠:** すべてのコード変更が等価ではない。インターフェース、依存性注入、ルーティング設定への変更は、単一モジュール内の変更と比べて構造的影響が大きい。

### 3.6 Breadth（広がり）

Breadthはエンジニアが触れるモジュールの多様性を測定する。

$$\text{Breadth}_a = |\{\text{著者 } a \text{ が触れたユニークモジュール数}\}|$$

チーム最大値に対して正規化され、`BreadthMax`（デフォルト: 5）で上限が設定される。

**根拠:** 多くのモジュールにまたがって貢献するエンジニアは、深い専門家とは異なる組織的影響を持つ。どちらが本質的に優れているわけではないが、チーム構成分析においてこの区別は意味がある。

### 3.7 Debt Cleanup（負債清算）

Debt Cleanupはエンジニアが他者の技術的負債を清算しているか、他者が清算すべき負債を生成しているかを測定する。

**アルゴリズム:**

1. 修正コミットを識別（Qualityの修正検出正規表現を使用）
2. 最大500件の修正コミットをサンプリング
3. 各修正コミットについて、親コミットで `git blame` を実行し、変更された行の元の著者を特定
4. カウント:
   - $\text{cleaned}_a$ = 著者 $a$ が他者のコードを修正した回数
   - $\text{generated}_a$ = 他者が著者 $a$ のコードを修正した回数

**シグナル:**

$$\text{DebtCleanup}_a = 50 + 50 \times \frac{\text{cleaned}_a - \text{generated}_a}{\text{cleaned}_a + \text{generated}_a}$$

- シグナル = 0: 純粋な負債生成者
- シグナル = 50: 中立（均衡または不十分なデータ）
- シグナル = 100: 純粋なクリーナー

合計インタラクションが `DebtThreshold`（デフォルト: 10）未満の著者は中立シグナル50を受ける。

**根拠:** このメトリクスは他のツールでは不可視な次元を捉える——エンジニアのコードが他者による修正を必要とする傾向があるのか、あるいは他者のコードを修正する傾向があるのか。数式は対称的で、純粋な生成者から純粋なクリーナーまでの範囲を持つ。

### 3.8 Indispensability（不可欠性）

Indispensabilityは個人レベルのバスファクターリスクを測定する。

**アルゴリズム:**

1. `git blame` の行をモジュール（パスの最初の2コンポーネント、例: `app/domain`）でグルーピング
2. 各モジュールで行数トップの著者を特定
3. 所有権シェアを計算: $\text{share} = \text{topCount} / \text{totalLines}$
4. シグナル:
   - クリティカルな所有権（$\text{share} \geq 0.80$）: モジュールあたり +1.0
   - 高い所有権（$\text{share} \geq 0.60$）: モジュールあたり +0.5

$$\text{Indispensability}_a = \text{criticalCount} \times 1.0 + \text{highCount} \times 0.5$$

チーム最大値に対して正規化。

**根拠:** 高いIndispensabilityは *リスク指標* であり、達成指標ではない。複数モジュールの80%以上を所有するエンジニアは単一障害点を表す。

### 3.9 正規化戦略

EISは2つの正規化戦略を使用する：

**最大値ベース相対正規化**（Survival、Design、Breadth、Indispensability向け）：

$$\text{Score}_a = \min\left(\frac{\text{raw}_a}{\max_b(\text{raw}_b)} \times 100,\; 100\right)$$

最高貢献者は常にシグナル100を獲得する。チームコンテキスト外で絶対値が無意味なメトリクスに適切。

**絶対正規化**（Production、Quality、Debt Cleanup向け）：

チーム間比較を可能にする固定参照点。Productionは日次参照率を使用し、QualityとDebt Cleanupは本質的に有界。

### 3.10 Impact

Impactは加重和：

$$\text{Total} = \sum_{i} w_i \times \text{Score}_i$$

デフォルトの重み：

| 軸 | 重み |
|----|------|
| Production | 0.15 |
| Quality | 0.10 |
| Survival | 0.25 |
| Design | 0.20 |
| Breadth | 0.10 |
| Debt Cleanup | 0.15 |
| Indispensability | 0.05 |

変更圧力データが利用可能な場合、Survivalは分割される：

$$\text{Survival寄与} = w_{\text{survival}} \times (0.80 \times \text{RobustSurvival} + 0.20 \times \text{DormantSurvival})$$

さらに、Designは証明係数によって減衰される：

$$\text{designDamping} = \max\left(\frac{\text{RobustSurvival}}{100} \times 0.8 + 0.2,\; \frac{\text{Production}}{100} \times 0.8 + 0.2\right)$$

$$\text{effectiveDesign} = \text{Design} \times \text{designDamping}$$

これにより、アーキテクチャファイルを所有しているが積極的に生産も行わず、コードが圧力下で生存もしていないエンジニアのDesignシグナルの膨張を防ぐ。

Robust Survivalがゼロのエンジニアには0.80倍のペナルティが適用される。これはコードが協働によるテストを一度も経ていないことを示す。

#### 重力シグナル（Gravity）

構造的影響を測定する別の複合シグナル：

$$\text{Gravity} = 0.40 \times \text{Indispensability} + 0.30 \times \text{Breadth} + 0.30 \times \text{Design}$$

---

## 4. 個人分類: 3軸トポロジー

生シグナルは、エンジニアの貢献の *性質* を記述する3つの直交軸に分類される。

### 4.1 ソフトマッチング関数

ハードな閾値の代わりに、EISはシグモイド的なソフトマッチング関数を使用する：

**highness(v)** — $v$ が「高い」（≥60）である信頼度：

$$\text{highness}(v) = \begin{cases} 1.0 & v \geq 80 \\ 0.5 + \frac{v-60}{40} & 60 \leq v < 80 \\ \frac{v-40}{40} \times 0.3 & 40 \leq v < 60 \\ 0 & v < 40 \end{cases}$$

**lowness(v)** — $v$ が「低い」（<30）である信頼度：

$$\text{lowness}(v) = \begin{cases} 1.0 & v < 10 \\ 0.5 + \frac{30-v}{40} & 10 \leq v < 30 \\ \frac{50-v}{40} \times 0.3 & 30 \leq v < 50 \\ 0 & v \geq 50 \end{cases}$$

**notLow(v)** — $v$ が「低くない」（≥50）である信頼度：

$$\text{notLow}(v) = \begin{cases} 1.0 & v \geq 50 \\ 0.5 + \frac{v-30}{40} & 30 \leq v < 50 \\ \frac{v-10}{40} \times 0.3 & 10 \leq v < 30 \\ 0 & v < 10 \end{cases}$$

これらの関数は [0, 1] の連続的な信頼度値を生成し、ハードカットオフの脆さを回避する。

### 4.2 Role軸 —— 「何を貢献しているか？」

| Role | 信頼度の計算式 | 解釈 |
|------|---------------|------|
| **Architect** | $\min(\text{highness(Design)},\; \text{highness(Survival)},\; \text{notLow(Breadth)})$ | 他者がその上に構築する構造を形成する |
| **Anchor** | $\min(\text{highness(Quality)},\; \text{notLow(Production)})$ | コードベース全体の品質を安定させる |
| **Cleaner** | $\min(\text{highness(Quality)},\; \text{highness(Survival)},\; \text{highness(DebtCleanup)})$ | 他者の技術的負債を解消する |
| **Producer** | $\text{notLow(Production)}$ | アウトプットを生成し、機能を前進させる |
| **Specialist** | $\min(\text{highness(Survival)},\; \text{lowness(Breadth)})$ | 狭い領域での深い専門性 |

変更圧力データが利用可能な場合、ArchitectはSurvival全体の代わりにRobust Survivalを使用する（ただしProductionが高い場合は、アクティブなビルダーにペナルティを与えないようSurvival全体を使用）。

**選択:** 最小閾値（0.10）以上で最も高い信頼度を持つRoleが選択される。0.15マージン内のタイの場合、先に定義されたルールが優先される。

### 4.3 Style軸 —— 「どのように貢献しているか？」

| Style | 信頼度の計算式 | 解釈 |
|-------|---------------|------|
| **Builder** | $\min(\text{high(Prod)},\; \text{high(Design)},\; \text{notLow(Debt)})$ | 設計し、大量に構築し、さらに後始末もする |
| **Resilient** | $\min(\text{high(Prod)},\; \text{low(Surv)},\; \text{notLow(RobustSurv)})$ | 大量にイテレーションし、圧力下で生存したものは堅牢 |
| **Rescue** | $\min(\text{high(Prod)},\; \text{low(Surv)},\; \text{high(Debt)})$ | レガシーコードを整理する高出力 |
| **Churn** | $\min(\text{notLow(Prod)},\; \text{low(Quality)},\; \text{low(Surv)})$ | 高出力だが品質が低く、常に書き直し |
| **Mass** | $\min(\text{high(Prod)},\; \text{low(Surv)})$ | 高出力だがコードが残らない |
| **Emergent** | $\min(\text{high(Gravity)},\; \text{notLow(Prod)},\; \text{low(RobustSurv)})$ | まだ実戦テストされていない新しい構造を創造中 |
| **Balanced** | $0.30$（固定） | 安定した貢献者、支配的パターンなし |
| **Spread** | $\min(\text{high(Breadth)},\; \text{low(Prod)},\; \text{low(Surv)},\; \text{low(Design)})$ | 広く存在するが、どこも浅い |

### 4.4 State軸 —— 「ライフサイクルのどこにいるか？」

| State | 信頼度の計算式 | 解釈 |
|-------|---------------|------|
| **Former** | $\min(\text{high(RawSurv)},\; \text{low(Surv)},\; \max(\text{high(Design)},\; \text{high(Indisp)}))$ | コードは残存しているが本人は非活動。重要人物だった |
| **Silent** | $\min(\text{low(Prod)},\; \text{low(Surv)},\; \text{low(Debt)})$ | 観測シグナルがすべて低い——生産・生存・負債清掃のいずれも検出されない。ロールのミスマッチや環境要因の可能性がある |
| **Fragile** | $0.85 + \frac{\text{dormantRatio} - 80}{200}$ | コードは誰も触らない場所でだけ生存 |
| **Growing** | $\min(\text{low(Prod)},\; \text{high(Quality)})$ | 低量だが高品質——成長軌道上 |
| **Active** | 最近活動している場合 $0.80$ | 現在貢献中 |

Fragileの条件: 休眠比率≥80%、Indispensability≥60、Production<40。これにより、高い生存率が幻想であるエンジニア——コードがデッドゾーンに残存しているだけ——を識別する。

### 4.5 複合ラベル

3軸は以下のようなラベルを生成する：

- **Architect Builder Active** — 活発に設計し、耐久性のある構造を構築中
- **Producer Mass Active** — 高出力だがコードが生存しない
- **Anchor Balanced Growing** — 品質重視、まだ幅を広げている途中
- **Architect Emergent Active** — まだ実証されていない新しいアーキテクチャパターンを創造中

---

## 5. チームレベル分析

### 5.1 メンバー分類

チームメンバーは3つのティアに分類される：

| ティア | 基準 | 用途 |
|--------|------|------|
| **Core** | `RecentlyActive` かつ `Impact ≥ 20` | 平均値と分布の計算 |
| **Risk** | State ∈ {Former, Silent, Fragile} | リスク検出（常に含まれる） |
| **Peripheral** | その他すべて | メトリクスから除外 |

**加重比率** がRole/Style分布に使用される：

$$w_a = \max\left(\frac{\text{Total}_a}{100},\; 0.1\right)$$

$$\text{weightedRatio}(\text{predicate}) = \frac{\sum_{a : \text{pred}(a)} w_a}{\sum_a w_a}$$

出力の多いメンバーがチームレベルの計算でより大きな重みを持つ。

### 5.2 ヘルスメトリクス

6つのヘルス指標がチーム状態の診断ビューを提供する：

#### Complementarity（補完性）

$$\text{Coverage} = \frac{|\text{ユニークロール数}|}{5} \times 80$$

$$\text{Bonus} = 10 \cdot \mathbb{1}[\text{Architect}] + 5 \cdot \mathbb{1}[\text{Anchor}] + 5 \cdot \mathbb{1}[\text{Cleaner}]$$

$$\text{Complementarity} = \text{clamp}(\text{Coverage} + \text{Bonus},\; 0,\; 100)$$

ロールの多様性を測定。5つのロールすべてとキートリオ（Architect、Anchor、Cleaner）を持つチームが100を獲得。

#### Growth Potential（成長余地）

$$\text{GrowthPotential} = \frac{|\text{Growing}|}{|\text{members}|} \times 60 + 20 \cdot \mathbb{1}[\text{Builder}] + 20 \cdot \mathbb{1}[\text{Cleaner}]$$

新しいスキルを積極的に開発しているメンバーと、メンタリング能力（Builder、Cleaner）を持つチームが高い成長余地を持つ。

#### Sustainability（持続性）

$$\text{Sustainability} = (1 - \text{RiskRatio}) \times 80 + 20 \cdot \mathbb{1}[\text{Architect}]$$

RiskRatio = Former/Silent/Fragileメンバーの割合。離脱率が低くアーキテクチャリーダーシップがあるチームは持続可能。

#### Debt Balance（負債バランス）

$$\text{DebtBalance} = \text{clamp}(\text{AvgDebtCleanup},\; 0,\; 100)$$

個人のDebt Cleanupシグナルの直接的な平均。50は中立、50以上はネットクリーニング、50以下はネット生成。

#### Productivity Density（生産性密度）

$$\text{ProductivityDensity} = \text{AvgProduction}_{\text{core}}$$

小規模チームボーナス付き：チーム≤3人で×1.2、≤5人で×1.1（AvgProduction≥50の場合）。

#### Quality Consistency（品質一貫性）

$$\text{QualityConsistency} = 0.6 \times \text{AvgQuality} + 0.4 \times \text{clamp}(100 - 2\sigma_{\text{Quality}},\; 0,\; 100)$$

高い平均品質と低い分散のバランス。全員が80%品質のチームは、95%/65%に分かれたチームより高シグナル。

### 5.3 チーム分類: 5軸システム

#### Character（性格・総合的なアイデンティティ）

| Character | 銀河アナロジー | 主要基準 | 解釈 |
|-----------|---------------|----------|------|
| **Spiral** | 渦巻銀河 — 強い中心核と活発な星形成 | アーキテクチャカバレッジ>0.4、生産性>35、均衡したAAR | 設計の重力核が生産を駆動。構造と星形成が両立 |
| **Elliptical** | 楕円銀河 — 成熟、安定、新しい星は少ない | アーキテクチャ構造 ＋ Stability文化 | 成熟し変化に強い。構造は堅固、エントロピーは低い |
| **Starburst** | スターバースト銀河 — 爆発的な星形成 | アーキテクチャ構造 ＋ Builder文化 | 急速に新領域へ拡張中。エネルギーが高く構造はまだ形成途中 |
| **Nebula** | 星雲 — 次世代の星が生まれる場所 | Builder文化 ＋ Scaling/Emergingフェーズ | 次世代エンジニアが育っている。星形成の条件が整っている |
| **Irregular** | 不規則銀河 — 重力中心がない | 高い生産性、低いアーキテクチャカバレッジ | 星がばらばらに形成。高出力だが構造的方向性がない |
| **Cluster** | 星団 — 密集だが弱い結合 | Deliveryチーム ＋ Mass Production文化 | 生産的だが構造がそれを束ねていない |
| **Collision** | 衝突銀河 — 構造的混乱 | Firefighting文化 | 力が衝突している。構造が乱れエネルギーが散逸 |
| **Dwarf** | 矮小銀河 — 小さいが長寿 | Maintenance構造 ＋ Stability文化 | コンパクトで安定。最小限のリソースで秩序を維持 |
| **Filament** | 宇宙フィラメント — 広く薄い構造 | Exploration文化 | 広い到達範囲、薄い深度。大規模構造を探索中 |

#### Structure（構造・ロール構成）

| Structure | 主要基準 |
|-----------|----------|
| **Architectural Engine** | Architect≥1、Anchor≥2、均衡したAAR、低い非構造化率 |
| **Architectural Team** | Architect≥1、Anchor≥1、低い非構造化率 |
| **Architecture-Heavy** | Architect≥1、AAR>2.0（設計が実装を上回る） |
| **Emerging Architecture** | Architect≥1、高い非構造化率 |
| **Delivery Team** | Producer>50% |
| **Maintenance Team** | Architectなし、Anchor≥40% |
| **Unstructured** | 未分類>50% |

#### Phase（フェーズ・ライフサイクル）

| Phase | 主要基準 |
|-------|----------|
| **Emerging** | Growing≥40% |
| **Scaling** | Growing 20-40%、高いGrowth Potential |
| **Mature** | Active≥80%、高いSustainability |
| **Stable** | Active≥60% |
| **Legacy-Heavy** | Riskメンバー≥30%、高い平均シグナル、Architect存在 |
| **Declining** | Riskメンバー≥30%、低いシグナルまたはArchitectなし |
| **Rebuilding** | GrowingメンバーとRiskメンバーの両方が存在 |

#### Risk（リスク・主要懸念）

| Risk | 主要基準 |
|------|----------|
| **Bus Factor** | メンバー≤5人、高い平均Indispensability |
| **Design Vacuum** | Architectなし、低いComplementarity |
| **Quality Drift** | Quality Consistency≤60 |
| **Debt Spiral** | Debt Balance≤45 |
| **Talent Drain** | リスク比率≥25% |
| **Healthy** | 重大なリスクが検出されない |

---

## 6. タイムライン分析

### 6.1 期間ベース観測

EISはリポジトリ履歴を設定可能な期間（デフォルト: 3ヶ月スパン）に分割することで経年分析をサポートする。

**アルゴリズム:**

1. リポジトリから全コミットを一度だけ収集
2. 各期間 $[t_{\text{start}}, t_{\text{end}})$ について：
   a. $\text{date} \leq t_{\text{end}}$ のコミットをフィルタ
   b. 境界コミット（$t_{\text{end}}$ 時点の最新コミット）を特定
   c. その境界コミットで `git blame` を実行: `git blame <hash> -- <file>`
   d. $\text{refTime} = t_{\text{end}}$ で全メトリクスを計算（`time.Now()` ではなく）
   e. `ActiveDays` を期間スパン全体をカバーするようオーバーライド
3. 著者別・チーム別タイムラインを組み立て

**重要な設計判断:** `ScoreAt(refTime)` 関数は、すべての最近性計算で `time.Now()` を期間の終了時刻に置き換える。これがなければ、過去の期間で全メンバーが誤って非活動と判定される。

### 6.2 遷移検出

連続する各期間ペアについて、EISはRole、Style、Stateの変化を検出する：

```
Role[t] ≠ Role[t-1] かつ どちらも "—" でない → Transition(Role, from, to, period)
Style[t] ≠ Style[t-1] かつ どちらも "—" でない → Transition(Style, from, to, period)
State[t] ≠ State[t-1] かつ どちらも "—" でない → Transition(State, from, to, period)
```

これらの遷移はキャリア軌跡を明らかにする：「Producer → Anchor」（品質への注力が発展）、「Mass → Builder」（耐久的な構築を学習）、「Active → Former」（離脱）。

### 6.3 チームタイムライン

チームレベルのタイムラインは以下を追跡する：

- 5軸すべてにわたる分類変化
- ヘルスメトリクスの軌跡
- メンバーシップ構成の変化
- 期間ごとのRole/Style/State分布

よく観測されるパターン：**Architectural Team → Maintenance Team → Architectural Engine** ——単一Architect依存からメンテナンスフェーズを経て、設計能力が分散した構造への進化。

---

## 7. 限界

### 7.1 正規化感度

相対正規化はチームメンバーの追加・離脱で全員のシグナルが変わりうることを意味する。最高貢献者は相対軸で常に100を獲得するため、これらの次元でのチーム間比較は不可能。

### 7.2 コミットメッセージ依存

品質検出はコミットメッセージの規約（`fix:`、`revert:` など）に依存する。コミットメッセージの規律が低いチームでは、Qualityシグナルの信頼性が低下する。スカッシュマージのワークフローは個人の貢献パターンを不明瞭にする可能性がある。

### 7.3 アーキテクチャパターン設定

デフォルトのアーキテクチャパターン（`*/repository/*interface*`、`*/router.go` など）はクリーンアーキテクチャの規約を反映している。異なるパターンを使用するチームは、意味のあるDesignシグナルのために設定のカスタマイズが必要。

### 7.4 モノレポの前提

blameベースの分析は単一リポジトリを前提とするか、`--recursive` モードでリポジトリ間を集約する。正規化戦略は極端に異種混合のモノレポでは適切に動作しない可能性がある。

### 7.5 パフォーマンス評価ツールではない

EISは *観測* ツールとして設計されており、評価ツールではない。限界を理解せずにパフォーマンスレビューに使用することは有害である。シグナルはコードベースで何が起きたかを反映しており、エンジニアの組織への貢献の価値を反映しているわけではない。

---

## 8. ユースケース

### 8.1 チームヘルス診断

チームリードが `eis analyze --team` を実行して以下を理解できる：
- 設計能力が集中しているか（Bus Factorリスク）分散しているか（Architectural Engine）
- チームがメンテナンスモードか、活発に進化しているか
- どのヘルスメトリクスが低下しているか

### 8.2 経年観測

`eis timeline` はポイントインタイムのスナップショットでは見えないパターンを明らかにする：
- エンジニアが6ヶ月かけてProducerからArchitectに遷移する過程
- キーメンバーの離脱後にチーム構造が劣化する様子
- 「遠慮」パターン——新しいチームに参加した際にシグナルが落ち込み、その後回復するエンジニア

### 8.3 採用とチーム構成

チームレベルのメトリクスは採用に関する質問にエビデンスベースの回答を提供する：
- 「もう一人Architectが必要か、それともAnchorか？」
- 「Complementarityシグナルは改善しているか低下しているか？」
- 「エンジニアXが抜けたらStructure分類はどうなるか？」

### 8.4 AI支援分析

JSONおよびHTML出力フォーマットはAIによる消費を想定して設計されている。`eis timeline --format json` の出力をLLMに与えることで、自然言語でのクエリが可能になる：「バックエンドチームの2024-H2に何が起きた？」 AIはシグナル変化、ロール遷移、ヘルスメトリクスの動きを相関させ、仮説を立てることができる。

---

## 9. 実装

EISはGoで実装され、単一バイナリとして配布される。

```bash
brew tap machuz/tap && brew install eis

# 個人分析
eis analyze ~/workspace/my-repo

# チーム分析
eis analyze --team ~/workspace/my-repo

# タイムライン（3ヶ月スパン、直近1年）
eis timeline --format html --output timeline.html ~/workspace/my-repo

# クロスリポジトリ分析
eis analyze --recursive ~/workspace
```

**パフォーマンス:** 500の追跡ファイルと4つの期間を持つリポジトリで、分析は約25秒かかる（`git blame` オペレーションが支配的）。blameは設定可能なワーカー数で並列化される。

**ソースコード:** [github.com/machuz/eis](https://github.com/machuz/eis)

---

## 10. 結論

EISは、Git履歴が一般的に抽出されるものよりはるかに多くのエンジニアリング貢献に関する情報を含んでいることを実証する。コミットベースのメトリクスとblameベースの生存分析、変更圧力分解を組み合わせることで、個人およびチームレベルのエンジニアリングパターンの多次元的なビューを構築することが可能である。

本研究の主要な貢献は：

1. **変更圧力分解** — 活発な開発下で生存するコードと、休眠モジュールに残存するコードの区別
2. **3軸個人分類** — 何を、どのように、ライフサイクル状態を同時に捉える
3. **5軸チーム分類** — コードレベルのデータから組織診断を提供
4. **ソフトマッチング関数** — 分類アーティファクトを生むハード閾値の回避

フレームワークはGitデータに意図的に限定されており、バージョン管理を使用するあらゆるチームに普遍的に適用可能である。その限界——正規化感度、コミットメッセージ依存、アーキテクチャパターン設定——は認識されており、設定によって管理可能である。

究極的な洞察はシンプルだ：**コードベースには重力的構造がある**。あるエンジニアが書いたコードを中心に他のコードが組み上がり、その構造が生存する。EISはこの重力を観測可能にする。

---

## 参考文献

1. Nagappan, N., Murphy, B., & Basili, V. (2008). The influence of organizational structure on software quality. *ICSE '08*.
2. Bird, C., Nagappan, N., Murphy, B., Gall, H., & Devanbu, P. (2011). Don't touch my code! Examining the effects of ownership on software quality. *ESEC/FSE '11*.
3. Forsgren, N., Humble, J., & Kim, G. (2018). *Accelerate: The Science of Lean Software and DevOps*. IT Revolution Press.
4. Tornhill, A. (2022). *Software Design X-Rays*. Pragmatic Bookshelf.

---

## 付録A: デフォルト設定

```yaml
tau: 180                    # Survival減衰定数（日）
sample_size: 500            # Debt分析でサンプリングする最大修正コミット数
debt_threshold: 10          # Debt観測の最小インタラクション数
breadth_max: 5              # Breadth軸の上限
active_days: 30             # 「最近活動」のウィンドウ
blame_timeout: 120          # ファイルごとのblameタイムアウト（秒）
production_daily_ref: 1000  # Production観測のベースライン

weights:
  production: 0.15
  quality: 0.10
  survival: 0.25
  design: 0.20
  breadth: 0.10
  debt_cleanup: 0.15
  indispensability: 0.05

bus_factor:
  critical: 0.80
  high: 0.60

exclude_file_patterns:
  - "package-lock.json"
  - "yarn.lock"
  - "go.sum"
  - "docs/swagger*"
  - "*generated*"
  - "mock_*"
  - "*.gen.*"

architecture_patterns:
  - "*/repository/*interface*"
  - "*/domainservice/"
  - "*/router.go"
  - "*/middleware/"
  - "di/*.go"
  - "*/core/"
  - "*/stores/"
  - "*/hooks/"
  - "*/types/"

blame_extensions:
  - "*.go"
  - "*.ts"
  - "*.tsx"
  - "*.py"
  - "*.rs"
  - "*.java"
  - "*.rb"
```

---

## 付録B: 用語集

| 用語 | 定義 |
|------|------|
| **AAR** | Architect-to-Anchor Ratio。設計ロールと安定化ロールのバランスを測定。 |
| **Architecture Coverage** | (Architect数 + Anchor数) / メンバー数。構造的に貢献するメンバーの割合。 |
| **Anchor Density** | Anchor数 / メンバー数。品質安定化メンバーの割合。 |
| **Change Pressure** | コミット数 / blameの行数（モジュール単位）。モジュールの開発活発度を示す。 |
| **Core Member** | 最近活動があり、Total≥20。チーム平均に含まれる。 |
| **Gravity** | Indispensability、Breadth、Designの複合。構造的影響を測定。 |
| **Risk Member** | State ∈ {Former, Silent, Fragile}。リスク計算に含まれる。 |
| **Robust Survival** | 高圧力モジュールのblame行（時間減衰付き）。協働によって実証されたコード。 |
| **Dormant Survival** | 低圧力モジュールのblame行（時間減衰付き）。協働によるテストを経ていないコード。 |
| **Tau (τ)** | Survival計算の指数減衰定数。デフォルト180日。 |

---

**ライセンス:** MIT

**引用:**
```
@software{eis2026,
  title = {Engineering Impact Signal},
  author = {machuz},
  url = {https://github.com/machuz/eis},
  year = {2026}
}
```
