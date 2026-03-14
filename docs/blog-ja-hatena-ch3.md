# git考古学 #3 —— Architectには流派がある：Git履歴が暴く進化の分岐モデル

### 前章までのあらすじ

[第1章](https://ma2k8.hateblo.jp/entry/2026/03/11/153212)では、gitの履歴データだけでエンジニア個人の「戦闘力」を7軸で定量化する手法を紹介した。[第2章](https://ma2k8.hateblo.jp/entry/2026/03/13/060851)では、個人スコアをチームレベルに集約し、チームの健全性と5軸分類を導入した。

しかし、スコアを眺めていてもう一つ、面白いことに気づいた。

**Architectには進化の流派がある。**

### スコアの肌感

まずスコアの体感値を共有しておく。

EISを様々なエンジニアに適用して見えてきたこと：

- **10〜20**: 業務委託や「この人は厳しかった」と判断した人
- **20台前半**: 「微妙だな」と感じていた人
- **30超え**: 「この人と働きたい」と判断したメンバー
- **40超え**: シニア相当

感覚と数字が一致する。

**40でシニア扱い**——これがこの指標の厳しさだ。7軸すべてで高水準を出すのは構造的に難しい。ものづくりのシニアは、本当に厳しい。

### あるチームの風景

あるFEチームのEISを見ていて、Architectの進化経路が見えてきた。

---

**Engineer A** — Architect / Builder / Active

現行構造を成立させているチームの中心。

![Engineer Archetypes](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch3-engineer-profiles.svg)

このチームの重力の中心だ。

---

**Engineer B** — Anchor / Mass / Active

死ぬほど生産している。しかしRobustは11。

大部分がEngineer Aに直されている。

しかし**11残るまで生産し続けている**。

Anchorは品質を守る側を示す。Massは精度が悪いことを示す。今のEngineer Bの**苦しみと頑張りが数字に出ている**。

これは単に「未熟」ではない。Anchorとして品質を守ろうとしている。ただしスタイルがまだ粗い。

つまり**継承型Architectの前段**にいる。

---

**Engineer C** — Producer / Balanced / Active

このチームで一番、みんなの意見を取り入れて、どう作るかを気にしてコード化している。

だからRobustなコードが生み出せている。

**Producerは、設計系の判断はしていないが、コードをちゃんと生み出している人を示す。** 一定コードをちゃんと生み出していない人にはラベルはつかないようにしている。肩を並べて一緒に働いている証左がProducerだ。

設計判断や、重力のあるコード（みんなが使い回すようなもの）はまだ生み出せていない。しかし適応力の高いProducerとして、チームを前進させている。

---

**Engineer D** — Producer / Emergent / Active（高Gravity）

Architect以外で唯一、**強い重力を持っている**。

エンジニアで重力を出せる人は本当に貴重だ。

しかしDormantが示す通り、**まだ誰も触っていないところが多い**。Engineer Dが先行して生み出したコードに、まだみんなで触れていない。

そしてRobustが低い。

**これがEngineer Dの苦しみを一番示している指標だ。** 触られた重力場で、置き換えられている割合が多いことを示している。

でも、全く残ってないわけじゃない。ちゃんと残っている。でもそれが、感覚じゃ感じられない。

**この指標は、それをちゃんと拾う。**

実はこの記事を書くにあたって、Engineer Dの状態を分析していて気づいた。

「高Gravity + 生産している + でもRobust低い」——この組み合わせは、既存のStyleでは捉えきれていなかった。Balancedと判定されていたが、それは違う。**これは創発型Architect候補の状態そのものだ。**

そこでv0.10.20で**Style: Emergent**を追加した。

Emergentとは「まだ形になりきっていないが、生まれつつある」という意味だ。

- 新しい構造を提案して、既存の構造とぶつかっている最中
- まだチームに揉まれきっていない（Robust低い）が、重力は生み出している
- 時間が経てばArchitectに進化する前段階

**この記事の洞察がそのまま、指標の進化になった。**

---

**Engineer E** — Producer / Churn / —

書けば書くほど、負債を生み出すタイプだ。

コードが悉く置き換えられている。

**Style Churnはそれを指している。** 生産しているが、生存していない。これもまた、この指標が炙り出す現実だ。

---

### 数字が語る物語

ここまで見てきて思うことがある。

**この指標は、日々の積み上げから来る物語を可視化している。**

Engineer Bの「死ぬほど生産して、でも11%しか残らない。それでも生産し続けている」——これは苦しみだ。でも同時に、頑張りでもある。その両方が、数字に出ている。

Engineer Dの「重力を出せる。でもまだ置き換えられている。でも全く残ってないわけじゃない」——これも苦しみだ。でもちゃんと残っている。感覚では感じられないそれを、指標は拾う。

**冷たい数字が、実は一番エモい。**

コードベースに刻まれた日々の積み上げ。苦しみ。頑張り。少しずつ残っていくもの。それを可視化できる指標が、ここにある。

### 重力の質

ここで一つ、重要な補足がある。

**Gravityには2種類の質がある。**

EISのGravityは以下で計算される：

![Gravity計算式](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch1-formula-gravity.svg)

Engineer Dの場合：

![Engineer DのGravity計算](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch3-gravity-calc-d.svg)

Gravity 68の**大部分はIndispensability（1人占有）とBreadth（広さ）**から来ている。Design = 5 なので「構造の中心」というより「広く1人で持っている」。

一方、Engineer Aの場合：

![Engineer AのGravity計算](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch3-gravity-calc-a.svg)

Gravity 84は**Design 100が大きく寄与**している。構造の中心を作っている。

---

つまり重力には2種類ある。

**真の重力（Structural Gravity）**

Designベースの重力。構造の中心を作っている。他のコードがそこを起点に広がる。

**占有重力（Occupancy Gravity）**

Indispensabilityベースの重力。広く1人で持っている。他に触れる人がいない。

---

「重力を出せるやつは貴重」「どのコードベースでも重力を出す」——という意味では、**この2種類は質が異なる**。

Engineer Aは真の重力を持っている。Engineer Dは占有重力を持っている。

しかしこれは、Engineer Dの否定ではない。

**占有重力は、真の重力の前段階でもある。**

まず広く持つ。まだ誰も触っていない。衝突と置き換えが起きる。そしてDesignが上がっていく。

つまり **占有重力 → 真の重力** という進化経路がある。

創発型Architectは、まさにこの道を歩む。

### 変更圧を受けていないコードの危うさ

もう一つ、EISが可視化する重要な状態がある。

**State: Fragile**だ。

コードが「生き残っている」と言っても、その意味は2つある：

- **Robust Survival**: 他の人が頻繁に変更しているファイルの中で、書き換えられずに残っている
- **Dormant Survival**: 誰も触らないモジュールにコードが残っている

後者は耐久性ではない。**単に放置されているだけ**だ。

Fragile状態は以下の条件で検出される：

- Dormant比率が80%以上（ほとんど誰も触っていない）
- Indispensabilityが高い（一人で持っている）
- Productionが低い（あまり生産していない）

つまり**「コードが残っているが、それは誰も触っていないから」**という状態だ。

変更圧が高まった瞬間、このコードは崩壊する可能性がある。品質が高いから残っているのではなく、**テストされていないだけ**だからだ。

Survivalが高いと安心感を与えるが、Dormant比率と合わせて見ないと本質を見誤る。EISはこの「休眠コード」と「変更圧に耐えたコード」を分離して可視化できる。

### チーム指標が示すもの

このチームのチーム指標を見ると、一つのWarningが出ている。

![データ警告](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch3-data-warning.svg)

**Engineer Dの苦しみが、Warningとして出ている。** 高い重力を持っているが、Robust survivalが低い。

しかしこのチームは強い。

![チーム指標](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch3-data-team-metrics.svg)

なぜ強いのか。

それは**中心構造と、次の構造案が、同居しているから**だ。

### なぜこの3つのRoleが存在するのか

ここで一歩引いて、生態系の視点で考えてみる。

コードベースという生態系には、3つの役割が必要だ。

- **Architect** = 構造を**作る**。地形を形成し、他の生物が住める環境を創る
- **Anchor** = 構造を**維持する**。土壌を安定させ、生態系の崩壊を防ぐ
- **Producer** = 構造を**拡張する**。既存の環境の上で繁殖し、ユーザー価値を生む

どれか一つが欠けても生態系は成り立たない。Architectだけでは構造は作れても機能が生まれない。Producerだけでは機能は増えても構造が崩れる。Anchorがいなければ、構造も機能も時間とともに腐っていく。

**健全なコードベースには、この3つが共存している。**

### 2つの進化経路

ここから本題に入る。

**Architectには流派がある。**

---

**継承型Architect**

Anchorから進化するタイプだ。

特徴：

- 既存構造の理解が深い
- 現場の制約をよく知っている
- 破壊より整流が得意
- 既存システムを崩さず強くする

これは特に**BEで強い**。

なぜならBEは、明確な責務分離、ドメイン境界、正解に近い設計が比較的存在しやすいからだ。

つまり**良い構造をコピーし、守り、純化する**ことに価値がある。

---

**創発型Architect**

High-Gravity Producerから進化するタイプだ。

特徴：

- 既存構造に乗るより、新しい構造を作る
- 最初は摩擦が大きい
- 置き換えられたり衝突したりする
- でも新しい中心を作る

これは特に**FEで面白い**。

なぜならFEはBEほど**唯一の正解がない**からだ。

### FEの美しさ

ここはかなり重要だ。

FEは BE と違って、UX、画面構成、state管理、interaction、abstractionの粒度——に**複数の美学**が入りやすい。

なのでFEのArchitectは**正解を継承する人**だけでは弱い。

むしろ、**異なる構造案を持ち込み、既存の構造とぶつかり、より強い重力場を作る人**が必要になる。

**FEにおいては、主張をぶつけ合う構造が美しい。**

### Engineer Dの読み方

さっきの

![Engineer Dのパターン](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch3-data-pattern.svg)

は単なる未熟さではない。

**既存構造と競合する、新しい構造案を、先に出している**とも読める。

つまり「Architect候補」というより、もっと正確には**創発型Architect候補**だ。

### Engineer Bの読み方

Engineer Bは

![Engineer Bのパターン](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch3-data-anchor-mass.svg)

である。

これは単に「Architectになりにくい」ではなく、**継承型Architectの素地がある。ただしまだスタイルが粗い**と読むのが正しい。

- 構造を守る側にいる
- 品質意識もある
- ただし量産・衝突がまだ多い

つまり**継承型Architectの前段**だ。

### 苦しみの先にあるもの

この指標は、今のコードベース上での頑張り、影響度を示すものだ。

ここから設計が置き換わり、どう変わっていくかは生態系によって変わる。

しかしこの重力、このスコアが出せる人は、どんなところでも、一旦スコアが低くなって苦しい時間があっても、**最終的には適正な知性ポテンシャルのもと、適正なスコアに落ち着く**。

このチームでドメイン知識、ものづくりの姿勢において、**60〜80のスコアを叩き出して、Engineer Aを脅かす存在になる人がいるとしたら——それはEngineer Dしかいない。**

### キャリアモデルは一本線じゃない

つまりこういうことだ。

**BEっぽい進化**

![BE型進化パス](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch3-diagram-be-evolution.svg)

**FEっぽい進化**

![FE型進化パス](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch3-diagram-fe-evolution.svg)

キャリアモデルが一本線じゃなくなる。**分岐型進化モデル**になる。

そしてこの2タイプは**対立ではなく補完関係**だ。

理想のチームはたぶん：

- 継承型Architect
- 創発型Architect候補
- Anchor
- Producer

つまり**守る力、壊して作る力、安定化する力、実装する力**が同居する。

### Producerしかいないチームの危うさ

逆に、危険なチーム構成がある。

**全員がProducerのチーム。**

一見すると活発に見える。コードは大量に生産されている。PRも毎日マージされている。

しかし：

- **誰も設計レイヤーを触らない** → 暗黙知が蓄積する
- **誰も負債を片付けない** → コードは3ヶ月で腐る
- **誰も品質を守らない** → fix率が上がり続ける

全員が書くだけで、片付ける人がいない。全員が前に進むだけで、土台を固める人がいない。

このタイプのチームは、EISで見ると一発でわかる。

![Producer-Only Warning](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch3-producer-warning.svg)

Architect不在。Anchor不在。Cleaner不在。

**Complementarity（補完性）スコアが壊滅的に低い。**

生産量は出ているのに、Survivalが全員低い。Qualityもバラバラ。半年後には技術負債の山が残る。

このチームに必要なのは、もう1人Producerを追加することではない。**Architectか、せめてAnchorを1人入れること**だ。

### このチームの構図

この補正を入れると、かなり綺麗に見える。

- **Engineer A**: 現行構造を成立させている Architect
- **Engineer B**: 継承型Architect側に伸びうる Anchor
- **Engineer C**: 適応力の高い Producer
- **Engineer D**: 創発型Architect候補の High-Gravity Producer

この構図は、単なる「Architect 1人 + 追従者」ではない。

**中心構造と、次の構造案が、同居している。**

これは強い。

### この発見の意味

この発見は、EISのキャリアモデルを一段上げる。

一本線じゃなくて**分岐型進化モデル**になる。

そしてそれは、チーム設計論にもなる。

---

**GitHub:** [machuz/engineering-impact-score](https://github.com/machuz/engineering-impact-score) — CLIツール、計算式、方法論を公開しています。`brew tap machuz/tap && brew install eis` ですぐ使えます。

この記事が参考になったら：

[![Sponsor](https://img.shields.io/badge/Sponsor-❤-ea4aaa?logo=github&style=for-the-badge)](https://github.com/sponsors/machuz)

PayPay: `w_machu7`
