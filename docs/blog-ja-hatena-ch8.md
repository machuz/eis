# git考古学 #8 —— Engineering Relativity：なぜ同じエンジニアでもスコアが変わるのか

*同じ物体でも、月では軽く、木星では重い。コードベースでも同じことが起きる。*

### 前章までのあらすじ

[第7章](https://ma2k8.hateblo.jp/entry/2026/03/14/213413)ではコードベースの宇宙的構造——重力、四つの力、練度の高い良い重力——について語った。

今回は、その重力に関するもう一つの重要な性質について書く。

---

## 重力は宇宙によって変わる

EISを見ていて、もう一つ気づいたことがある。

**重力は宇宙によって変わる。**

EISはコードベースの中で「どれだけ重力を作ったか」を測っている。しかし重力という概念には、一つ重要な性質がある。

**重力は空間によって変わる。**

物理の世界では、地球、月、木星——それぞれ重力が違う。同じ物体でも、重くなったり軽くなったりする。

コードベースでも同じ現象が起きる。

**同じエンジニアでも、コードベースが変わるとEISが変わる。**

---

## 成熟した宇宙と若い宇宙

成熟したコードベースでは：

- 構造が安定している
- Architectが存在する
- 抽象化が確立している
- 「練度の高い良い重力」が既にある

このような環境では、新しい重力を作るのは簡単ではない。既存の構造が強固であればあるほど、その中心を動かすには大きなエネルギーが必要になる。**EISは上がりにくい。**

逆に、構造が弱いコードベースでは：

- 中心構造が存在しない
- 設計がばらばらである
- 抽象化が不足している

このような環境では、新しい重力が生まれやすい。最初にまともな設計を導入した人が、一気にArchitectになれる。**EISは上がりやすい。**

---

## EISは絶対値ではない

つまり、EISは**絶対値ではない**。

EISはエンジニアの能力だけで決まるのではなく、**コードベースの重力場との相互作用**で決まる。

これはある意味——

**Engineering Relativity**

である。

同じエンジニアでも、宇宙が変われば重力の見え方が変わる。

---

## 数字の罠

この視点はエンジニア評価にも重要な意味を持つ。

例えばあるエンジニアのスコアが：

![Repo Scores](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch8-repo-scores.svg)

普通は60の方が「優秀」に見える。

しかし **Repo Aが非常に強い重力場を持つコードベース** だった場合——Architectが複数存在し、構造が高度に洗練されたチームだった場合——**35というスコアはむしろ非常に高い**可能性がある。

ここでまず思いつくのは「正規化の罠」だ。EISの相対正規化では、チーム内の最高貢献者が100になる。つまり、あるリポジトリのトップスコアは、別のリポジトリでは平凡なスコアかもしれない。

しかし今回の話はそれよりもっと根本的だ。正規化の問題は計算の話である。Engineering Relativityは**構造の話**である。

**コードベースそのものが、スコアの「意味」を変えてしまう。**

これがEngineering Relativityだ。

---

## EISで相対性を読む方法

ではどうすれば、この相対性を考慮してEISを読めるのか。いくつかのヒントがある。

### 1. チーム分類を見る

`eis analyze --team` の結果を見る。

![Structure Comparison](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch8-structure-comparison.svg)

同じTotal: 40でも、Architectural Engineの中の40と、Unstructuredの中の40は意味が全く違う。

### 2. Architect密度を見る

チーム内のArchitectが多いほど、Design軸は上がりにくくなる。これは相対正規化の影響だ。Architectが3人いるチームでDesign: 60を出すことは、Architectが1人もいないチームでDesign: 100を出すことより、おそらく難しい。

### 3. `--per-repo` でクロスリポ分析する

![リポ別分析](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch8-bash-per-repo.svg)

`--per-repo` フラグを付けると、ドメイン全体の集約スコアに加えて、リポジトリごとの独立したスコアリング結果が表示される。あるリポではProducerだが、別のリポではArchitectになっている——そういうパターンは、そのエンジニアの適応力と潜在能力を示している。

### 4. タイムラインで「重力場の変化」を見る

![タイムラインコマンド](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch8-bash-timeline.svg)

コードベースの構造は固定ではない。チームメンバーの入退場、リファクタリング、新規機能の追加——これらによって重力場は変化する。タイムラインで見ると、「構造が弱まった瞬間にスコアが上がるエンジニア」と「構造の強さに関わらず安定してスコアを維持するエンジニア」が区別できる。

---

## Architectの再現性

ここでもう一つ面白い現象が観測される。

複数のコードベースを横断してEISを見ていると、あるタイプのエンジニアが存在することに気づく。**宇宙が変わっても重力を作るエンジニア**だ。

コードベースが変わっても、チームが変わっても、技術スタックが変わっても、再び**構造の中心**を作る。

これは **Architectの再現性** と言えるかもしれない。

`--recursive --per-repo` でワークスペース全体を分析したとき、複数のリポジトリで一貫してArchitectになっているエンジニアは、特定のコードベースに依存しない「汎用的な設計力」を持っている。逆に、一つのリポジトリでだけArchitectになっているエンジニアは、そのリポジトリの特殊な文脈で重力を作っている。これも価値があるが、性質が異なる。

EISの `--per-repo` 分析は、この再現性を**数値で確認**できる。

![Per-Repo Breakdown](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch8-per-repo-breakdown.svg)

---

## 重力レンズ効果：他人のスコアに自分の重力が映る

もう一つ、天体物理学から借りた面白い現象がある。

物理学では、巨大な天体を直接見なくても、その背後にある天体の光が曲がることで存在を検出できる。これが**重力レンズ効果**だ。

コードベースでも似たことが起きる。Architectの重力は、本人のスコアよりも**周囲のスコアをどう歪めているか**に最も如実に現れることがある。

強いArchitectが存在するとき：
- 他のエンジニアのSurvivalスコアが低くなる（Architectのコードがblameを支配する）
- チームのDesign軸分布が偏る（一人がアーキテクチャ変更の大部分を吸収する）
- 新規参加者のスコアに特徴的な「立ち上がりカーブ」が現れる——最初は低く、徐々に既存構造に貢献し始める

そのArchitectが去ったとき：
- 複数のエンジニアのスコアが同時に変動する
- Design Vacuumリスクが出現する
- スコア分布の「平坦化」が、重力の中心が失われたことを示す

これは `eis timeline --team` で観測できる。重力の中心が消えた瞬間、チーム全体のメトリクスが波紋のように揺れる。**重力は実在していた——その全貌を見るには、他者への影響を見る必要があっただけだ。**

---

## Great engineers create gravity in every universe.

本当に優秀なエンジニアは、どの宇宙でも重力を作る。

しかしその重力は、宇宙によって見え方が変わる。

**それがEngineering Relativityだ。**

---

### シリーズ

- [第1章：履歴だけでエンジニアの「戦闘力」を定量化する](https://ma2k8.hateblo.jp/entry/2026/03/11/153212)
- [第2章：エンジニアの「戦闘力」から、チームの「構造力」へ](https://ma2k8.hateblo.jp/entry/2026/03/13/060851)
- [第3章：Architectには流派がある：Git履歴が暴く進化の分岐モデル](https://ma2k8.hateblo.jp/entry/2026/03/14/135648)
- [第4章：Backend Architectは収束する：成仏という聖なる仕事](https://ma2k8.hateblo.jp/entry/2026/03/14/155124)
- [第5章：タイムライン：スコアは嘘をつかないし、遠慮も映る](https://ma2k8.hateblo.jp/entry/2026/03/14/180329)
- [第6章：チームは進化する——タイムラインが暴く組織の法則](https://ma2k8.hateblo.jp/entry/2026/03/14/184223)
- [第7章：コードの宇宙を観測する](https://ma2k8.hateblo.jp/entry/2026/03/14/213413)
- **第8章：Engineering Relativity：なぜ同じエンジニアでもスコアが変わるのか**（本記事）

**GitHub**: [engineering-impact-score](https://github.com/machuz/engineering-impact-score) — CLIツール、計算式、方法論すべてオープンソース。`brew tap machuz/tap && brew install eis` でインストール。

この記事が参考になったら：

[![Sponsor](https://img.shields.io/badge/Sponsor-❤-ea4aaa?logo=github&style=for-the-badge)](https://github.com/sponsors/machuz)

PayPay: `w_machu7`
