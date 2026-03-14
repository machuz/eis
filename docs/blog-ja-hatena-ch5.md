# git考古学 #5 —— タイムライン：スコアは嘘をつかないし、遠慮も映る

*3ヶ月ごとのスナップショットを並べたとき、数字はストーリーを語り始める。*

### 前章までのあらすじ

[第4章](https://ma2k8.hateblo.jp/entry/2026/03/14/xxxxxx)では、BackendチームのArchitect集中構造と、退職したArchitectの「成仏」について語った。

しかしあの分析には限界がある。**ある一時点のスナップショットでしかない**。

エンジニアは変化する。成長もするし、遠慮もする。チームとの関係が変われば、コードへの関わり方も変わる。

**その変化を見るには、時系列が必要だ。**

---

## `eis timeline` — 時間軸を手に入れる

EISに `timeline` コマンドを追加した。

```bash
# デフォルト：直近1年を3ヶ月刻み
eis timeline --recursive ~/workspace

# 2024年から3ヶ月ごと
eis timeline --span 3m --since 2024-01-01 --recursive ~/workspace

# 半年ごと全期間
eis timeline --span 6m --periods 0 --recursive ~/workspace

# 特定メンバーだけ
eis timeline --author alice,bob --recursive ~/workspace

# JSON出力（AIに分析させる用）
eis timeline --format json --recursive ~/workspace
```

仕組みはシンプルだ。

1. 全コミットを1回だけ収集
2. 期間境界（3ヶ月ごと）でコミットをスライス
3. 各期間で `git blame <boundary-commit> -- <file>` を実行して当時のblame状態を再現
4. 各期間でスコアリングパイプラインを実行

**「collect once, slice many」戦略**。コミット収集は1回。blameだけ期間ごとに走る。

これで、各メンバーのスコア・Role・Style・Stateが3ヶ月ごとに並ぶ。変化が見える。

---

## 実データで見る：FEチームのタイムライン

うちのFEチームの2024年Q3以降のタイムラインを並べてみる。

主要メンバー3人に注目する。

---

### Engineer Fの軌跡：Architectは退場しても語る

```
--- Engineer F (Backend) ---
Period              Total  Prod  Qual  Surv  Design  Role         Style
2024-Q1 (Jan)        90.0   100    69   100     100  Architect    Builder
2024-Q2 (Apr)        94.4   100    71   100      87  Architect    Builder
2024-Q3 (Jul)        72.5    59    72   100      71  Producer     Balanced
2024-Q4 (Oct)        90.6   100    77   100     100  Architect    Builder
2025-Q1 (Jan)        79.2   100    82   100      28  Anchor       Balanced
2025-Q2 (Apr)        68.4    36    84   100      58  Anchor       Balanced
2025-Q3 (Jul)        49.1    81    77    51       4  Anchor       Balanced
2025-Q4 (Oct)        31.2    18    78    23       8  —            Balanced     Fragile
2026-Q1 (Jan)        11.3     0     0    34       0  —            —            Former
```

**2024年前半のEngineer Fは、machuz並みの数値を叩き出していた。**

Total 90超え。Architect Builder。Production 100、Design 100、Survival 100。

これは「強い」とかいうレベルではない。**コードベースの設計者そのもの**だ。

2024-Q3で一瞬Producerに落ちているが、Q4で即座にArchitect Builderに復帰している。この揺れは「設計に関与しない期間があった」だけであり、翌四半期で巻き返せるだけの構造理解がある証拠だ。

2025年に入ると徐々にスコアが下がり始める。Architect → Anchor → Fragile → Former。

**これは退職の軌跡だ。**

しかし注目すべきは、2025-Q2の時点でもSurvival 100を維持していること。コードが残っている。設計が生きている。

第4章で書いた「成仏」の対象は、まさにこの人だ。そしてこのタイムラインを見れば、**成仏させるべき資産がどれほど大きいか**が一目でわかる。

---

### Engineer Jの軌跡：Architect Builderだった

```
--- Engineer J (Frontend) ---
Period              Total  Prod  Qual  Surv  Design  Role         Style
2024-Q1 (Jan)        28.1    26    73    33       2  Anchor       —            Growing
2024-Q2 (Apr)        15.5     8   100    16       0  —            —            Growing
2024-Q3 (Jul)        61.9    52    72    38     100  Architect    Balanced
2024-Q4 (Oct)        91.7   100    74    96     100  Architect    Builder
2025-Q1 (Jan)        63.9    90    85    15      61  Anchor       Emergent
2025-Q2 (Apr)        63.8    48    73    76      81  Architect    Balanced
2025-Q3 (Jul)        44.7    62    70    18      18  Producer     Emergent
2025-Q4 (Oct)        39.4    62    60    50       0  Producer     Balanced     Former
2026-Q1 (Jan)        54.2    43    61   100       1  Producer     Balanced     Active
```

**2024-Q4のEngineer J、Total 91.7。Architect Builder。**

この数字は驚異的だ。同時期のmachuz（Backend）が64.1であることを考えると、**この四半期のEngineer Jはチーム全体で最も高い構造的影響力を持っていた**。

Design 100。Production 100。Survival 96。

つまりこの四半期のFEの構造は、**Engineer Jが作った**と言っていい。

その後のRoleの遷移が面白い。

```
Architect → Anchor → Architect → Producer → Producer → Producer
```

Architect Builderとして構造を作り切った後、Anchorに移行し、再びArchitectに戻り、最終的にProducerに落ち着いている。

これは「Architectとしての仕事が一段落した」ことを意味する。構造を作り終えたから、今度は構造の上で生産する側に回った。

**健全な遷移だ。**

---

### Engineer Iの軌跡：参画初日からArchitect

```
--- Engineer I (Frontend) ---
Period              Total  Prod  Qual  Surv  Design  Role         Style
2024-Q3 (Jul)        56.1   100    97    60       2  Anchor       Balanced
2024-Q4 (Oct)        75.7    59    84   100      78  Architect    Balanced
2025-Q1 (Jan)        87.5   100    93   100     100  Architect    Builder
2025-Q2 (Apr)        73.2    67    91   100     100  Architect    Builder
2025-Q3 (Jul)        72.4    73    97   100      73  Anchor       Balanced
2025-Q4 (Oct)        81.7   100    68   100     100  Architect    Balanced
2026-Q1 (Jan)        78.1   100    84    83     100  Anchor       Builder      Active
```

**参画2四半期目でArchitect。その後ずっとArchitect圏内。**

2024-Q3に参画し、初四半期はAnchor。しかし翌四半期でArchitectに昇格している。

これが「ずっとArchitectの動きをしていた」の正体だ。

Total 75.7 → 87.5 → 73.2 → 72.4 → 81.7 → 78.1。**コンスタントに70超え**。

Design 100を複数の四半期で叩き出している。これはアーキテクチャファイルへの変更が継続的に行われていることを意味する。

そしてここに、**興味深い揺れ**がある。

---

### 2025-Q3の「遠慮」

```
2025-Q2 (Apr)        73.2    67    91   100     100  Architect    Builder
2025-Q3 (Jul)        72.4    73    97   100      73  Anchor       Balanced     ← ここ
2025-Q4 (Oct)        81.7   100    68   100     100  Architect    Balanced
```

2025-Q3でArchitect → Anchorに落ちている。StyleもBuilder → Balanced。

Totalはほぼ変わらない（73.2 → 72.4）。Productionは上がっている（67 → 73）。Qualityも上がっている（91 → 97）。

**能力は落ちていない。設計関与が減っただけだ。**

Design 100 → 73。これが「Anchorに落ちた」原因。

この四半期に何があったか。

**チームと衝突した。**

具体的には、FEのアーキテクチャ方針について意見の相違があった。

Engineer Iは参画直後からArchitectとして設計に関与してきた。その設計方針がチームの既存メンバーと噛み合わない局面があった。

結果、**設計判断への関与を意図的に減らした**。

EISはそれを正確に捉えている。

- Design: 100 → 73（設計ファイルへのコミット減少）
- Style: Builder → Balanced（構造を作る側から、バランスよく既存構造に合わせる側へ）
- Role: Architect → Anchor（設計者から構造維持者へ）

**数字は遠慮を映す。**

---

### そして復帰

```
2025-Q4 (Oct)        81.7   100    68   100     100  Architect    Balanced
2026-Q1 (Jan)        78.1   100    84    83     100  Anchor       Builder      Active
```

翌四半期、Design 100に復帰。Total 81.7。Architect。

衝突を経て、チームとの距離感を掴んだ上で、再び設計に関与し始めた。

この「一度引いて、また出る」パターンは、**Architectとしての成熟**を示している。

若いArchitectは衝突すると引くか、押し通すかの二択になりがちだ。しかし成熟したArchitectは、**一度引いてチームの反応を見て、改めて出る**ことができる。

Engineer Iのタイムラインは、その成熟過程を3ヶ月刻みで記録している。

---

## Transitions：変化の要約

`eis timeline` は変化を自動検出する。

```
Notable transitions:
  • Engineer I: Role Anchor→Architect (2024-Q4 (Oct))
  • Engineer I: Style Balanced→Builder (2025-Q1 (Jan))
  • Engineer I: Role Architect→Anchor (2025-Q3 (Jul))     ← 衝突
  • Engineer I: Style Builder→Balanced (2025-Q3 (Jul))     ← 遠慮
  • Engineer I: Role Anchor→Architect (2025-Q4 (Oct))      ← 復帰
  • Engineer I: Role Architect→Anchor (2026-Q1 (Jan))
  • Engineer I: Style Balanced→Builder (2026-Q1 (Jan))
```

RoleとStyleの変化が並ぶだけで、何が起きたかが見える。

Engineer JのTransitionsも面白い。

```
  • Engineer J: Style Balanced→Builder (2024-Q4 (Oct))      ← 構造構築期
  • Engineer J: Role Architect→Anchor (2025-Q1 (Jan))       ← 構造安定化
  • Engineer J: Role Anchor→Architect (2025-Q2 (Apr))       ← 再び設計
  • Engineer J: Role Architect→Producer (2025-Q3 (Jul))     ← 構造完成
  • Engineer J: State Former→Active (2026-Q1 (Jan))         ← 復帰
```

Architect → Anchor → Architect → Producer。

**構造を作り → 安定させ → 再び作り → 完成して生産側に回る。**

Architectの仕事が終わったことがTransitionsだけで読み取れる。

---

## Engineer Fとmachuzの比較

タイムラインを並べると、もう一つ見えるものがある。

```
            Engineer F (BE)                    machuz (BE)
2024-Q1     90.0  Architect Builder           —
2024-Q2     94.4  Architect Builder          31.5  Anchor Balanced
2024-Q3     72.5  Producer Balanced          73.8  Anchor Builder
2024-Q4     90.6  Architect Builder          64.1  Anchor Builder
2025-Q1     79.2  Anchor Balanced            61.7  Anchor Builder
2025-Q2     68.4  Anchor Balanced            49.2  Anchor Balanced
2025-Q3     49.1  Anchor Balanced            93.2  Architect Builder
2025-Q4     31.2  — Fragile                  87.7  Architect Builder
2026-Q1     11.3  — Former                   92.4  Architect Builder
```

**Engineer Fが退場するタイミングで、machuzのアーキテクチャが構造の主軸になっている。**

2025-Q3。Engineer Fが49.1まで落ちた四半期で、machuzが93.2を叩き出しArchitect Builderに。

これは偶然ではない。

第4章で書いた「Backend Architectは集中する」という構造がここに現れている。結果的に、**BEで同時にArchitectが2人存在する期間は一つもなかった**。Engineer Fが先にArchitect Builderとして構造を支えていた時期、machuzはまだAnchorだった。machuzがArchitectに到達したのは、Engineer Fのスコアが落ちた後だ。

これがBEの単一設計軸（DB・API方針）に起因するArchitect集中の構造的帰結なのか、単にアーキテクチャ浸透のタイミングの問題なのかは、このサンプルだけでは断定できない。しかし少なくともこのチームでは、**世代交代として起きた**ことは確かだ。

タイムラインはこの世代交代を可視化する。

---

## タイムラインが語ること

一時点のスナップショットでは見えなかったものが、タイムラインでは見える。

| 一時点のスナップショット | タイムライン |
|---|---|
| 「今強い」 | 「いつから強くなったか」 |
| 「Architectだ」 | 「いつArchitectになったか」 |
| 「遠慮している」とは読めない | 「一時的に設計関与が減った」が見える |
| 退職 = データ消失 | 退職の軌跡が残る |
| チーム構造 = 静的 | チーム構造 = 動的（世代交代が見える） |

**数字は嘘をつかない。そして、遠慮も映す。**

Engineer Iの2025-Q3の「一歩引いた」動きは、おそらく本人も意識的にやっていたことだ。しかしそれが3ヶ月の数字として残り、前後の四半期と並べたときに初めて「あ、あの時か」とわかる。

Engineer Fの退場とmachuzのアーキテクチャ浸透も、タイムラインなしでは「今この構造になっている」としか言えない。しかしタイムラインがあれば「この世代交代は2025-Q3に起きた」と特定できる。

---

## 使い方のヒント

タイムラインの実用的な使い方をいくつか。

### 1. 1on1の材料にする

```bash
eis timeline --author alice --recursive ~/workspace
```

メンバー個別のタイムラインを出して、1on1の冒頭に見せる。「この四半期、Designが下がってるね。何かあった？」

数字は攻撃のためではない。**対話のきっかけ**として使う。

### 2. 採用判断の振り返り

新メンバーの参画後3〜6ヶ月でタイムラインを見る。GrowingからActiveへの遷移が見えれば成功。半年経ってもRole/Styleが空欄なら、オンボーディングに問題がある。

### 3. 退職の予兆検出

Active → Fragile → Former の遷移パターンを見れば、退職の軌跡がわかる。**逆に言えば、Active → Fragile の段階で手を打てる可能性がある。**

ただしEngineer Fのケースは通常の退職パターンではない。本人の意思で辞めたのではなく、**会社間の投資関係の解消**によってチームから離脱せざるを得なかった。つまり個人の不満やモチベーション低下ではなく、外部要因によるFragile→Formerだ。

それでもEISはこの変化を正確に捉えている。理由が何であれ、**コードベースへの関与が減れば数字に出る**。Fragileが出た時点で「なぜか」を確認する——それが自発的な退職準備なのか、外部要因なのかは、数字だけではわからない。しかし**変化が起きていること自体は検出できる**。

むしろ特殊なケースでも機能していることが重要だ。通常の退職予兆であれば、もっと早い段階で介入できる可能性がある。

### 4. チームタイムラインで組織の変遷を追う

`eis timeline` はチームレベルのタイムラインも自動出力する。

```
═══ Backend / Backend — Team Timeline ═══

Classification:
  Period            2024-Q4         2025-Q4         2026-Q1
  Character         Guardian        Balanced        Elite
  Structure         Maintenance     Unstructured    Architectural Engine
  Phase             Declining       Declining       Mature
  Risk              Quality Drift   Design Vacuum   Healthy
```

Guardian → Balanced → Elite。Declining → Mature。Design Vacuum → Healthy。

**チームが健全化していく過程が見える。**

---

## この発見の意味

第1章でスナップショットを作った。第2章でチームを見た。第3章でArchitectの流派を見た。第4章で成仏を見た。

第5章で**時間軸を手に入れた**。

スナップショットは「今」を映す。タイムラインは「なぜ今こうなっているか」を映す。

Engineer Fが作った構造の上で、machuzのアーキテクチャが浸透しArchitectとして数字に現れた。Engineer Jが構造を作り切って、Producerに落ち着いた。Engineer Iが一度引いて、また出た。

**全部、数字に残っていた。**

冷たい数字が、最もエモいストーリーを語る。それがタイムラインの本質だ。

---

**GitHub:** [machuz/engineering-impact-score](https://github.com/machuz/engineering-impact-score) — CLIツール、計算式、方法論を公開しています。`brew tap machuz/tap && brew install eis` ですぐ使えます。

この記事が参考になったら：

[![Sponsor](https://img.shields.io/badge/Sponsor-❤-ea4aaa?logo=github&style=for-the-badge)](https://github.com/sponsors/machuz)

PayPay: `w_machu7`
