# git考古学 #6 —— チームは進化する——タイムラインが暴く組織の法則

*個人が変われば、チームも変わる。チームタイムラインは、その変化に法則があることを教えてくれる。*

### 前章までのあらすじ

[第5章](https://ma2k8.hateblo.jp/entry/2026/03/14/xxxxxx)では `eis timeline` で個人のタイムラインを追い、Engineer Fの退場とmachuzのアーキテクチャ浸透、Engineer Iの「遠慮」と復帰を読み取った。

しかしあれは個人の話だった。**チームは個人の集合体だ。** 個人が変われば、チームの性質も変わる。

`eis timeline` はチームレベルの集計も自動出力する。今回はそのチームタイムラインを読み解き、さらにそこから見える**エンジニアの進化法則**を抽出する。

---

## チームタイムライン：何が見えるか

チームタイムラインは、各期間でチーム全体を分類する。

- **Character** — チームの性格（Elite, Guardian, Factory, Balanced, Explorer, Firefighting）
- **Structure** — 構造（Architectural Engine, Delivery Team, Maintenance Team, Unstructured）
- **Culture** — 文化（Builder, Stability, Exploration, Firefighting）
- **Phase** — フェーズ（Mature, Emerging, Declining）
- **Risk** — リスク（Healthy, Design Vacuum, Quality Drift）

加えて、Health指標（Complementarity, Growth Potential, Sustainability等）とScore Averagesが期間ごとに並ぶ。

**個人のRole/Styleの変化が、チームのCharacter/Structureの変化として表面化する。** これが見どころだ。

---

## 実データ：Backendチームの変貌

```
═══ Backend — Team Timeline ═══

Classification:
  Period            2024-H2         2026-H1
  Character         Balanced        Elite
  Structure         Unstructured    Architectural Engine
  Culture           Stability       Builder
  Phase             Declining       Mature
  Risk              Design Vacuum   Healthy
```

（2024-H2〜2025-H2はメンバー数が閾値未満のため分類なし。2024-H2と2026-H1を比較する。）

**Balanced → Elite。Unstructured → Architectural Engine。Declining → Mature。Design Vacuum → Healthy。**

すべての軸で改善している。

なぜか。第5章で見た通りだ。

- **2024年**: Engineer Fが Architect Builder として構造を支えていた
- **2025-H1**: Engineer Fが Anchor → Fragile へ。構造の担い手が不在に
- **2025-H2**: machuz が Architect Builder に到達
- **2026-H1**: チームが Elite / Architectural Engine / Mature / Healthy に

**個人の世代交代が、チームの性質変化として現れている。**

Engineer Fの退場は一時的にDesign Vacuumを生んだ。しかしmachuzがArchitectを継承し、新メンバーが参画したことで、チームはかつてより健全な状態に到達した。

Score Averagesを見ると：

```
Score Averages:
  Period            2024-H2         2026-H1
  Production        0.0             57.7
  Quality           0.0             64.6
  Survival          0.0             39.2
  Design            0.0             36.4
  Total             0.0             48.3
```

Design 36.4はまだ低い。Architectがmachuz一人に集中しているためだ。他のメンバーのDesignスコアはほとんど0〜30台。

**Eliteチームの次の課題は、Design力の分散だ。**

---

## 実データ：Frontendチームの変遷

Frontendはデータ期間が長く、変遷が読みやすい。

```
═══ Frontend — Team Timeline ═══

Classification:
  Period            2024-H2         2025-H1         2025-H2         2026-H1
  Character         Guardian        Factory         Guardian        Balanced
  Structure         Maintenance     Delivery        Maintenance     Maintenance
  Culture           Stability       Stability       Stability       Builder
  Phase             Declining       Declining       Declining       Mature
  Risk              Quality Drift   Quality Drift   Quality Drift   Design Vacuum
```

まず目を引くのは、**Declining → Mature の遷移が2026-H1でようやく起きている**こと。Backendより遅い。

そして**Culture: Stability → Builder** の変化。これは第5章で追ったEngineer I（参画当初からArchitect）の影響が大きい。Engineer Iが継続的に設計ファイルに関与し続けたことで、チーム文化がStability（守り）からBuilder（攻め）に変わった。

一方でRiskは **Quality Drift → Design Vacuum** に変化している。これは一見悪化に見えるが、意味が違う。

- **Quality Drift**: 品質にばらつきがある（Producerが多い状態）
- **Design Vacuum**: 設計者が不足している（Architectが離脱or不在）

Engineer Jが2025-H2にProducerとして安定し、Engineer Iも時期によってAnchorに振れる。**常にArchitectが2人いる状態ではない**ため、Design Vacuumリスクが出ている。

Frontendの面白い点はもう一つある。

```
Transitions:
  [2025-H1] Character: Guardian → Factory
  [2025-H1] Structure: Maintenance Team → Delivery Team
  [2025-H2] Character: Factory → Guardian
  [2025-H2] Structure: Delivery Team → Maintenance Team
```

**2025-H1にFactory / Delivery Teamに変わり、2025-H2に戻っている。**

この一時的な変化は何か。2025-H1はEngineer Iが Architect Builder として最高スコア（87.5）を叩き出した時期だ。同時期にEngineer Jも Anchor Emergent。つまり**設計者と生産者が同時に高稼働した**時期。結果、チームは一時的にFactory（大量生産型）/ Delivery Team（配信チーム）の特性を見せた。

しかしその「最大出力」は一時的だった。翌半期にはGuardian / Maintenanceに戻る。

**チームの性格は、個人の出力によって四半期単位で揺れる。** これはチームサイズが小さいほど顕著だ。

---

## ドメイン横断：InfraとFirmware

### Infra：Explorer / Emerging

```
Classification (2026-H1):
  Character         Explorer
  Structure         Balanced
  Culture           Exploration
  Phase             Emerging
  Risk              Design Vacuum
```

InfraはExplorer / Exploration / Emerging。**まだ形を成していないチーム**だ。

メンバーのスコアを見ると、全員がGrowingまたはSpread。Architectが一人もいない。Design Vacuum は必然だ。

しかしPhaseがEmergingであることは、**成長過程にある**ことを意味する。Decliningではない。

### Firmware：Firefighting

```
Classification (2026-H1):
  Character         Firefighting
  Structure         Maintenance Team
  Culture           Firefighting
  Phase             Declining
  Risk              Design Vacuum
```

Firmwareは2人しかいない。Character: Firefighting、Culture: Firefighting。

Production 100, Quality 84.6。**生産力はある。しかし設計がない。**

これは「今ある問題に対処し続けている」チームだ。構造を作る余裕がない。Architectが不在で、Design 0。

**小規模チームほど、一人の参入・離脱がチームの性質を根本から変える。** Firmwareにもし一人のArchitectが加わったら、Firefighting → Builder に変わるだろう。

---

## ここからが本題：進化モデル

個人のタイムラインとチームタイムラインを並べて見ると、**法則**が浮かび上がる。

誰がどう進化し、どの条件でRoleが変わり、チームにどう影響するのか。

タイムラインデータから抽出した進化モデルを、一つずつ紐解いていく。

---

### モデル1：Architect顕在化の経路は一つではない

machuzのBackendタイムライン：

```
2024-H1     27.6  Anchor  —        Growing
2024-H2     76.4  Anchor  Builder
2025-H1     58.4  Producer Balanced
2025-H2     92.5  Architect Builder
2026-H1     92.4  Architect Builder  Active
```

**Anchor → Producer → Architect。**

一見「成長の階段」に見えるが、実態は違う。machuzは他の現場で既にArchitectとしての経験を積んでいた。しかしこのチームには先代Architect（Engineer F）がいた。

machuzがやったのは、**先代の構造を尊重しながら、改善しつつ大量に機能を追加すること**だった。Anchor期は既存構造の理解、Producer期は既存構造の上での大量生産。その過程で、自分のアーキテクチャが徐々にコードベースに浸透していった。

Engineer Fのスコアが落ち始めた2025-H2、machuzのアーキテクチャが構造の主軸になった。EISはそれをArchitect Builderとして捉えている。

これは「成長してArchitectになった」のではなく、**「既に持っていたArchitectとしての設計思想が、コードベースに浸透した結果が数字に現れた」**パターンだ。

一方、Engineer IのFrontendタイムライン（第5章のデータ）：

```
2024-Q3     56.1  Anchor   Balanced
2024-Q4     75.7  Architect Balanced
```

**参画2四半期目で即Architect。** Engineer Iも外部でArchitectの経験を持っていた。しかしmachuzとは異なり、既存構造に合わせる期間を短くして、すぐに自分のアーキテクチャで設計を始めた。

**同じ「経験者の参画」でも、顕在化の速度が異なる。**

1. **浸透型** — 先代の構造を尊重し、生産しながら徐々に自分の設計を浸透させる（machuz型）
2. **即時型** — 短いAnchor期の後、すぐに自分のアーキテクチャで設計を開始する（Engineer I型）

浸透型は時間がかかるが、既存構造との連続性が保たれる。即時型は速いが、第5章で見たように**チームとの衝突リスクがある**。

---

### モデル2：Backend Architectは集中し、Frontend Architectは流動する

第5章でも触れたが、タイムラインでさらに明確になる。

**Backend:**

```
2024-H1   ponsaaan  93.5  Architect Builder    machuz  27.6  Anchor
2024-H2   ponsaaan  84.1  Architect Builder    machuz  76.4  Anchor Builder
2025-H1   ponsaaan  72.7  Anchor Balanced      machuz  58.4  Producer
2025-H2   ponsaaan  37.5  Anchor               machuz  92.5  Architect Builder
```

結果的に、**同時にArchitectが2人存在する期間は一つもなかった。** Engineer Fが降りた後にmachuzのアーキテクチャが主軸になっている。これがBEの単一設計軸（DB・APIの設計方針が一つ）に起因する構造的な制約なのか、単にアーキテクチャ浸透のタイミングの問題なのかは、このサンプルだけでは断定できない。しかし少なくとも観察事実として、BE Architectは集中する傾向がある。

**Frontend:**

```
2024-H2   Engineer I  72.7  Anchor       Engineer J  74.9  Architect Builder
2025-H1   Engineer I  83.8  Architect    Engineer J  54.3  Anchor
2025-H2   Engineer I  85.1  Architect    Engineer J  38.6  Anchor
```

FEでは、Engineer Iが Architect になった時点でEngineer Jは Anchor に降りている。**一見、BEと同じ「一人」パターンに見える。**

しかし3ヶ月刻みのデータ（第5章）では：

```
2025-Q2   Engineer I  73.2  Architect    Engineer J  63.8  Architect
```

**2025-Q2に、Architectが2人同時に存在していた。**

BEでは観察されなかった現象が、FEでは起きている。

なぜか。仮説として、**BEは「一つの構造」を共有する。** データベーススキーマ、APIの設計方針、共通ライブラリ。設計の判断軸が一つに収束しやすいため、Architectが集中しやすい。

**FEは「複数の構造」が並立できる。** コンポーネント設計、状態管理、ルーティング。それぞれの領域で独立した設計判断が可能だから、複数のArchitectが共存しやすい。

ただしBEの「一人」パターンがサンプルの限界によるものである可能性は残る。チーム規模が大きくなれば、BEでも複数Architectが共存するケースはあり得る。

チームタイムラインに出ている**FEのDesign Vacuumリスク**は、この流動性の裏返しだ。Architectが複数いても、同時にいるとは限らない。一人がProducerに移行した瞬間、Design Vacuumが生まれる。

---

### モデル3：Producerは代謝である

Engineer Jの遷移を改めて見る：

```
Architect → Anchor → Architect → Producer → Producer → Producer
```

**構造を作り切ったArchitectは、Producerになる。**

これは退化ではない。**代謝だ。**

Architectが構造を作る。構造ができあがる。するとArchitectの仕事は減る。設計ファイルへの変更は不要になり、Designスコアが下がり、Roleが自然とProducerに落ちる。

そしてProducerとして「構造の上で生産する」フェーズに入る。

machuzのBackendでも同じパターンが予測される。Architect Builderとして92を維持している今、いずれ構造が安定すれば、machuzもProducerに移行するだろう。**その時が来たら、次のArchitectが必要になる。**

ここでチームタイムラインのBackend Health指標が意味を持つ：

```
Growth Potential: 20.0
```

Growth Potential 20。**次世代Architectの芽がまだ弱い。** これはBackendチームの中期的なリスクだ。

---

### モデル4：創業Architectのライフサイクル

Frontendの6ヶ月タイムラインで、一人だけ異質な軌跡を持つエンジニアがいる。

```
--- Engineer K (Frontend) ---
2024-H1     87.8  Architect Builder
2024-H2     14.6  — —
2025-H1      7.1  — — Silent
2025-H2      3.2  — —
2026-H1      3.2  — —
```

**2024-H1にTotal 87.8、Architect Builder。** Production 81、Survival 100、Design 100。

翌半期にはTotal 14.6。以降、事実上ゼロ。

これは**創業Architectのライフサイクル**だ。

Engineer Kは、FEの初期構造を作った人物。2024-H1の時点ではコードベースの大半がEngineer Kのblameで埋まっていた。Architect Builderは当然の結果だ。

しかしチームが成長し、他のエンジニア（Engineer I、Engineer J）が参画して構造を書き換え始めると、Engineer KのSurvivalは急速に下がる。blameが他のメンバーに置き換わっていく。

**創業Architectは、チームが成長すればするほどスコアが下がる。**

これは失敗ではない。むしろ**成功の証**だ。自分一人で作った構造の上に、他のエンジニアが乗り、発展させていく。それが起きているからスコアが下がる。

```
Engineer K:  87.8 → 14.6 → 7.1 → 3.2 → 3.2
Engineer I:   — → 72.7 → 83.8 → 85.1 → 78.1
Engineer J:  25.9 → 74.9 → 54.3 → 38.6 → 54.2
```

**Engineer KのスコアがEngineer IとJに移転している。** スコアの合計が保存されているわけではないが、構造的な影響力が世代交代していることは明確だ。

Engineer Fの退場（第5章）とは異なる種類の「退場」だ。Engineer Fは**チーム離脱**による退場。Engineer Kは**チーム成長**による退場。

両方ともEISが捉えている。

---

### モデル5：BuilderはArchitectの前提条件

タイムラインのデータを見ると、**ArchitectになるにはBuilderを経由する**パターンが圧倒的に多い。

```
machuz:     Anchor → Anchor Builder → Producer Balanced → Architect Builder
Engineer I: Anchor Balanced → Architect Balanced → Architect Builder
Engineer J: Anchor Growing → Architect Balanced → Architect Builder
Engineer F: (初登場時点で) Architect Builder
```

machuzはAnchor Builderを経てArchitectに。Engineer IはArchitect Balanced → Architect Builderへ。Engineer JもArchitect Balanced → Architect Builderへ。

**Builderになれないエンジニアは、Architectにもなれない。**

Builderとは何か。新しいコードを書いて、それが残るエンジニアだ。既存コードの修正（Balanced）ではなく、新規の構造を追加する。

これは「設計する」行為の本質と一致する。設計とは「既存の修正」ではなく「新しい構造の創出」だから。

チームタイムラインのFrontend Culture が Stability → Builder に変わった瞬間、チームPhaseがDeclining → Matureに変わっているのは偶然ではない。

**Builder文化がないチームは成熟できない。** 守りだけでは衰退する。

---

### モデル6：Producer Vacuum

チームにProducerがいなくなると何が起きるか。

Backendの2024-H2を見る。machuzがAnchor Builder（76.4）、Engineer Fが Architect Builder（84.1）。しかしProducerが一人もいない。

この時期のチームClassification：

```
Character: Balanced
Phase: Declining
```

**Architectが構造を作り、Anchorが維持する。しかし誰もその構造の上で生産していない。**

構造だけあって生産がない。これがProducer Vacuumだ。

2026-H1のBackendを見ると：

```
Effective Members: 5
Total: 48.3（平均）
```

machuz（Architect Builder）に加え、複数のAnchor/Producerが存在する。構造の上で生産する人がいる。だからElite / Architectural Engine になれた。

**Architectだけでは、チームは機能しない。** Architectが作った構造を使って生産するProducerがいて、初めてチームは回る。

---

### モデル7：Producer期はアーキテクチャ浸透の仕込み

machuzのタイムラインをもう一度見る。

```
2024-H2     76.4  Anchor  Builder
2025-H1     58.4  Producer Balanced
2025-H2     92.5  Architect Builder
```

**2025-H1のProducer期を経て、Architect Builderとして数字に現れている。**

machuzのProducer期は「成長期間」ではない。**先代Architect（Engineer F）の構造の上で大量に機能を追加しながら、自分の設計思想を改善という形で少しずつ浸透させていた期間**だ。

既存の構造を使い倒すことで、構造の限界と可能性の両方を体得する。そして改善の中に自分のアーキテクチャを織り込んでいく。

Architect Builderとして92.5を叩き出した2025-H2は、その浸透が閾値を超えた瞬間だ。Engineer Fのスコア低下（構造の主軸が移動）と、machuzのDesign 100（設計変更が構造ファイルに及ぶ）が同時に起きている。

**Producer期は、自分のアーキテクチャをコードベースに浸透させるための仕込み期間になりうる。** 表面上はProducerだが、水面下では設計思想の移植が進行している。

この「仕込み」がない場合——既存構造への理解なしに設計を始めると、「机上の設計」になるリスクがある。構造を使う経験がないから、使いにくい構造を作ってしまう。

Engineer Iが参画直後にArchitectになれたのは、**前職で同種のアーキテクチャ経験を積んでいた**からだと推測できる。外部で構造を使い倒した経験がある。だから新しい環境でもすぐに「何を設計すべきか」がわかる。仕込み期間が不要だった。

---

## 法則の全体像

タイムラインから抽出した進化モデルをまとめる。

```
┌───────────────────────────────────────────────────┐
│                  進化モデル全体像                    │
├───────────────────────────────────────────────────┤
│                                                   │
│  [Growing] → [Anchor] → [Producer] → [Architect]  │
│                  ↑            │            │       │
│                  │            │     構造完成 │       │
│                  │            ←─────────────┘       │
│                  │         (代謝：再びProducerへ)     │
│                                                   │
│  ※ 浸透型:  先代の構造を尊重しつつ徐々に浸透         │
│             [Anchor] → [Producer] → [Architect]    │
│  ※ 即時型:  [Anchor(短期)] → [Architect] 直行       │
│  ※ 創業型:  [Architect] → スコア漸減（成功の証）     │
│                                                   │
├───────────────────────────────────────────────────┤
│  BE: Architect席 = 1（集中モデル）                   │
│  FE: Architect席 = 1〜2（流動モデル）                │
├───────────────────────────────────────────────────┤
│  Builder条件: Architect到達にはBuilder経験が必要      │
│  Producer燃料: 構造を使い倒す経験が設計力の源泉        │
│  Producer Vacuum: 生産者不在だと構造が活きない         │
└───────────────────────────────────────────────────┘
```

これらの法則は、うちのチームの実データから帰納的に導いたものだ。他のチームで同じ法則が成り立つかはわからない。

しかし `eis timeline` があれば、**自分のチームの法則を、自分で発見できる。**

---

## チームタイムラインの使い方

### 1. 組織レビュー

```bash
eis timeline --span 6m --periods 0 --recursive ~/workspace
```

半年ごとのチームタイムラインを出して、Phase / Risk の推移を見る。Declining が続いているなら、何かを変える必要がある。

### 2. Architect計画

チームにArchitectがいない（Design Vacuum）場合、誰がArchitectになれるかをタイムラインで判断する。

- Builder経験があるか？
- Producer期を経ているか？
- Anchorとして構造を理解しているか？

### 3. Producer Vacuumの検出

Architectがいるのにチームが機能していない場合、Producer Vacuumを疑う。構造はあるのに生産がない状態。メンバーの個人タイムラインを見て、Producerが不在ならリソース配分を見直す。

### 4. 創業Architectの正しい評価

スコアが下がっている創業メンバーを見つけたら、**それが失敗なのか成功なのかを判断する**。チームの成長に伴うスコア低下なら、むしろ評価すべきだ。

---

## この発見の意味

第5章で時間軸を手に入れた。第6章で**法則を手に入れた**。

スナップショットは「今」を映す。タイムラインは「変化」を映す。そして法則は「次に何が起きるか」を予測する。

- machuzがProducerに移行したら、次のArchitectが必要
- FEのDesign Vacuumが続くなら、Engineer IのArchitect復帰を待つか新しいArchitectを育てる
- InfraがEmerging → Mature に進むには、まずBuilderが必要

**冷たい数字から法則を導き、法則から未来を読む。** それがタイムラインの本当の力だ。

---

**GitHub:** [machuz/engineering-impact-score](https://github.com/machuz/engineering-impact-score) — CLIツール、計算式、方法論を公開しています。`brew tap machuz/tap && brew install eis` ですぐ使えます。

この記事が参考になったら：

[![Sponsor](https://img.shields.io/badge/Sponsor-❤-ea4aaa?logo=github&style=for-the-badge)](https://github.com/sponsors/machuz)

PayPay: `w_machu7`
