### 前回のあらすじ

[第1章](https://ma2k8.hateblo.jp/entry/2026/03/11/153212)では、gitの履歴データだけでエンジニア個人の「戦闘力」を7軸で定量化する手法を紹介した。3軸トポロジー（Role / Style / State）による分類も加え、「この人はどんなエンジニアか」を一言で表現できるようになった。

しかし、個人スコアだけでは見えないものがある。

**チーム**だ。

### 個人指標の限界

個人スコアが全員80点超えのチームが必ずしも「強い」とは限らない。全員がProducerタイプで、Architect不在。誰もアーキテクチャを触らない。誰も負債を片付けない。生産量は凄まじいが、コードは3ヶ月で腐る。

逆に、平均スコアが50点台でもArchitectが1人いて、Cleanerがいて、Growing状態の若手が育っているチームは強い。半年後にはもっと強くなる。

**「強いチーム」は個人の総和では測れない。構成と補完性が重要だ。**

### `eis team` — チーム分析コマンド

個人分析の `eis analyze` に加え、チームレベルの分析を行う `eis team` コマンドを追加した。

```bash
# ドメイン全体を1チームとして分析（設定不要）
eis team --recursive ~/workspace

# チーム定義ありで分析
eis team --config eis.yaml --recursive ~/workspace

# JSON出力（AIに貼って深い分析をしてもらう用）
eis team --format json --recursive ~/workspace
```

teams設定が無い場合はドメイン（Backend / Frontend / Infra）ごとに全メンバーを1チームとして扱う。設定なしでもすぐ使える。

```yaml
# eis.yaml
teams:
  backend-core:
    domain: Backend
    members:
      - alice
      - bob
      - charlie
  frontend-app:
    domain: Frontend
    members:
      - dave
      - eve
```

### チーム健全性 — 7つの指標

チームの「健全性」を以下の7軸で評価する。個人の7軸と対になる設計だ。

#### 1. Complementarity（補完性）

Roleの多様性をカバレッジとして測る。既知のRole 5種（Architect, Anchor, Cleaner, Producer, Specialist）のうち何種類がチームにいるか。

```
coverage = uniqueRoles / 5
bonus = Architect在籍(+10) + Anchor在籍(+5) + Cleaner在籍(+5)
score = coverage × 80 + bonus  (0-100)
```

**Architectが不在のチームは、補完性スコアで真っ先にわかる。** ArchitectはRoleの中で最もスコアを押し上げるボーナスを持つ。これは意図的な設計で、「設計を担える人」がチームにいるかどうかは補完性の核心だからだ。

#### 2. Growth Potential（成長力）

チーム内のGrowing状態メンバーの割合 + メンタリング環境の有無。

```
growingRatio = growingCount / memberCount
score = growingRatio × 60 + Builder在籍(+20) + Cleaner在籍(+20)
```

BuilderやCleanerは「手本になる人」がいる指標。**Growingの若手がいても、手本が不在なら育たない。** 両方揃って初めてスコアが上がる。

#### 3. Sustainability（持続性）

リスク状態（Former, Silent, Fragile）の逆数 + Architectの安定性。

```
riskRatio = (Former + Silent + Fragile) / memberCount
score = (1 - riskRatio) × 80 + Architect在籍(+20)
```

Former（元メンバーのコードが残っている）、Silent（コード書かないが居る）、Fragile（変更圧力のない場所でしか生き残っていない）——これらが多いチームは、見た目のメンバー数に反して実質的な戦力が少ない。

#### 4. Debt Balance（負債バランス）

メンバーのDebt Cleanupスコアの平均。50が中立で、50以上ならチーム全体がクリーン傾向。

```
score = avg(members.DebtCleanup)  // 0-100
```

50を大きく下回るチームは負債を生み出し続けている。50を超えるチームは自浄作用がある。

#### 5. Productivity Density（生産密度）

**この量のコードを、この人数で書いている**——という密度感。少人数で高いアウトプットを出しているチームにボーナスが付く。

```
base = avg(members.Production)
小チームボーナス: 3人以下で平均50以上 → ×1.2
                  5人以下で平均50以上 → ×1.1
```

3人で大規模APIサーバーを回しているようなチームは、この指標で異常値として可視化される。「すごいけど危険」——それが数字で見える。

#### 6. Quality Consistency（品質一貫性）

チーム全体の品質レベルと、そのバラつき。平均品質が高く、かつ標準偏差が小さいほどスコアが高い。

```
stdev = sqrt(sum((member.Quality - avgQuality)²) / memberCount)
score = avgQuality × 0.6 + (100 - stdev × 2) × 0.4
```

**全員が80点台のチームと、95点と40点が混在するチームは、平均が同じでも全く違う。** 後者はレビュー負荷が偏り、品質ゲートが形骸化している可能性がある。

#### 7. Risk Ratio（リスク人材割合）

Former + Silent + Fragile状態のメンバーが全体の何%を占めるか。直球の指標。

```
riskRatio = (Former + Silent + Fragile) / memberCount × 100  (%)
```

25%を超えたら要注意。50%超えは危機的。

### チーム5軸分類——コード→エンジニア→構造を逆算する（v0.10.0〜）

第1章でエンジニア個人の3軸トポロジー（Role / Style / State）を導入した。v0.10.0では、この個人トポロジーを**チームレベルに集約**して、チームの「型」を5つの軸で分類する。

**コードからエンジニアの特性を読み取り、エンジニアの分布からチームの構造を逆算している。** git logとgit blameという生データから出発して、個人→チーム→組織構造と、ボトムアップに全体像を組み上げていく。

#### 5つの軸

| 軸 | 何から導出するか | 問い |
|---|---|---|
| **Structure**（構造） | メンバーのRole分布 | チームにどんな構造的役割があるか |
| **Culture**（文化） | メンバーのStyle分布 | チームがどういうやり方で仕事をしているか |
| **Phase**（フェーズ） | メンバーのState分布 | チームが今どんなライフサイクルにあるか |
| **Risk**（リスク） | 健全性指標 | どんなリスクを抱えているか |
| **Character**（キャラクター） | 上4軸の複合 | 一言で言うとどんなチームか |

Characterは他の4軸から合成されるメタ分類。チームの「顔」を一言で表す。

#### 重み付き分類——強い人間ほどチームの色を塗る

分類で面白いのは、**メンバーの総合スコアを分類の重みに使う**点だ。

```
weight = member.Total / 100  (最低0.1)
```

総合スコア90点のArchitectと、15点のArchitectが同じチームにいても、前者の方が圧倒的にチームの「色」を決めている。民俗学的に言えば、**強いやつ、アウトプットが多いやつは、チームにより多くの文化を伝播させる**。それをそのまま数式にした。

最低重みを0.1にしているのは、「存在していること自体に意味がある」から。スコアが低くても、Growingメンバーが3人いればチームのPhaseに影響する。

#### Structure（構造）の分類

メンバーのRole分布から導出される、チームの構造的特徴。

| ラベル | 条件 | 意味 |
|---|---|---|
| **Architectural Engine** | Architect+Anchor大、AAR 0.3-0.8、カバレッジ高 | 設計と品質の両輪が回るチーム |
| **Architectural Team** | Architect多い | 設計力が厚い |
| **Architecture-Heavy** | Architectに偏重 | 設計はあるが実装が追いつかない |
| **Emerging Architecture** | Architect少数、Anchor/Producerが主 | 設計文化が芽生えつつある |
| **Delivery Team** | Producer主体 | 出荷重視 |
| **Maintenance Team** | Cleaner/Anchor主体 | 保守運用重視 |
| **Unstructured** | 「—」が大半 | 明確な構造がない |
| **Balanced** | 上記いずれにも該当しない | バランス型 |

**AAR（Architect-to-Anchor Ratio）** がStructure分類の鍵になる。Architectが多すぎても設計だけで実装が進まない。Anchorが多すぎても安定するだけで設計革新が起きない。AAR 0.3〜0.8が健全レンジ。

#### Culture（文化）の分類

メンバーのStyle分布から導出。

| ラベル | 主なStyle | 意味 |
|---|---|---|
| **Builder** | Builder多い | 作って残す文化 |
| **Stability** | Balanced/Resilient多い | 安定志向 |
| **Mass Production** | Mass多い | 量重視 |
| **Firefighting** | Churn/Rescue多い | 火消し文化 |
| **Exploration** | Spread多い | 探索型 |
| **Mixed** | 偏りなし | 混合 |

#### Phase（フェーズ）の分類

メンバーのState分布から導出。

| ラベル | 主なState | 意味 |
|---|---|---|
| **Emerging** | Growing多い | 成長期 |
| **Scaling** | Active多い + Growing存在 | 拡大期 |
| **Mature** | Active主体 | 成熟期 |
| **Stable** | Active + Balanced多い | 安定期 |
| **Declining** | Former/Silent多い | 衰退期 |
| **Rebuilding** | Active + Former混在 | 再構築期 |

#### Risk（リスク）の分類

健全性指標から導出。

| ラベル | 条件 | 意味 |
|---|---|---|
| **Design Vacuum** | Complementarity低い | 設計リーダー不在 |
| **Talent Drain** | Risk Ratio高い | 人材流出中 |
| **Debt Spiral** | Debt Balance低い | 負債が蓄積中 |
| **Quality Erosion** | Quality Consistency低い | 品質が崩壊中 |
| **Healthy** | 上記いずれもなし | 健全 |

#### Character（キャラクター）——チームの「顔」

Structure × Culture × Phase × Risk + 構造指標（AAR、Anchor Density、Productivity Density）から合成される、チームの総合的なキャラクター。

| キャラクター | 条件の概要 | 意味 |
|---|---|---|
| **Elite** | SC高い、AAR適正、PD高い | 設計力と生産性を兼ね備えた精鋭チーム |
| **Fortress** | Structure良好、Culture安定 | 堅牢で安定した守りのチーム |
| **Pioneer** | Phase成長期、Culture Builder | 新領域を切り拓く開拓チーム |
| **Academy** | Growing多い、Builder在籍 | 人材育成が活発なチーム |
| **Feature Factory** | Producer主体、Architect不在 | 機能を量産するが設計が弱い |
| **Guardian** | Anchor/Cleaner主体 | 保守と品質を守るチーム |
| **Firefighting** | Churn/Rescue文化 | 常に火消しに追われるチーム |

**SC（Structure-Culture complementarity）** はStructureとCultureがどれだけ噛み合っているかの指標。Architectural EngineのStructure + Builder Cultureは最高の組み合わせ。Delivery Team + Firefighting Cultureは最悪。

#### 構造指標

5軸分類に加え、チームの構造的な健全性を測る指標を3つ追加した。

**AAR（Architect-to-Anchor Ratio）**: Architectの数 ÷ Anchorの数。0.3〜0.8が健全レンジ。高すぎると設計過多（実装が追いつかない）、低すぎると安定過多（設計革新が起きない）。Architectがいるのにanchorが0だとAAR=∞でArchitect孤立を示す。

**Anchor Density**: アクティブメンバー中のAnchorの割合。品質と安定性の基盤がどれだけ厚いか。

**Architecture Coverage**: （Architect + Anchor）÷ チーム全員。設計と品質に関与するメンバーがチーム全体の何%を占めるか。

これらの構造指標は、チームの「骨格」を数値化する。Roleの分布だけではわからない、**構造の質**を見る窓になる。

### 強いチームの条件

チーム健全性の7軸と5軸分類を運用してみて見えてきたパターン：

**強いチームに共通する特徴：**
- Architect + Builder が在籍（設計する人と、設計を実装に落とせる人）
- Role多様性が3種以上（最低限 Architect / Anchor / Producer）
- Growing率が20%以上（若手が育っている）
- Risk Ratio が0%（リスク人材がいない、または少ない）
- Quality Consistencyが70以上（品質のバラつきが小さい）

**危険なチーム構成：**
- Mass/Churn偏重：大量に書くが生き残らないコードが溢れる
- Architect不在：誰も設計レイヤーを触らない → 暗黙知の蓄積
- Silent蓄積：形式上はメンバーだが実質的に貢献していない人が増える
- Producer一色：全員が書くだけで、片付ける人がいない

### 社会学的インサイト

この分析を複数チームで回して気づいたことがある。

**「こういう人がいるチームは人が育つ」** ——BuilderかCleanerが1人でもいると、Growing状態のメンバーが翌四半期にActiveに遷移する確率が体感で倍になる。コードレビューで「こう書くべき」を示す手本が存在するからだろう。

**「Architect不在のチームは品質が劣化する」** ——Complementarityが低いチームは、半年スパンで見るとQuality Consistencyも下がっていく。設計の道標がないと、各自が好き勝手にコードを書き始める。

**「少人数チームの異常値は、すごさとリスクの両面」** ——Productivity Densityが高いチームは確かに生産性が異常だが、1人抜けたときの崩壊リスクも高い。バス係数の個人版がIndispensabilityだとすれば、チーム版がProductivity Densityの裏側にある。

### 使い方

```bash
# インストール
brew tap machuz/tap && brew install eis

# チーム分析（最もシンプル）
eis team --recursive ~/workspace

# JSON出力 → AIに貼って深掘り
eis team --format json --recursive ~/workspace | pbcopy
# → Claude / ChatGPT に「このチームの強みとリスクを分析して」と投げる
```

深い洞察（Insights）の自動生成は意図的にスコープ外にしている。定量データの算出はツールの仕事、そこから何を読み取るかは人間（またはAI）の仕事。この住み分けが重要だと思っている。

### まとめ — 個人 × チームの両輪で組織を見る

第1章で作った個人スコアは「この人はどんなエンジニアか」を答える。
第2章で追加したチーム分析は「このチームはどんな状態か」を答える。

両方を組み合わせることで：
- **採用**: どのRoleが足りないかが可視化される → ポジション定義に使える
- **チーム編成**: 補完性を最大化する組み合わせが検討できる
- **1on1**: 個人スコアの推移をベースに成長の方向性を議論できる
- **リスク管理**: Risk Ratioの悪化を早期に検知できる

全部gitの履歴から出てくる。追加のツール導入も、エンジニアへのアンケートも不要。

**測れるものは改善できる。測れないものは祈るしかない。**

チームの強さを、祈りから指標に変えよう。
