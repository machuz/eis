---
title: "Backend Architectは収束する：成仏という聖なる仕事"
---

![Cover](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/hatena/cover-ch4.png)

*退職したArchitectの魂を成仏させる——それは聖なる仕事だ。*

---

## Backendチームの構図

今回分析したBackendチームの指標はこうなっている。

![Backend Team](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch4-backend-team.svg)

そしてチーム指標。

![Team Classification](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch4-team-classification.svg)

**Architect 1人。Anchor 3人。Producer 0人。**

これはFEとは全く違う構造だ。

---

## Backend Architectはなぜ1人に集中するのか

FEではArchitectが複数生まれる可能性がある。

しかしBackendではそうならないことが多い。

理由は単純だ。

**Backendには構造の正解が存在しやすい。**

例えばBackendでは、次のような構造が比較的安定した解として存在する：

- Domain層
- Presentation層
- UseCase層
- Infrastructure層
- Transaction境界
- Event境界

このような設計パターンは、多くのシステムで再利用可能な構造を持っている。

つまり：

| FE | 構造の美学が複数ある |
|---|---|
| BE | 構造の正解が存在する |

この違いがある。

その結果、Backendでは**Architectが増殖するより、集中する**。

宇宙で例えるなら：

![Backend Structure](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch4-structure.svg)

という構造だ。

---

## 退職したArchitectの魂

ここでY.Y.に注目したい。

![Y.Y.のプロフィール](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch4-data-engineer-f.svg)

Y.Y.は**退職済みのArchitect**だ。

しかし時間減衰をかけてもなお、**チーム2位のImpactを叩き出している**。

これが何を意味するか。

**この人が作った資産が、今もコードベースに大量に残っている。**

---

### Former状態の意味

EISでは、**残っている資産が多い場合にのみ**Former状態が検出される。

Former状態の条件：

- Raw Survival（時間減衰なし）が高い
- Survival（時間減衰あり）が低い
- かつ、Design または Indispensability が高い

つまり**「かつてこの人は構造を作っていた。今は活動していない。でもコードは残っている」**という状態だ。

単に退職しただけではFormerにはならない。**残すべき資産を残した人だけ**がFormerになる。

Y.Y.は、まさにその条件を満たしている。

---

### 時間減衰を入れてもなお2位

ここがエモい。

EISは**時間減衰付きのSurvival**を使っている。2年前のコードは重み0.02まで下がる。

それでもY.Y.は2位だ。

つまり**退職後もなお、コードベースに大きな影響を持ち続けている**。

これは「強かった人が去った後の残像」ではない。

**今もコードベースを支えている現実の資産**だ。

---

## Formerを成仏させる

しかしFormer状態には、もう一つの意味がある。

**成仏させるべき対象**だ。

チーム指標を見ると：

![フェーズとリスク指標](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch4-data-phase-risk.svg)

Legacy-Heavyとは、**強いが履歴が重いチーム**という意味だ。

退職したメンバーのコードが多く残っている。それ自体は悪いことではない。良い設計だったからこそ残っている。

---

### 良い設計はコモンセンスを生む

ここで重要な事実がある。

**このチームは崩壊していない。**

Y.Y.しか触っていなかったモジュールがいくつか残っている。git blameの大部分がFormerメンバーのものだ。

しかし実際には、チームは普通に機能している。

なぜか。

**それらのモジュールが、十分に整理された設計のもとで作られていたからだ。**

口頭での引き継ぎは受けた。しかし完全なドキュメントや知識移転が行われたわけではない。それでも、**構造としてコードに埋め込まれた設計が、後から読むエンジニアに一定の理解を与えている**。

これは「良い設計がコモンセンスを生む」という現象だと思う。

優れた設計は、ドキュメントや知識の完全な移転を必ずしも必要としない。**コードの構造自体が、そのモジュールの意図と使い方を伝える。**

強い設計は、人ではなく構造に知識を残す。そして、その構造がチームの共通理解を作る。

