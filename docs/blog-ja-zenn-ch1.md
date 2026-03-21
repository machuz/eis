---
title: "git考古学 #1 —— 履歴だけでエンジニアの「戦闘力」を定量化する"
emoji: "⚔️"
type: "tech"
topics: ["git", "engineering", "productivity"]
published: true
---

*コミット数、PR数、行数——なぜそれらではエンジニアの本当の強さが見えないのか*

![7軸シグナル可視化](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch1-iconic.png?v=4)

## スコアからシグナルへ

スコアには、ずっと居心地の悪さを感じていた。

努力、構造を作る力、チームへのインパクト——そういうものを可視化したかった。でもそこには変数が多すぎる。コードベースの状態、プロジェクトの方向性、チームの力学、個人の事情。

**コードだけから、これらすべてを一つの絶対スコアに還元する方法は存在しない。**

しかし、あることに気づいた。絶対的なスコアリングは不可能でも、**強さの変化**を観測することはできる。

ある時点では、あるエンジニアがシステムを牽引している。別の時点では、別の誰かがその役割を担う。これはスコアではない。**これはシグナルだ。**

人間にもシステムにも固定値は割り当てられない。しかし観測はできる——どこにモメンタムがあるか、どこに圧力が蓄積しているか、どこで安定が形成されているか、あるいは崩壊しているか。

スコアは評価しようとする。**シグナルは明らかにする。**

冗談半分でエンジニアの**「戦闘力」**と呼んでいる。正式名称は **Engineering Impact Signal**——**EIS**（エイス、"ace" と読む）。しかし実際に測っているのは、もっと精確なものだ：

> **コードベース自体に記録された、観測可能な構造的シグナル**

---

## 核心：残り続けるコードこそが強さ

最強のエンジニアは、ただコードを書くだけではない。**数ヶ月後も書き直されることなく存在し続けるコード**を書く。

だからこのモデルで最も重要なシグナルは**コードの生存力（Survival）**だ。

ただし、生存力の扱いには注意が必要だ。素のgit blameは初期の貢献者に有利になる。この問題を補正するために、モデルは**時間減衰付き生存力**を適用する——最近のコードは古いコードよりはるかに重く数える。

| 経過日数 | 重み |
|---|---|
| 7日 | 0.96 |
| 30日 | 0.85 |
| 90日 | 0.61 |
| 180日 | 0.37 |
| 1年 | 0.13 |
| 2年 | 0.02 |

![時間減衰Survival重み曲線：730日間の指数減衰](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/survival-decay-curve.png?v=5)

これにより、退職したメンバーのシグナルは時間とともに自然に減衰する。歴史的に最もコードを書いた人ではなく、**今、耐久性のあるコードを書いている人**を近似する。

### Dormant（休眠）vs Robust（堅牢）——最も重要な区別

コードが「生存している」には、まったく異なる2つの意味がある：

- **Dormant Survival**：誰も触らないモジュールに残っているコード。耐久性があるのではなく、ただ放置されているだけ
- **Robust Survival**：**他のエンジニアが活発に変更しているファイル**の中で残り続けるコード。実際の変更圧の下で生き残ったコードだけがカウントされる

この区別こそがEISの最も重要な革新だ。全体のSurvivalは低いがRobust Survivalはそこそこあるエンジニアは、激しく試行錯誤しながら変更圧に強いコードを生み出している（**Resilient** スタイル）。逆に、Survivalは高いがRobust Survivalが低い場合、コードは誰も触らないから残っているだけ（**Fragile** ステート）。

**時間減衰付き生存力はゲーミングに強い。** 忙しそうに見せる作業ではシグナルは膨らまない——数ヶ月後にコードベースに残っているコードだけがカウントされる。そして負債清掃軸があることで、他者に仕事を発生させてImpactを上げることも構造的に不可能だ。

---

## 7軸の観測

![EIS Framework Overview](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/engineering-impact-framework-diagram-fixed.png?v=5)
*Git履歴から7軸シグナル・3軸トポロジー（Role/Style/State）・Gravityを算出*

