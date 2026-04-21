# git考古学 #1 —— 履歴だけでエンジニアの「戦闘力」を定量化する

*コミット数、PR数、行数——なぜ既存の指標ではエンジニアの本当の強さが見えないのか*

![7軸シグナル可視化](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/hatena/ch1-iconic.png)

## スコアからシグナルへ

スコアという概念に、ずっと違和感があった。

誰かが注ぎ込んだ努力、長く残る構造を作る力、チームへの影響。可視化したいものはたくさんある。だがこれらはコードベースの状態、プロジェクトの方向性、チームの力学、さらには個人的な事情まで——あまりにも多くの変数に左右される。

**コードだけから、すべてを一つの絶対スコアに還元する方法はない。**

しかし、気づいたことがある。絶対的なスコアリングは不可能でも、**強さの変化**は観測できる。

あるタイミングではあるエンジニアがシステムを駆動している。別のタイミングでは別の誰かがその役割を担う。これはスコアではない。**これはシグナルだ。**

人やシステムに固定値を割り振ることはできない。だが、観測はできる——どこに勢いがあるか、どこに圧力がかかっているか、どこで安定が形成され、どこで崩れているか。

スコアは評価しようとする。**シグナルは明らかにする。**

僕はこれをエンジニアの**「戦闘力」**と半ば冗談で呼んでいる。正式名称は **Engineering Impact Signal**——**EIS**（エイス、"ace" と読む）。ただし、実際に測定しているのはもっと精密なもの——

> **コードベース自体に記録された、観測可能な構造的シグナル**

---

## 核心：生き残ったコードが最も重要

最強のエンジニアは、ただコードを書くだけではない。**数ヶ月後も書き直される必要なく存在し続ける**コードを書く。

だからこのモデルで最も重要なシグナルは**コードの生存（Survival）**だ。

ただし、生存も慎重に扱う必要がある。素朴な `git blame` は初期のコントリビュータを過大評価してしまう。これを補正するために、モデルは**時間減衰型の生存率**を適用する——最近のコードは古いコードよりもはるかに重く評価される。

| 経過日数 | 重み |
|---|---|
| 7日 | 0.96 |
| 30日 | 0.85 |
| 90日 | 0.61 |
| 180日 | 0.37 |
| 1年 | 0.13 |
| 2年 | 0.02 |

