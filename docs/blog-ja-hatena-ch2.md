# git考古学 #2 —— エンジニアの「戦闘力」から、チームの「構造力」へ

![Cover](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/hatena/cover-ch2.png)

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
# Simplest: domain = team
❯ eis team --recursive ~/workspace

# With explicit team definitions
❯ eis team --config eis.yaml --recursive ~/workspace

# JSON output
❯ eis team --format json --recursive ~/workspace
```

teams設定が無い場合はドメイン（Backend / Frontend / Infra）ごとに全メンバーを1チームとして扱う。設定なしでもすぐ使える。

```yaml
# eis.yaml (optional)
teams:
  backend-core:
    domain: Backend
    members: [alice, bob, charlie]
  frontend-app:
    domain: Frontend
    members: [dave, eve]
```

### チーム健全性 — 7つの指標

チームの「健全性」を以下の7軸で評価する。個人の7軸と対になる設計だ。

#### 1. Complementarity（補完性）

Roleの多様性をカバレッジとして測る。既知のRole 5種（Architect, Anchor, Cleaner, Producer, Specialist）のうち何種類がチームにいるか。

```
coverage = uniqueRoles / 5
bonus    = Architect(+10) + Anchor(+5) + Cleaner(+5)
score    = coverage × 80 + bonus  (clamped 0-100)
```

**Architectが不在のチームは、補完性スコアで真っ先にわかる。** ArchitectはRoleの中で最もスコアを押し上げるボーナスを持つ。これは意図的な設計で、「設計を担える人」がチームにいるかどうかは補完性の核心だからだ。

#### 2. Growth Potential（成長力）

チーム内のGrowing状態メンバーの割合 + メンタリング環境の有無。

```
score = growingRatio × 60 + Builder(+20) + Cleaner(+20)
```

BuilderやCleanerは「手本になる人」がいる指標。**Growingの若手がいても、手本が不在なら育たない。** 両方揃って初めてスコアが上がる。

#### 3. Sustainability（持続性）

リスク状態（Former, Silent, Fragile）の逆数 + Architectの安定性。

```
riskRatio = (Former + Silent + Fragile) / memberCount
score     = (1 - riskRatio) × 80 + Architect(+20)
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
bonus:
  ≤3 members && base ≥ 50 → ×1.2
  ≤5 members && base ≥ 50 → ×1.1
```

3人で大規模APIサーバーを回しているようなチームは、この指標で異常値として可視化される。「すごいけど危険」——それが数字で見える。

#### 6. Quality Consistency（品質一貫性）

チーム全体の品質レベルと、そのバラつき。平均品質が高く、かつ標準偏差が小さいほどスコアが高い。

```
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