| 軸 | 重み | 何を観測するか |
|---|---|---|
| Production | 15% | 日あたりの変更量（絶対尺度） |
| Quality | 10% | fix/revertコミットの少なさ |
| Survival | **25%** | 今も存在するコード（時間減衰付き） |
| Design | 20% | アーキテクチャファイルへの貢献 |
| Breadth | 10% | 関与しているリポジトリ数 |
| Debt Cleanup | 15% | 他者が作った問題の修正 |
| Indispensability | 5% | バスファクターリスク |

**Survivalが最高の重み（25%）を持つ**——これがこのモデルの核心的命題だ：*あなたは、残る設計を書いているか？*

2つの軸に特に注目してほしい：

**Debt Cleanup** ——この指標を入れたとき、チームの「静かなヒーロー」が可視化された。他の全員のバグを黙々と直し続けている人。まさにこういう人を、従来の指標は見えなくしていた。

**Design** ——アーキテクチャファイル（リポジトリインターフェース、ドメインサービス、ルーター、ミドルウェア）への頻繁なコミットは、構造設計への関与を示す。意思決定が*正しかったか*ではなく、**誰が構造の形成に参加しているか**。このパターンはプロジェクトごとにカスタマイズする必要があり、その設定作業自体が設計の対話になる。

> 詳細な計算式と正規化：[Whitepaper](https://github.com/machuz/eis)

---

## Impactと観測モデル

シグナルを重み付けして一つの**Impact**値に集約する：

```
impact =
  production       × 0.15
  + quality        × 0.10
  + survival       × 0.25
  + design         × 0.20
  + breadth        × 0.10
  + debt_cleanup   × 0.15
  + indispensability × 0.05
```

スケールは意図的に厳しい：

| Impact | 評価 |
|---|---|
| 80+ | Supernova。チームに1〜2人がせいぜい |
| 60〜79 | Near-core。強い |
| **40〜59** | **シニアレベル。40+は本当に強い** |
| 30〜39 | ミドルレベル |
| 20〜29 | ジュニア〜ミドル |
| <20 | ジュニア |

**40 = シニア。** 7つの軸で同時にそこそこの数字を出すには、本物の総合力が必要だ。40台のエンジニアはどの市場でも通用する。

**重要な注意：** EISは**このコードベースにおけるインパクト**を測定するものであり、絶対的なエンジニアリング能力ではない。高いSurvivalは、リファクタリングできないコードを意味しているかもしれない。観測結果が直感と合わない場合、それは調査に値する——人の問題ではなくコードベースの設計の問題を明らかにしている可能性がある（これを[Engineering Relativity](https://ma2k8.hateblo.jp/entry/2026/03/14/233602)と呼ぶ）。

![Impact Guide](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/score-guide.png?v=5)

---

## 3軸トポロジー

シグナルを算出すると、認識可能なパターンが浮かび上がる。EISはエンジニアのトポロジーを**3つの独立した軸**に分解する：

![エンジニアトポロジー](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/engineering-archetypes-paper-figure.png?v=5)

### Role——*何を*貢献するか

| Role | シグナル | 説明 |
|---|---|---|
| **Architect** | Design↑ Surv↑ | システム構造を形成。設計が健全だからコードが残る |
| **Anchor** | Qual↑ Surv↑ Debt↑ | コードベースを安定させる。耐久性のあるコードを書き、他者のバグを静かに直す |
| **Cleaner** | Debt↑ | 主に他者が生んだ負債を修復 |
| **Producer** | Prod↑ | 高い産出量。その産出が*良い*かはStyleとStateによる |
| **Specialist** | Indisp↑ Breadth↓ | 狭い領域での深い専門性 |

### Style——*どう*貢献するか

| Style | シグナル | 説明 |
|---|---|---|
| **Builder** | Prod↑ Surv↑ Design↑ | 設計し、大量に構築し、かつ維持する |
| **Resilient** | Prod↑ RobustSurv○ | 激しく試行錯誤するが、変更圧の下で残るものは堅牢 |
| **Rescue** | Prod↑ Surv↓ Debt↑ | レガシーコードの引き取りと浄化 |
| **Churn** | Prod○ Qual↓ Surv↓ | 生産はあるが生存しない |
| **Mass** | Prod↑ Surv↓ | 大量に書くが何も残らない |
| **Balanced** | 均等な分布 | バランス型 |
| **Spread** | Breadth↑ Prod↓ Surv↓ | 広く存在するが深さがない |

### State——*ライフサイクルフェーズ*

| State | シグナル | 説明 |
|---|---|---|
| **Active** | 直近コミット Surv↑ | 今、耐久性のあるコードを書いている |
| **Growing** | Qual↑ Prod↓ | 産出は少ないが品質が高い。レベルアップ中 |
| **Former** | Raw Surv↑ Surv↓ | コードは残るが著者は不活動。引き継ぎ優先 |
| **Silent** | Prod↓ Surv↓ Debt↓ | すべてのシグナルが低い。ロールのミスマッチか、その人の強みを活かせていない環境の可能性 |
| **Fragile** | Surv↑ Prod↓ Qual<70 | コードは誰も変更しないから残っているだけ |

> トポロジーの進化の深掘り：[第3章](https://ma2k8.hateblo.jp/entry/2026/03/14/135648)と[第4章](https://ma2k8.hateblo.jp/entry/2026/03/14/155124)

---

## エンジニア重力

Impactは*どれくらい強いか*を教える。トポロジーは*どんなタイプか*を教える。**Gravity（重力）**は、その人がどれだけの構造的影響力を持つかを教える。

```
Gravity = Indispensability × 0.40 + Breadth × 0.30 + Design × 0.30
```

高いGravityは自動的に良いことではない——**健全性の次元**がある：

```
health = Quality × 0.6 + RobustSurvival × 0.4

Gravity < 20  → dim gray  （低い影響力）
health ≥ 60   → green     （健全な重力）
health ≥ 40   → yellow    （中程度）
health < 40   → red       （脆い重力——危険）
```

赤い重力は**「システムがこの人に依存している、かつそのコードが脆い」**を意味する。最も危険な組み合わせが、一瞬で可視化される。

---

## 実測結果

自チーム（14リポ、10名以上のエンジニア）で計測した。シグナルはチームの肌感とほぼ完全に一致した。

![Backend Rankings](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch1-backend-table.png?v=4)

**R.S.** ——Production 17は目立つ数字ではない。しかしSurvival 50（チーム2位）は最近書いたコードが残り続けていることを意味する。Debt Cleanup 88は他の全員のバグを黙々と直していること。**まさにこういう人を可視化するために、Debt Cleanupを設計した。** Anchor Roleがこれを的確に捉えている。

**Y.Y.** ——Design 67、Breadth 81。元のアーキテクト。Indispensability 100——現役メンバーの誰よりも多くのモジュールがこの人に帰属している。トポロジーは `Architect / — / Former`——Roleはコードに残り続け、Stateだけが遷移した。[第4章](https://ma2k8.hateblo.jp/entry/2026/03/14/155124)で「弔うべき魂」と呼ぶもの——批判ではなく、引き継ぎの優先度を示すシグナルだ。

**Z.** ——Impact 24.9。Breadthだけが唯一高いシグナル。トポロジーは `— / Spread / —`——広く存在するが深さがないパターン。このシグナルを早期に観測していれば、より適切な役割設計ができた可能性がある。

---

## 既存メトリクスが見落としていること

DORAはデプロイ速度を測る。SPACEはサーベイを使う。Git分析ツールはコードが*いつ*書かれたかを追跡する。**そのコードが実際に生き残ったかどうかを問うものは、どれもない。**

EISがそのギャップを埋める：時間減衰付き生存力 + Robust/Dormant分離 + 負債清掃の追跡。すべて既に手元にあるデータから。

---

## 限界と誠実さ

このモデルは人間の価値を測るものでは**ない**。コードベースで観測可能な技術的影響力を推定するものだ。

gitが捉えられない貢献は無数にある：メンタリング、ドメイン知識、ドキュメント、心理的安全性。EISはgitが記録するものだけを捉える——それ以上でもそれ以下でもない。

低いImpactは弱いエンジニアを意味しない。曖昧な仕様、組織の摩擦、プランニングの甘さ——これらすべてがシグナルを下げる。チーム全体のImpactが低いなら、個人を調べる前に組織を調べるべきだ。

モデルの精度はコードベースの設計品質に比例する。混沌としたコードベースでは、高いSurvivalはただの死んだコードかもしれない。**指標の精度が低いこと自体が、一つのシグナル**だ。

---

## 使ってみる

```bash
❯ brew tap machuz/tap && brew install eis
❯ eis analyze --recursive ~/projects
```

AIトークンゼロ。APIキーゼロ。`git log` と `git blame` だけ。

![ターミナル出力](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/terminal-output.png?v=0.11.0)

本当の価値は**時系列での変化を追跡すること**にある。四半期ごとにSurvivalが上がっていれば設計力が伸びている。Debt Cleanupが上がっていればチーム貢献が増えている。

> 詳細な方法論：[Whitepaper](https://github.com/machuz/eis) · [README](https://github.com/machuz/eis)

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/logo-full.png?v=2)

**GitHub**: [eis](https://github.com/machuz/eis) — CLIツール、計算式、方法論すべてオープンソース。`brew tap machuz/tap && brew install eis` でインストール。

この記事が参考になったら：

[![Sponsor](https://img.shields.io/badge/Sponsor-❤-ea4aaa?logo=github&style=for-the-badge)](https://github.com/sponsors/machuz)

PayPay: `w_machu7`

---

### シリーズ

- **第1章：履歴だけでエンジニアの「戦闘力」を定量化する**（本記事）
- [第2章：エンジニアの「戦闘力」から、チームの「構造力」へ](https://ma2k8.hateblo.jp/entry/2026/03/13/060851)
- [第3章：Architectには流派がある：Git履歴が暴く進化の分岐モデル](https://ma2k8.hateblo.jp/entry/2026/03/14/135648)
- [第4章：Backend Architectは収束する：成仏という聖なる仕事](https://ma2k8.hateblo.jp/entry/2026/03/14/155124)
- [第5章：タイムライン：シグナルは嘘をつかないし、遠慮も映る](https://ma2k8.hateblo.jp/entry/2026/03/14/180329)
- [第6章：チームは進化する——タイムラインが暴く組織の法則](https://ma2k8.hateblo.jp/entry/2026/03/14/184223)
- [第7章：コードの宇宙を観測する](https://ma2k8.hateblo.jp/entry/2026/03/14/213413)
- [第8章：Engineering Relativity：なぜ同じエンジニアでもImpactが変わるのか](https://ma2k8.hateblo.jp/entry/2026/03/14/233602)
- [第9章：Origin：コード宇宙のビッグバン](https://ma2k8.hateblo.jp/entry/2026/03/15/054313)
- [第10章：Dark Matter：見えない重力](https://ma2k8.hateblo.jp/entry/2026/03/15/062608)
- [第11章：Entropy：宇宙は常に無秩序に向かう](https://ma2k8.hateblo.jp/entry/2026/03/15/062609)
- [第12章：Collapse：良いArchitectとBlack Hole Engineer](https://ma2k8.hateblo.jp/entry/2026/03/15/062610)
- [第13章：Cosmology of Code：コード宇宙論](https://ma2k8.hateblo.jp/entry/2026/03/15/062611)
- [第14章：Civilization：なぜ一部のコードベースだけが文明になるのか](https://ma2k8.hateblo.jp/entry/2026/03/15/215211)
- [第15章：AI Creates Stars, Not Gravity](https://ma2k8.hateblo.jp/entry/2026/03/15/221250)
- [最終章：The Engineers Who Shape Gravity：重力を作るエンジニアたち](https://ma2k8.hateblo.jp/entry/2026/03/15/231040)

---

[第2章：チームの「構造力」 →](https://ma2k8.hateblo.jp/entry/2026/03/13/060851)