---

### Legacy-Heavyの収束

EISは現時点では履歴やコード生存率といった定量的な指標を扱っている。「設計によるコモンセンス」——つまり、Formerメンバーのコードがなぜ今でも健全に動いているのか——は、まだ直接的には観測できていない。

もしそれが可能になれば、「履歴として重いだけの健全な構造」と「本当に危険な依存構造」を区別できるようになるはずだ。

ただ、実際にはそこまで測りに行く必要はないかもしれない。

**強いチームであれば、Formerメンバーのコードを徐々に自分たちのコードに置き換えていく。Legacy-Heavyは時間とともに解消される。あるべき姿に収束していく。**

EISはその収束の過程を、Survivalの推移やRisk Ratioの変化として自然に捉えることができる。

---

しかし時間が経つにつれて、**そのコードを理解できる人が減っていく**リスクもある。

だからFormerを成仏させる作業が必要になる。

---

### 魂はDebtに吸い込まれていく

Formerを成仏させるとは何か。

それは**Debt Cleanupに吸い込まれていく**ということだ。

![チーム平均Debt Cleanupスコア](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch4-data-debt-avg.svg)

現役メンバーがFormerのコードを触る。理解する。直す。書き換える。

するとFormerのblame行は減っていく。現役メンバーの行に置き換わる。

これが**成仏**だ。

EISではこのプロセスが可視化できる：

- FormerのSurvivalが徐々に下がる
- 現役メンバーのDebt Cleanupが上がる
- チームのLegacy-Heavy度が下がる

**尊い作業だ。**

退職したArchitectの魂が、現役メンバーの手によって、少しずつコードベースに溶け込んでいく。

---

## Anchorの本当の意味

このBackendチームには3人のAnchorがいる。

- Y.Y.（Former）
- P.
- R.S.

Anchorは単なる「品質守護者」ではない。

**Anchorとは、構造理解型エンジニア**である。

Anchorの特徴：

- 既存構造の理解が深い
- 構造を壊さない
- 品質を守る
- システムの整合性を維持する

つまりAnchorは**構造を理解し、構造の上で生産するエンジニア**だ。

そしてBackendでは、このAnchorが**Architectへの進化経路**になる。

---

## Backend Architectの進化

FEでは：

![FE進化経路：ProducerからEmergent Architect](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch3-diagram-fe-evolution.svg)

という進化が見られた。

しかしBackendでは、こうなることが多い：

![BE進化経路：ProducerからInheritance Architect](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch3-diagram-be-evolution.svg)

つまり**収束型進化モデル**である。

---

### 再現型Architect

Backend Architectの特徴は**構造を再現できること**にある。

例えばあるコードベースで成立した：

![設計パターン：Domain + Presentation + UseCase + Infrastructure](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch4-data-design-pattern.svg)

という設計があったとする。

優れたBackend Architectは、**その構造を別のシステムでも再現できる**。

つまりBackend Architectとは：

- 構造を生み出す人
- **かつ、構造を再現できる人**

FE Architectが「新しい構造を創発する」タイプだとすれば、Backend Architectは「**構造を純化し、再現する**」タイプだと言える。

---

## Producer Vacuum

もう一つ興味深い点がある。

このチームには**Producerが存在しない**。

![ロール分布：Architect 1、Anchor 3、Producer 0](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch4-data-role-count.svg)

Producerとは「構造を完全には理解していないが、構造の上で生産する人」である。

Producerがいない場合、チーム構造はこうなる：

![Producer Vacuumの図](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch4-diagram-producer-vacuum.svg)

これは**Producer Vacuum**と呼べる状態だ。

Producerが存在すると：

![三層構造：Architect、Anchor、Producer](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch4-diagram-three-layer.svg)

という三層構造が成立する。

この形がBackendチームの最も安定した形だ。

実際、自分のチームはArchitect-Builder × 1、Anchor × 2の体制だ。Producerは人間では埋めていない。その役目はAIが担っている。