![チーム5軸分類フロー：Code → Engineer → Team のボトムアップ構造推定](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/team-classification-flow.png)

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
weight = member.Total / 100  (minimum 0.1)
```

総合スコア90点のArchitectと、15点のArchitectが同じチームにいても、前者の方が圧倒的にチームの「色」を決めている。民俗学的に言えば、**強いやつ、アウトプットが多いやつは、チームにより多くの文化を伝播させる**。それをそのまま数式にした。

最低重みを0.1にしているのは、「存在していること自体に意味がある」から。スコアが低くても、Growingメンバーが3人いればチームのPhaseに影響する。

#### Structure（構造）の分類

メンバーのRole分布から導出される、チームの構造的特徴。

| ラベル | 条件 | 意味 |
|---|---|---|
| **Architectural Engine** | Architect+Anchor大、AAR 0.3-0.8、カバレッジ高 | 設計と品質の両輪が回るチーム |
| **Architectural Team** | Architect多い | 設計力が厚い |
| **Architecture-Heavy** | Architectに偏重（ただしArchitect/Builderは除外） | 設計はあるが実装が追いつかない |
| **Emerging Architecture** | Architect少数、Anchor/Producerが主 | 設計文化が芽生えつつある |
| **Delivery Team** | Producer主体 | 出荷重視 |
| **Maintenance Team** | Cleaner/Anchor主体 | 保守運用重視 |
| **Unstructured** | 「—」が大半 | 明確な構造がない |
| **Balanced** | 上記いずれにも該当しない | バランス型 |

**AAR（Architect-to-Anchor Ratio）** がStructure分類の鍵になる。Architectが多すぎても設計だけで実装が進まない。Anchorが多すぎても安定するだけで設計革新が起きない。AAR 0.3〜0.8が健全レンジ。

ただし例外がある。**Architect/Builder**（設計も実装もこなすタイプ）は、AARが高くても「設計過多で実装が追いつかない」問題を起こさない。全員がArchitect/Builderのチームであれば、Architecture-Heavyとは判定されない。むしろ、全員がArchitect/Builderのチームは最強の構成かもしれない。

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

**AAR（Architect-to-Anchor Ratio）**: Architectの数 ÷ Anchorの数。0.3〜0.8が健全レンジ。高すぎると設計過多（実装が追いつかない）、低すぎると安定過多（設計革新が起きない）。Architectがいるのにanchorが0だとAAR=∞でArchitect孤立を示す。ただし、ArchitectがBuilder Styleを兼ねている場合、設計と実装を1人でこなせるためAAR過多の警告は緩和される。

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

### 成長モデル — 3つの層を登る

EISのRole分類は、エンジニアの成長段階を3つの層として捉えることができる。

![成長モデル](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch2-diagram-growth-model.svg)

**実装層**: コードを書いて出す。Growingはここからスタートする。生産量は出るがSurvivalはまだ低い。

**安定化層**: 品質が上がり、コードが生き残り始める。他人のコードも直せるようになる。AnchorやCleanerはここにいる。

**設計層**: アーキテクチャファイルに手を入れ、構造を決める側に回る。Architectはここ。

成長とは、この層を登ること。EISのスコア推移で観測できる：

- Survival上昇 → 実装層から安定化層へ
- Design上昇 → 安定化層から設計層へ
- DebtCleanup上昇 → チーム貢献の幅が広がっている

チームの文脈で言えば、**Growth Potential（成長力）が高いチームはこの階段を登りやすい環境がある**。安定化層にAnchorがいて、設計層にArchitectがいる。手本が存在するから、Growingメンバーが次の層に進める。手本がなければ、実装層で回り続ける。

そしてもう一つ。この階段には**落ちる方向**もある。

![衰退モデル](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch2-diagram-decline.svg)

メンバーが層を登るのを支援し、Risk状態に落ちるのを早期に検知すること。**それがEISで見えるマネジメントの仕事だ。** スコアの推移を四半期ごとに追えば、誰が登っているか、誰が止まっているか、誰が落ちかけているかが数字でわかる。

### 社会学的インサイト

この分析を複数チームで回して気づいたことがある。

**「こういう人がいるチームは人が育つ」** ——BuilderかCleanerが1人でもいると、Growing状態のメンバーが翌四半期にActiveに遷移する確率が体感で倍になる。コードレビューで「こう書くべき」を示す手本が存在するからだろう。

**「Architect不在のチームは品質が劣化する」** ——Complementarityが低いチームは、半年スパンで見るとQuality Consistencyも下がっていく。設計の道標がないと、各自が好き勝手にコードを書き始める。

**「少人数チームの異常値は、すごさとリスクの両面」** ——Productivity Densityが高いチームは確かに生産性が異常だが、1人抜けたときの崩壊リスクも高い。バス係数の個人版がIndispensabilityだとすれば、チーム版がProductivity Densityの裏側にある。

### メンバー3層分類 & 自動Warnings（v0.10.3）

v0.10.3で、`eis team` に **3層メンバー分類** と **自動警告** を追加した。

#### Core / Risk / Peripheral

git上に名前が出る全員が「チームメンバー」とは限らない。ちょっと手伝っただけの人を含めると健全性指標が歪む。v0.10.3ではメンバーを3層に分ける：

| Tier | 条件 | 使われる場所 |
|---|---|---|
| **Core** | `最近アクティブ && Total >= 20` | 平均スコア、ProdDensity、QualityConsistency |
| **Risk** | Stateが Former/Silent/Fragile | 分布、RiskRatio、分類 |
| **Peripheral** | それ以外 | TotalMemberCountのみ |

ヘッダーは `4 core + 3 risk / 16 total` と表示される。ちょっと手伝った人はPeripheralとして除外され、Silent/Formerなメンバーはリスクとして検知される。

「みんな忙しいのに指標上はそうでもない → 中を覗いたらSilentがいた」というインサイトが自然に浮き上がる設計。

#### 自動 Warnings

危険な指標の組み合わせをプレーンテキストの警告として出力する：

![Team Warnings](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch2-warnings.svg)

警告の種類：

- **Bus factor risk**: 少人数のcoreが多くのリポを支えている
- **Risk ratio**: 無効化・リスク状態のメンバー比率
- **Top contributor concentration**: トップ貢献者が抜けた場合のProdDensity低下シミュレーション
- **Silent accumulation**: ヘッドカウントと実質貢献者数の乖離
- **Gravity warnings**: 脆い影響力集中、Architect在籍なのに構造カバレッジが低い

#### Phase 分類の精緻化

Phase軸に **Legacy-Heavy** と **Mature with Attrition** を追加。履歴が長いチームが一律「Declining（衰退）」と判定される問題を修正した：

| Label | 条件 | 意味 |
|---|---|---|
| **Legacy-Heavy** | Risk高いがAvgTotal≥40 + Architect在籍 | 強いが履歴が重いチーム |
| **Mature with Attrition** | 中程度のRisk（20-40%）、Active core健在 | 成熟チームからの自然減 |
| **Declining** | Risk高 + コアが弱い | 本当の衰退 |

Architectがスコア90超えで、Silent/Formerが何人かいるBackendチームは「衰退」ではない。**Legacy-Heavy**——強いが履歴が重い。この区別が次のアクションを変える。

### 実測結果 — うちのチーム

実際のプロダクト（Backend 12リポ + Frontend 9リポ）に対して `eis team` を実行した結果：

**Backend — Elite / Legacy-Heavy**:

- Core 4人で12リポを運用、Risk 3人（Silent 2 + Former 1）
- Architect + Anchor 2人 = AAR 0.50（健全レンジ）
- ProdDensity 60 ——4人としてはまずまずだが、トップ貢献者が生産の46%を占める
- Phase: `Legacy-Heavy` ——衰退ではないが、歴史の重みが載っている

**Frontend — Pioneer / Mature**:

- Core 6人、Risk 0人 —— 全員がアクティブ
- Architect + Anchor在籍、構造カバレッジ33%
- Sustain 100/100、RiskRatio 0% —— 健全そのもの
- Gravity警告：1人が高い構造影響力を持つが、robust survivalが低い

AIに診断させた結果をまとめると：

- **Backend**: 強いが履歴の重いElite。Characterは最上位だが脆い——1人の離脱がすべてを変える。
- **Frontend**: Mature（成熟）フェーズのPioneer。Architectが機能しており、Risk 0%。Gravity警告が1件残るが、チームとしては健全。

**数字が物語を持ち始めた。** 「誰が強いか」だけでなく「チームがどんな状態で、次に何が起きるか」が見えるようになった。

### 良い設計はコモンセンスを生む

BEチームがLegacy-Heavyと分類される理由は明確で、前任の馬力のあるアーキテクトが退任したが、作ってくれた量が膨大なため、彼しか触っていなかったモジュールがいくつか残っているからだ。git blameの大部分がFormerメンバーのものになる。

しかし実際には、チームは崩壊していない。

なぜか。それらのモジュールが十分に整理された設計のもとで作られていたからだ。口頭での引き継ぎは受けたが、完全なドキュメントや知識移転が行われたわけではない。それでも、構造としてコードに埋め込まれた設計が、後から読むエンジニアに一定の理解を与えている。

**強い設計は、人ではなく構造に知識を残す。** そして、その構造がチームの共通理解を作る。

これは「良い設計がコモンセンスを生む」という現象だと思う。優れた設計は、ドキュメントや知識の完全な移転を必ずしも必要としない。コードの構造自体が、そのモジュールの意図と使い方を伝える。

EISは現時点では履歴やコード生存率といった定量的な指標を扱っている。こうした「設計によるコモンセンス」——つまり、Formerメンバーのコードがなぜ今でも健全に動いているのか——は、まだ直接的には観測できていない。

しかし、もしそれが可能になれば、現在のような単純なLegacy-Heavy警告ではなく、**「履歴として重いだけの健全な構造」と「本当に危険な依存構造」**を区別できるようになるはずだ。

ただ、実際にはそこまで測りに行く必要はないかもしれない。強いチームであれば、Formerメンバーのコードを徐々に自分たちのコードに置き換えていき、Legacy-Heavyは時間とともに解消される。あるべき姿に収束していく。EISはその収束の過程を、Survivalの推移やRisk Ratioの変化として自然に捉えることができる。

### 使い方

```bash
# Install
❯ brew tap machuz/tap && brew install eis

