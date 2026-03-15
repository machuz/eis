# git考古学 #12 —— Collapse：良いArchitectとBlack Hole Engineer

![Cover](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/hatena/cover-ch12.png)

*宇宙にはもう一つの性質がある。崩壊だ。*

### 前章までのあらすじ

[第11章](https://ma2k8.hateblo.jp/entry/2026/03/15/062609)ではEntropy——宇宙は常に無秩序に向かう——について書いた。

今回は、重力のもう一つの性質について書く。

崩壊だ。

---

## 恒星は永遠ではない

宇宙にはもう一つの性質がある。

崩壊だ。

恒星は永遠ではない。銀河も永遠ではない。

重力が崩れると、宇宙の構造は一気に変わる。

コードベースでも同じ現象が起きる。

---

## Architectが去るとき

Architectは宇宙を作る。

設計を定義し、抽象化を作り、依存関係を整え、重力の中心を作る。

しかし重要なのはここだ。

**本当に優れたArchitectは「自分がいなくなった後の宇宙」まで設計している。**

良いArchitectが作った宇宙では、その人が去った後でも秩序が維持される。

構造が残るからだ。

設計の重力場が宇宙に残っている。

---

## Black Hole Engineer

しかし、すべての強い重力が良い重力とは限らない。

宇宙にはブラックホールが存在する。ブラックホールは極端に強い重力を持つ天体だ。しかしその重力は——構造を作るのではなく、すべてを吸い込む。

コード宇宙にも同じタイプのエンジニアが存在する。

**Black Hole Engineer。**

特徴はこうだ。

![Black Hole Pattern](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch12-black-hole-pattern.svg)

技術力は高い。生産量も多い。影響力も強い。

しかし——構造を作らない。

代わりに——依存が集中する。

---

## ブラックホール型の宇宙

Black Hole Engineerの周りではこういう現象が起きる。

巨大なサービス。巨大なユーティリティ。巨大なモジュール。

仕事が集まり、依存が集まり、コードが集まる。

結果として——**宇宙の中心が一人のエンジニアになる。**

![Good Architect vs Black Hole](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch12-good-vs-blackhole.svg)

良いArchitectは重力を分散させる。構造を残し、宇宙に秩序を与える。

Black Hole Engineerは重力を集中させる。自分自身が宇宙の中心になる。

---

## 崩壊

問題は、そのエンジニアが去ったときだ。

ブラックホールが消えると、宇宙の中心が消える。

すると何が起きるか。

![Collapse Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch12-collapse-timeline.svg)

設計判断が止まる。依存関係が壊れる。誰も触れなくなる。

コード宇宙は一気に崩壊する。

---

## 良い重力

良いArchitectはブラックホールとは違う。

彼らは重力を集中させない。構造を分散させる。

抽象化を共有し、境界を明確にし、宇宙の秩序を残す。

だから——去った後でも宇宙は崩壊しない。

これが**練度の高い重力**である。

第4章の「成仏」を思い出してほしい。良いArchitectは成仏できる。去った後もコードが生き続ける。Survival 100は、構造が残っている証拠だ。

**Black Hole Engineerは成仏できない。**

去った瞬間にコード宇宙が崩壊するからだ。

---

## EISでCollapseを防ぐ

### 1. Bus Factorを監視する

`eis analyze --team` のBus Factorが1に近いチームは、Black Hole型の危険がある。

```bash
❯ eis analyze --team --recursive ~/workspace
```

Bus Factor = 1は「一人が去ったら崩壊する」という意味だ。これがBlack Holeの最も明確な兆候である。

### 2. Indispensabilityの集中を検出する

`--per-repo` で個人のスコア分布を見る。

```bash
❯ eis analyze --recursive --per-repo ~/workspace
```

一人だけIndispensabilityが極端に高く、他のメンバーが極端に低い——この分布がBlack Hole型の特徴だ。

### 3. タイムラインで「一人だけ突出し続ける」パターンを警戒する

```bash
❯ eis timeline --author engineer-x --periods 0 ~/workspace
```

良いArchitectのタイムラインでは、Architect → Producerへの移行が見られる（第5章のO.のパターン）。構造を作り終えたら、その上で生産に移る。

Black Hole Engineerのタイムラインでは、**ずっとArchitectのまま**だ。構造を手放さない。重力を集中させ続ける。

### 4. 周囲のスコアで重力の質を判定する

第8章の重力レンズ効果を使う。

良いArchitectの周囲：
- チームメイトのDesignスコアが徐々に上がる（構造を学び、貢献し始める）
- 新規参加者の立ち上がりが速い（構造が明確なので理解しやすい）

Black Hole Engineerの周囲：
- チームメイトのDesignスコアが低いまま（構造に触れない・触れられない）
- 新規参加者の立ち上がりが遅い（一人に聞かないとわからない）

**重力の質は、周囲のスコアに映る。**

---

## 崩壊を防ぐのはリーダーの仕事

EISは崩壊を検出できる。しかし検出だけでは崩壊は防げない。

崩壊を防ぐのはリーダーの仕事だ。

具体的には：

- Bus Factor = 1を見つけたら、**意図的にペア作業やコードレビューの範囲を広げる**
- Indispensabilityの集中を見つけたら、**そのエンジニアに「教える」時間を作る**
- 一人だけArchitectが続くパターンを見つけたら、**設計判断を分散させる仕組みを作る**

EISが見せてくれるのは、宇宙の構造だ。その構造をどう整えるかは、人間の判断である。

---

## 崩壊からの再生——Black Holeを置き換えるエンジニア

崩壊は必ずしも終わりではない。

第5章でR.M.が新しい宇宙を作ったように、崩壊した宇宙に新しい重力を持ち込めるエンジニアが存在する。

このタイプのエンジニアには特徴がある。

- **Architectの再現性**を持つ（第8章）。どの宇宙でも構造を作れる
- 既存の重力場を**読める**。崩壊した構造を理解し、何が失われたかを把握できる
- 重力を**分散させる**設計を選ぶ。Black Holeの二の轍を踏まない

タイムラインで見ると、こういうパターンになる。

崩壊後のチームにこのエンジニアが入ると：
- チーム分類がUnstructured → Guardian → Balanced と回復する
- Bus Factorが1から2、3と上がる
- 複数メンバーのDesignスコアが同時に上がり始める

**崩壊を再生に変えられるのは、構造を分散させるArchitectだけだ。**

Black Holeの置き換えに必要なのは、同じ強さの重力ではない。**質の違う重力**だ。

---

## 恒星は永遠ではない。だから構造が要る。

宇宙では、恒星が死んでも、その恒星が作った元素は宇宙に残る。鉄も、酸素も、炭素も——すべて恒星の核融合が作ったものだ。

良いArchitectも同じだ。去った後に残るのは、コードではなく**構造**だ。

Black Hole Engineerが残すのは——虚空だ。

**恒星は永遠ではない。だから構造が要る。**

---

### シリーズ

- [第1章：履歴だけでエンジニアの「戦闘力」を定量化する](https://ma2k8.hateblo.jp/entry/2026/03/11/153212)
- [第2章：エンジニアの「戦闘力」から、チームの「構造力」へ](https://ma2k8.hateblo.jp/entry/2026/03/13/060851)
- [第3章：Architectには流派がある：Git履歴が暴く進化の分岐モデル](https://ma2k8.hateblo.jp/entry/2026/03/14/135648)
- [第4章：Backend Architectは収束する：成仏という聖なる仕事](https://ma2k8.hateblo.jp/entry/2026/03/14/155124)
- [第5章：タイムライン：スコアは嘘をつかないし、遠慮も映る](https://ma2k8.hateblo.jp/entry/2026/03/14/180329)
- [第6章：チームは進化する——タイムラインが暴く組織の法則](https://ma2k8.hateblo.jp/entry/2026/03/14/184223)
- [第7章：コードの宇宙を観測する](https://ma2k8.hateblo.jp/entry/2026/03/14/213413)
- [第8章：Engineering Relativity：なぜ同じエンジニアでもスコアが変わるのか](https://ma2k8.hateblo.jp/entry/2026/03/14/233602)
- [第9章：Origin：コード宇宙のビッグバン](https://ma2k8.hateblo.jp/entry/2026/03/15/054313)
- [第10章：Dark Matter：見えない重力](https://ma2k8.hateblo.jp/entry/2026/03/15/062608)
- [第11章：Entropy：宇宙は常に無秩序に向かう](https://ma2k8.hateblo.jp/entry/2026/03/15/062609)
- **第12章：Collapse：良いArchitectとBlack Hole Engineer**（本記事）
- [第13章：Cosmology of Code：コード宇宙論](https://ma2k8.hateblo.jp/entry/2026/03/15/062611)
- [第14章：Civilization：なぜ一部のコードベースだけが文明になるのか](https://ma2k8.hateblo.jp/entry/2026/03/15/062612)
- [第15章：AI Creates Stars, Not Gravity](https://ma2k8.hateblo.jp/entry/2026/03/15/062613)
- [第16章：The Engineers Who Shape Gravity：重力を作るエンジニアたち](https://ma2k8.hateblo.jp/entry/2026/03/15/062614) — **最終章**

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/logo-full.png?v=2)

**GitHub**: [engineering-impact-score](https://github.com/machuz/engineering-impact-score) — CLIツール、計算式、方法論すべてオープンソース。`brew tap machuz/tap && brew install eis` でインストール。

この記事が参考になったら：

[![Sponsor](https://img.shields.io/badge/Sponsor-❤-ea4aaa?logo=github&style=for-the-badge)](https://github.com/sponsors/machuz)

PayPay: `w_machu7`

---

← [第11章：Entropy](https://ma2k8.hateblo.jp/entry/2026/03/15/062609) | [第13章：Cosmology of Code →](https://ma2k8.hateblo.jp/entry/2026/03/15/062611)