構造を理解し、設計し、品質を守る——その部分は人間が持つ。構造の上で量を生産する部分は、AIに任せる。

三層構造の下層をAIで補完する形だ。

---

## Architect Bus Factor = 1

EISはこのチームに対して警告を出している。

![Bus Factor警告](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch4-data-bus-factor.svg)

つまり**Architect Bus Factor = 1**である。

Architectが抜けると、チームの生産密度が大きく下がる。

これはBackendチームにとって最も典型的なリスクだ。

FEでは創発型Architect候補が生まれることで、この問題が緩和される。

しかしBackendでは**構造の正解が存在するがゆえに、Architectが集中しやすい**。

結果として、Bus Factorリスクが高くなる。

---

## このチームの処方箋

このBackendチームに必要なのは何か。

1. **Producer層の補充**: Anchor → Producer の逆流を許容する。構造を完全に理解していなくても生産できる環境を作る
2. **Formerの成仏促進**: 退職したArchitectのコードを現役メンバーが引き継ぐ。Debt Cleanupを意識的に上げる
3. **Anchor → Architect のパス開通**: 構造理解型のAnchorが、設計判断に参加できるようにする

特に2番目は重要だ。

Legacy-Heavy状態は「悪い」のではない。**良い設計が残っている**証拠でもある。

しかしそれを理解し、引き継ぎ、必要なら書き換える作業は、**誰かがやらなければならない**。

その作業はDebt Cleanupとして可視化される。

---

## FEとBEの進化モデル比較

まとめると：

| | FE | BE |
|---|---|---|
| 構造 | 複数の美学が存在 | 正解が存在しやすい |
| Architect | 分散しうる | 集中しやすい |
| 進化モデル | 分岐型（創発 or 継承） | 収束型（継承が主流） |
| リスク | 構造の衝突 | Bus Factor集中 |

どちらが良いという話ではない。

**ドメインの性質が、進化モデルを規定する。**

EISはその違いを可視化し、それぞれに適した処方箋を示すことができる。

---

## 正直に言うと

自分の Impact が 92.4 で 1 位なのは、Architect として観測されたシグナルとしては強い。