# Team analysis
❯ eis team --recursive ~/workspace

# JSON → paste into AI
❯ eis team --format json --recursive ~/workspace | pbcopy
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

---

### シリーズ

- [第1章：履歴だけでエンジニアの「戦闘力」を定量化する](https://ma2k8.hateblo.jp/entry/2026/03/11/153212)
- **第2章：エンジニアの「戦闘力」から、チームの「構造力」へ**（本記事）
- [第3章：Architectには流派がある：Git履歴が暴く進化の分岐モデル](https://ma2k8.hateblo.jp/entry/2026/03/14/135648)
- [第4章：Backend Architectは収束する：成仏という聖なる仕事](https://ma2k8.hateblo.jp/entry/2026/03/14/155124)
- [第5章：タイムライン：スコアは嘘をつかないし、遠慮も映る](https://ma2k8.hateblo.jp/entry/2026/03/14/180329)
- [第6章：チームは進化する——タイムラインが暴く組織の法則](https://ma2k8.hateblo.jp/entry/2026/03/14/184223)
- [第7章：コードの宇宙を観測する](https://ma2k8.hateblo.jp/entry/2026/03/14/213413)
- [第8章：Engineering Relativity：なぜ同じエンジニアでもスコアが変わるのか](https://ma2k8.hateblo.jp/entry/2026/03/14/233602)
- [第9章：Origin：コード宇宙のビッグバン](https://ma2k8.hateblo.jp/entry/2026/03/15/054313)
- [第10章：Dark Matter：見えない重力](https://ma2k8.hateblo.jp/entry/2026/03/15/062608)
- [第11章：Entropy：宇宙は常に無秩序に向かう](https://ma2k8.hateblo.jp/entry/2026/03/15/062609)
- [第12章：Collapse：良いArchitectとBlack Hole Engineer](https://ma2k8.hateblo.jp/entry/2026/03/15/062610)
- [第13章：Cosmology of Code：コード宇宙論](https://ma2k8.hateblo.jp/entry/2026/03/15/062611)

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/logo-full.png?v=2)

**GitHub**: [engineering-impact-score](https://github.com/machuz/engineering-impact-score) — CLIツール、計算式、方法論すべてオープンソース。`brew tap machuz/tap && brew install eis` でインストール。

この記事が参考になったら：

[![Sponsor](https://img.shields.io/badge/Sponsor-❤-ea4aaa?logo=github&style=for-the-badge)](https://github.com/sponsors/machuz)

PayPay: `w_machu7`

---

← [第1章：履歴だけでエンジニアの「戦闘力」を定量化する](https://ma2k8.hateblo.jp/entry/2026/03/11/153212) | [第3章：Architectには流派がある →](https://ma2k8.hateblo.jp/entry/2026/03/14/135648)
