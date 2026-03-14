# git考古学 #5 —— タイムライン：スコアは嘘をつかないし、遠慮も映る

*3ヶ月ごとのスナップショットを並べたとき、数字はストーリーを語り始める。*

### 前章までのあらすじ

[第4章](https://ma2k8.hateblo.jp/entry/2026/03/14/155124)では、BackendチームのArchitect集中構造と、退職したArchitectの「成仏」について語った。

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

![Engineer F Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch5-engineer-f-timeline.svg)

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

![Engineer J Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch5-engineer-j-timeline.svg)

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

![Engineer I Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch5-engineer-i-timeline.svg)

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

### `--per-repo` が暴く「遠慮」の真の構造

タイムラインだけでは、Design 100 → 73 の変化は「設計関与が減った」としか読めない。

しかし `eis analyze --recursive --per-repo` でリポジトリ単位に分解すると、もっと精密な構造が見える。

Engineer Iのリポジトリ別コミット分布を見る。

![Per-Repo Commits](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch5-per-repo-commits.svg)

**Q3の「遠慮」の正体が見えてくる。**

Q3にEngineer Iは既存Repo Bに274コミット（過去最多）を叩き込んでいる。生産量は落ちていない——むしろ増えている。しかしそれは**既に確立された設計の上での生産**であり、アーキテクチャを動かす仕事ではなかった。

Design 100 → 73の原因はここにある。既存のリポジトリで大量に作っても、構造そのものを動かさなければDesign軸は上がらない。

---

### 「待ってくれ」という会話

実は、この「遠慮」の裏には会話があった。

Engineer Iにはずっと頭にある設計像があった。FEの構造をこうすべきだ、という確信。技術力も設計センスも、私はそれを信じていた。

しかしタイミングがあった。

**「エンジニア組織としての証明フェーズが終わって、事業側がエンジニアに全力ベットするタイミングが来たら、そのとき任せる。だから今は待ってくれ」**

Engineer Iの設計が正しいことは確信していた。その人がやりやすい構造にした方が全体最適の面でも良いことも。しかしスタートアップとして、まだ事業側がエンジニアリングに全力で投資する段階には至っていなかった。テックカンパニーにしたい意志はある。しかしその折り合いがつくまでは——少し待ってくれ、と。

Q3の「遠慮」は、衝突だけではなかった。**戦略的な我慢**でもあった。

---

### そして、新しい宇宙

```
Quarter      Repo A（既存）   Repo B（既存）   Repo C（新規）
2025-Q4             5               5            1,352          ← ここ
```

2025-Q4、新しいプロダクトが始まった。

Engineer Iに任せた。

既存リポへのコミットは5件まで激減し、代わりに新規リポに1,352コミット。**3ヶ月で1,352コミット**。翌四半期も1,333コミット。半年で2,685コミット。

この数字は異常だ。一人のエンジニアが半年でこの量を叩き出すのは、単純な生産力だけでは説明できない。

```
2025-Q3 (Jul)        72.4    73    97   100      73  Anchor       Balanced     ← 既存構造で生産
2025-Q4 (Oct)        81.7   100    68   100     100  Architect    Balanced     ← 新規構造を創造
```

Design 73 → 100。Anchor → Architect。

**グリーンフィールドで設計者の本領が爆発した。**

Engineer Iは特異なエンジニアだった。コードの設計だけでなく、UIデザインもできた。

当初、新プロダクトは既存のデザインを踏襲する方針でアウトプットが出ていた。しかしEngineer Iは1から作りたかった。

そして**2週間**で持ってきた。ダークテーマ、モバイル対応、美しいビジュアル、サイドペインによる優れた表現力の基盤——既存の延長線では到達できない完成度のプロトタイプだった。

面白いことが起きた。プロジェクトには外部の優秀なデザイナーが参画していた。そのデザイナーはEngineer Iのデザインを見て、自分の役割を再定義してくれた。ビジュアルデザインで張り合うのではなく、**熟練の情報設計スキルに絞って貢献する**方向に一歩引いた。

結果、エンジニアのデザイン力とデザイナーの情報設計力が噛み合い、新しいプロダクトは大成功を収めた。チームは新たなコードベースを設計の軸とすることを決めた。

---

### 「遠慮」の再解釈

`--per-repo`で見えた構造を踏まえると、「遠慮」の解釈がより立体的になる。

1. **表面**: チームとの衝突で設計関与を減らした
2. **構造**: 既存リポでの重い生産作業に集中していた（274コミット）
3. **文脈**: 新プロダクトを任せるタイミングを待っていた

そしてQ4で全てが繋がる。新しい宇宙が生まれ、Engineer Iはそこで重力を作った。

これは第8章で書く「Engineering Relativity」そのものだ——**同じエンジニアでも、宇宙が変われば重力の出方が変わる**。既存の強い重力場ではAnchorだったEngineer Iが、新しい宇宙では一瞬でArchitectになった。

能力は変わっていない。**宇宙が変わった。**

---

### そして復帰

```
2025-Q4 (Oct)        81.7   100    68   100     100  Architect    Balanced
2026-Q1 (Jan)        78.1   100    84    83     100  Anchor       Builder      Active
```

翌四半期もDesign 100を維持。web-adminの構造構築は続いている。

衝突を経て、戦略的に待ち、新しい宇宙で本領を発揮した。この「一度引いて、新しい場所で出る」パターンは、**Architectとしての成熟**を示している。

若いArchitectは衝突すると引くか、押し通すかの二択になりがちだ。しかし成熟したArchitectは、**一度引いてチームの反応を見て、改めて出る**ことができる。そして優れたリーダーは、**そのタイミングを見て場を用意する**。

Engineer Iのタイムラインと `--per-repo` は、その成熟過程を3ヶ月刻みで、リポジトリ単位で記録している。

---

## Transitions：変化の要約

`eis timeline` は変化を自動検出する。

![Transitions](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch5-transitions.svg)

RoleとStyleの変化が並ぶだけで、何が起きたかが見える。

Architect → Anchor → Architect → Producer。

**構造を作り → 安定させ → 再び作り → 完成して生産側に回る。**

Architectの仕事が終わったことがTransitionsだけで読み取れる。

---

## Engineer Fとmachuzの比較

タイムラインを並べると、もう一つ見えるものがある。

![Comparison](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch5-comparison-table.svg)

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

![Team Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch5-team-timeline.svg)

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