ただ正直に言うと、**稼働時間が極端に長い**のも大きい。これは [心理OS の強さ](https://library.orbitlens.io/psychological-os/#ch2)——**回る状態の構造を最初にやり切ることへの執着**——が外に出ている形だ。最初に構造をやり切るのが結局一番効率的であること、組織構造的に今が最も伝播しやすいタイミングであること——内側ではそれが見えている。だから内圧で動き続ける。数字はその積算として立ち上がった。

そしてもう一つ嬉しいことがある。

**Y.Y.から十分良い設計を引き継ぎ、それをさらに良くできたからこそ、ArchitectとしてこのImpactに乗ることができている。**

Backendには太古から伝わる良い設計がある。先人たちが磨き上げてきた構造だ。自分はそれを学習し、積み上げ、Y.Y.もまた学習し、積み上げ、同じ認識を持てる状態で出会えた。だからコモンセンスが生まれ、引き継ぐことができた。

その上に、教科書通りではないが有用だと確信している `delegateProcess` と `partProcess` を載せた。既存の事業インパクトのある概念をさらに高める機能追加をした。新たな事業の核となる機能の概念整理と、良いモデリングをした。

その積み重ねが、今の数字になっている。

---

P.とは昔一緒に働いたことがある。その時に設計知識の合流をした。だからお互いの設計について共通知が取れている。今回のImpactがどうであれ、馬力があることは知っている。**うちのチームで80台を叩き出せる力があることも知っている。**

R.S. は本当に良い働きをしている。声をかけたとき、「経験が浅いので自信がない」と言っていた。でも、彼の [心理OS の強さ](https://library.orbitlens.io/psychological-os/#ch2)——熱の純度が高く、内圧で動ける状態——は見えていた。だから確信はあった。**そして今、Anchor として構造に残っている。** Spiral チームの Anchor だ。数字としても、誇れる形だ。

---

そしてもっと正直に言うと、**本当に見せたいのはチームの方**だ。

![Spiralチーム分類](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch4-data-elite.svg)

このチームがSpiral判定されること。Legacy-Heavyを抱えながらも機能していること。退職したArchitectの資産を成仏させながら前に進んでいること。

**EIS を作った動機の一つは、このチームの強さを恣意性の混じらない土俵で見せることだった。**

「うちのチーム強いんだよ」と言いたかった。でも「なんとなく強い」では説得力がない。

そして重要なのは、**恣意的なものが入り得ない土俵で、このチームが強いと言える状態にしたかった**ということだ。

誰かの主観で「あの人は強い」と言っても、それは評価者のバイアスかもしれない。政治がうまい人が高評価を受けているだけかもしれない。

しかしgit履歴は嘘をつかない。

コードが残っているか。設計に関与しているか。負債を生んでいるか、片付けているか。

**履歴からの逆算には、恣意性が入る余地がない。**

だからこの土俵で数字を出した。

もちろん、これが全てではない。

**それによって生まれたプロダクトが良いものか、マーケットに適合しているか——それは別の軸で判断する必要がある。**

EISはコードベース上での技術的影響度を測る。しかしプロダクトの価値は、コードの質だけでは決まらない。ユーザーに届いているか、ビジネスとして成立しているか、それは別の話だ。

EISは「技術組織としての強さ」を可視化する。その上で何を作るかは、また別の意思決定になる。

そしてその数字が、自分の直感と一致した。チームは本当に強かった。

---

## 仕様変更とRobust

チームメンバーから質問があった。

> 仕様変更などによるコードの上書きはRobustのシグナルには反映されないですか？

**反映されます。**

仕様変更でコードが書き換えられれば、元のコードを書いた人のRobustは下がる。

つまり**プランナーの精度もこのシグナルに影響してくる**。

仕様がブレればコードは書き換えられる。プランニングが甘ければ、エンジニアがどれだけ良いコードを書いても、それは消える。

このチームでRobustが出せているのは、エンジニアだけの成果ではない。**プランナーを含めた開発組織全体の成果**だ。

---

### リニューアルとImpact

もう一つ質問があった。

> 大幅リニューアルなどを行うとリニューアル以前のコードを書いていた人はリニューアルによってImpactが下がってしまうということになりますかね？

**そうです。**

リニューアルで以前のコードが置き換われば、以前のコードを書いていた人のImpactは下がる。

しかしここがポイントだ。

**そこに適応できるか。**

- リニューアル後の設計に自分の意見を織り込めるか
- リニューアル後の良い設計の上でもImpactが出せるか
- もっと言うと、良くても悪くても、**複数のコードベース上でImpactが出せるか**

それがエンジニアの真の強さだ。

一つのコードベースで高いImpactを出すのは、環境に恵まれれば可能だ。

しかし**どのコードベースでも、どのチームでも、適応して重力を出せる**——それが本物のエンジニアだと思う。

EISは一つのコードベース上でのインパクトを観測する。しかしその観測を複数のコードベースで繰り返せば、**環境を超えた再現性**が見えてくる。

---

## この発見の意味

第3章ではFEの分岐型進化モデルを示した。

第4章ではBEの収束型進化モデルを示した。

そしてもう一つ、**Formerの成仏**という概念が見えてきた。

退職したArchitectの魂は、コードベースに残る。

それを現役メンバーが引き継ぎ、理解し、書き換えていく。

**その尊い作業が、Debt Cleanupとして可視化される。**

冷たい数字こそが、最も熱いものを映す。

それがEISの本質なのかもしれない。

---

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/logo-full-zenn.png)

**GitHub**: [eis](https://github.com/machuz/eis) — CLIツール、計算式、方法論すべてオープンソース。`brew tap machuz/tap && brew install eis` でインストール。