![730日にわたる時間減衰Survival重み曲線（指数関数的減衰）](https://raw.githubusercontent.com/machuz/eis/main/docs/images/survival-decay-curve.png?v=5)

退職したメンバーのシグナルは時間とともに自然に減衰する。これは「歴史的に最も多くのコードを書いた人」ではなく、**「今、耐久性のあるコードを書いている人」**を近似する。

### Dormant vs Robust——決定的な区別

コードが「生き残っている」ことには、まったく異なる二つの意味がある：

- **Dormant Survival（休眠生存）**: 誰も触らないモジュールにコードが残っている。耐久性があるのではなく、ただ放置されているだけ
- **Robust Survival（堅牢生存）**: **他のエンジニアが活発に変更を加えているファイル**にコードが残っている。実際の変更圧の中で生き残ったコードだけがカウントされる

この区別はEISの最も重要なイノベーションだ。全体のSurvivalは低いがRobust Survivalはそこそこあるエンジニアは、激しくイテレーションしながらも変更に強いコードを生み出している（**Resilient** スタイル）。逆に、Survivalは高いがRobust Survivalが低い場合、そのコードは誰も触らないから生き残っているだけだ（**Fragile** ステート）。

**時間減衰型Survivalはゲーミングに耐性がある。** 忙しさだけではインパクトを膨らませられない——数ヶ月後もコードベースに残ったコードだけがカウントされる。そしてDebt Cleanup軸の存在により、他人に仕事を生み出すことで高いインパクトを得ることは構造的に不可能になっている。

---

## 7つの軸で観測する

![EIS Framework Overview：Git履歴から7軸シグナル・3軸トポロジー・Gravityを算出](https://raw.githubusercontent.com/machuz/eis/main/docs/images/engineering-impact-framework-diagram-fixed.png?v=5)
*Git履歴 → 7軸シグナル → 3軸トポロジー（Role/Style/State） → Gravity*

| 軸 | 重み | 観測対象 |
|---|---|---|
| Production | 15% | 日あたりの変更量（絶対値） |
| Quality | 10% | fix/revertコミットの少なさ |
| Survival | **25%** | 現在も存在するコード（時間減衰付き） |
| Design | 20% | アーキテクチャファイルへの貢献 |
| Breadth | 10% | 関与リポジトリ数 |
| Debt Cleanup | 15% | 他人が生んだ問題の修正 |
| Indispensability | 5% | バスファクターリスク |

**Survivalが最大の重み（25%）を持つ**——核心のテーゼだ：*あなたは長く残る設計を書いているか？*

二つの軸は特に注目に値する：

**Debt Cleanup** ——この指標を追加したとき、チームの「サイレントヒーロー」が可視化された。他の全員のバグを黙々と直し続けている人がいた。まさにこういう人を、従来の指標は見えなくしてしまう。

**Design** ——アーキテクチャファイル（リポジトリインターフェース、ドメインサービス、ルーター、ミドルウェア）への頻繁なコミットは、アーキテクチャへの関与を示す。判断が*正しかったか*ではなく、**誰が構造を形作ることに参加しているか**。これらのパターンはプロジェクトごとにカスタマイズが必要——設定作業自体が設計の会話になる。

> 詳細な計算式と正規化手法：[Whitepaper](https://github.com/machuz/eis)

---

## Impactと観測モデル

シグナルを重み付けして一つの**Impact**値にまとめる：

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
| 80+ | 超新星。チームに1〜2人いるかどうか |
| 60–79 | コア級。強い |
| **40–59** | **シニアレベル。40+は本当に強い** |
| 30–39 | ミドル |
| 20–29 | ジュニア〜ミドル |
| <20 | ジュニア |

**40 = シニア。** 7つの軸で同時にそこそこの数値を出すには、本物の総合力が必要だ。40台に乗るエンジニアはどの市場でも戦える。

**重要な注意：** EISが測定しているのは**このコードベースへのインパクト**であり、エンジニアリング能力の絶対値ではない。Survivalが高いことが、リファクタリングできないコードを意味している場合すらある。観測結果が直感と合わない場合、調べる価値がある——それは人の問題ではなく、コードベースの設計上の問題を示しているかもしれない。（これを[Engineering Relativity](https://ma2k8.hateblo.jp/entry/2026/03/14/233602)と呼んでいる。）

![Impact Guide](https://raw.githubusercontent.com/machuz/eis/main/docs/images/score-guide.png?v=5)

---

## 3軸トポロジー

シグナルを算出すると、認識可能なパターンが浮かび上がる。EISはエンジニアのトポロジーを**3つの独立した軸**に分解する：

![Engineer Topology](https://raw.githubusercontent.com/machuz/eis/main/docs/images/engineering-archetypes-paper-figure.png?v=5)

### Role——*何を*貢献するか

| Role | シグナル | 説明 |
|---|---|---|
| **Architect** | Design↑ Surv↑ | システム構造を形作る。設計が堅実だからコードが生き残る |
| **Anchor** | Qual↑ Surv↑ Debt↑ | コードベースを安定させる。耐久性のあるコードを書き、他者のバグを静かに直す |
| **Cleaner** | Debt↑ | 主に他者が生んだ負債を修正する |
| **Producer** | Prod↑ | 高い出力量。その出力が*良い*かどうかはStyleとStateに依存する |
| **Specialist** | Indisp↑ Breadth↓ | 狭い領域での深い専門性 |

### Style——*どう*貢献するか

| Style | シグナル | 説明 |
|---|---|---|
| **Builder** | Prod↑ Surv↑ Design↑ | 設計し、大量に書き、かつメンテナンスする |
| **Resilient** | Prod↑ RobustSurv○ | 激しくイテレーションするが、変更圧の中で生き残るコードは堅牢 |
| **Rescue** | Prod↑ Surv↓ Debt↑ | レガシーコードの引き継ぎとクリーンアップ |
| **Churn** | Prod○ Qual↓ Surv↓ | アウトプットはあるがSurvivalがない |
| **Mass** | Prod↑ Surv↓ | 大量に書くが何も残らない |
| **Balanced** | 均等分布 | バランス型 |
| **Spread** | Breadth↑ Prod↓ Surv↓ | 広く浅く。深さがない |

### State——*ライフサイクルフェーズ*

| State | シグナル | 説明 |
|---|---|---|
| **Active** | 直近コミットあり、Surv↑ | 現在進行形で耐久性のあるコードを書いている |
| **Growing** | Qual↑ Prod↓ | 出力は少ないが品質が高い。成長中 |
| **Former** | Raw Surv↑ Surv↓ | コードは残っているが著者は非アクティブ。引き継ぎ優先 |
| **Silent** | Prod↓ Surv↓ Debt↓ | 全シグナルが低い。ロールミスマッチ、またはこの人の強みが活きていない環境の可能性 |
| **Fragile** | Surv↑ Prod↓ Qual<70 | 誰も変更しないからコードが残っているだけ |

> トポロジー進化の詳細：[Chapter 3](https://ma2k8.hateblo.jp/entry/2026/03/14/135648) と [Chapter 4](https://ma2k8.hateblo.jp/entry/2026/03/14/155124)

---

## Engineer Gravity

Impactは*どれだけ強いか*を示す。トポロジーは*どんな種類か*を示す。**Gravity**はその人がどれだけの構造的影響力を持つかを示す。

```
Gravity = Indispensability × 0.40 + Breadth × 0.30 + Design × 0.30
```

高いGravityが自動的に良いわけではない——**健全性の次元**がある：

```
health = Quality × 0.6 + RobustSurvival × 0.4

Gravity < 20  → dim gray （低い影響力）
health ≥ 60   → green    （健全なGravity）
health ≥ 40   → yellow   （中程度）
health < 40   → red      （脆いGravity——危険）
```

Red gravityは**「システムがこの人に依存しており、かつコードが脆い」**ことを意味する。最も危険な組み合わせが、一目で分かる。

---

## 実際のチームで動かした結果

自分のチーム（14リポジトリ、10人以上のエンジニア）で実行した。シグナルはチームの体感とほぼ完璧に一致した。

![Backend Rankings](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ch1-backend-table.png?v=4)

**R.S.** —— Production 17は目を引かない。しかしSurvival 50（チーム2位）は、最近のコードが残っていることを意味する。Debt Cleanup 88は、他の全員のバグを黙々と直しているということだ。**Debt Cleanupが可視化するために設計された、まさにこういう人だ。** AnchorというRoleがこれを完璧に捉えている。

**Y.Y.** —— Design 67、Breadth 81。オリジナルのアーキテクト。Indispensability 100——アクティブなメンバーの誰よりも多くのモジュールがまだこの人に帰属している。トポロジーは `Architect / — / Former`——Roleはコードの中で、退職後もなお持続している。これは[Chapter 4](https://ma2k8.hateblo.jp/entry/2026/03/14/155124)が「成仏させるべき魂」と呼ぶシグナルだ——引き継ぎの優先対象。

**Z.** —— Impact 24.9。唯一高かったシグナルはBreadth。トポロジーは `— / Spread / —`——広くて浅い。このパターンの早期観測は、より適切なロールアライメントの判断材料になり得た。

---

## 既存の指標が見逃していること

DORAはデプロイ速度を測る。SPACEはサーベイを使う。Git分析ツールはコードが*いつ*書かれたかを追跡する。**そのどれも、コードが実際に生き残ったかどうかを問わない。**

EISはそのギャップを埋める：時間減衰Survival + Robust/Dormant分離 + Debt Cleanupの追跡。すでに手元にあるデータだけで。

---

## 限界と誠実さ

このモデルは**人間の価値を測るものではない**。コードベースに観測可能な技術的影響力を推定するものだ。

エンジニアはgitが記録できない形でも貢献する：メンタリング、ドメイン知識、ドキュメンテーション、心理的安全性。EISはgitが記録したものだけを捉える——それ以上でも以下でもない。

低いImpactは弱いエンジニアを意味しない。曖昧な仕様、組織の摩擦、雑な計画はすべてシグナルを減少させる。チーム全体のImpactが低いなら、個人を調べる前に組織を調べるべきだ。

モデルの精度はコードベースの設計品質に比例する。カオスなコードベースでは、高いSurvivalが単にデッドコードを意味しているかもしれない。**指標の精度が低いこと自体がシグナルだ。**

---

## 使ってみる

```bash
❯ brew tap machuz/tap && brew install eis
❯ eis analyze --recursive ~/projects
```

AIトークン不要。APIキー不要。`git log` と `git blame` だけ。

![Terminal Output](https://raw.githubusercontent.com/machuz/eis/main/docs/images/terminal-output.png?v=0.11.0)

本当の価値は**経時変化の追跡**から生まれる。Survivalが四半期ごとに上昇していれば、設計スキルが成長している。Debt Cleanupが上昇していれば、チームへの貢献が増えている。

> 完全な方法論：[Whitepaper](https://github.com/machuz/eis) · [README](https://github.com/machuz/eis)

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-full.png?v=2)

**GitHub**: [eis](https://github.com/machuz/eis) — CLIツール、計算式、方法論すべてオープンソース。`brew tap machuz/tap && brew install eis` でインストール。

もし役に立ったら：[Sponsor on GitHub](https://github.com/sponsors/machuz)

---

### シリーズ

- [Chapter 0: もしGit履歴が「誰が最も強いエンジニアか」を教えてくれるとしたら？](https://ma2k8.hateblo.jp/entry/2026/03/18/150222)
- **Chapter 1: 履歴だけでエンジニアの「戦闘力」を定量化する**
- [Chapter 2: 個人を超えて：Git履歴からチームの健全性を測る](https://ma2k8.hateblo.jp/entry/2026/03/13/060851)
- [Chapter 3: アーキテクトへの2つの道：エンジニアはどう進化するか](https://ma2k8.hateblo.jp/entry/2026/03/14/135648)
- [Chapter 4: BEアーキテクトは収斂する：魂を成仏させる聖なる仕事](https://ma2k8.hateblo.jp/entry/2026/03/14/155124)
- [Chapter 5: タイムライン：シグナルは嘘をつかない、そして躊躇も記録する](https://ma2k8.hateblo.jp/entry/2026/03/14/180329)
- [Chapter 6: チームは進化する：タイムラインが明かす組織の法則](https://ma2k8.hateblo.jp/entry/2026/03/14/184223)
- [Chapter 7: コードの宇宙を観測する](https://ma2k8.hateblo.jp/entry/2026/03/14/213413)
- [Chapter 8: Engineering Relativity：同じエンジニアが異なるシグナルを得る理由](https://ma2k8.hateblo.jp/entry/2026/03/14/233602)
- [Chapter 9: 起源：コード宇宙のビッグバン](https://ma2k8.hateblo.jp/entry/2026/03/15/054313)
- [Chapter 10: ダークマター：見えない重力](https://ma2k8.hateblo.jp/entry/2026/03/15/062608)
- [Chapter 11: エントロピー：宇宙は常に無秩序に向かう](https://ma2k8.hateblo.jp/entry/2026/03/15/062609)
- [Chapter 12: 崩壊：良いアーキテクトとブラックホールエンジニア](https://ma2k8.hateblo.jp/entry/2026/03/15/062610)
- [Chapter 13: コードの宇宙論](https://ma2k8.hateblo.jp/entry/2026/03/15/062611)
- [Chapter 14: 文明——なぜ一部のコードベースだけが文明になるのか](https://ma2k8.hateblo.jp/entry/2026/03/15/215211)
- [Chapter 15: AIは星を生む、重力ではなく](https://ma2k8.hateblo.jp/entry/2026/03/15/221250)
- [最終章: 重力を形作るエンジニアたち](https://ma2k8.hateblo.jp/entry/2026/03/15/231040)

---

[Chapter 2: チームの健全性 →](https://ma2k8.hateblo.jp/entry/2026/03/13/060851)
