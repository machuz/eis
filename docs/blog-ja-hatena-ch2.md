# git考古学 #2 —— 個人のシグナルだけでは足りない：Git履歴からチームの健全性を観測する

![Cover](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/hatena/cover-ch2.png)

*個人のシグナルは「誰のシグナルが強いか」を教えてくれる。チームの健全性は「来四半期もこのチームが強いかどうか」を教えてくれる。*

![Team structure and health radar](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ch2-iconic.png?v=4)

## 個人のシグナルだけでは足りない

全員がImpact 80超えのチームが必ずしも「強い」とは限らない。全員がProducerなら、誰もアーキテクチャを触らない。誰も負債を片付けない。出荷は速いが、コードはもっと速く腐っていく。

逆に、平均50点台でも——Architectが1人、Cleanerが1人、Growingの若手が2人——というチームのほうが、はるかに健全な状態にあるかもしれない。

**強いチームは個人のシグナルの総和ではない。構成と補完性こそが重要だ。**

---

## なぜ売上ではエンジニアリングの健全性がわからないのか

「売上が伸びているから開発もうまくいっている」——危険な仮定だ。売上が測っているのは**プロダクト・マーケット・フィット**であって、**エンジニアリングの健全性**ではない。

売上は車のスピード。エンジニアリングの健全性はエンジンの状態。エンジンが壊れかけていてもスピードは出る——下り坂であれば。

Git履歴には、売上では見えないシグナルがある：

- **コードの耐久性** —— 同じ機能を毎四半期書き直していないか？
- **技術的負債** —— 1機能追加すると2件のバグ修正が生まれていないか？
- **バス係数** —— 1人が抜けたら死ぬモジュールがいくつあるか？

**売上が成長していても、Survival低下 + 負債増加 + バス係数の集中が同時進行していれば、組織はスケール時に崩壊する。**

---

## チーム健全性の7軸

`eis team` は個人のシグナルをチームレベルの健全性に集約する：

```bash
❯ eis team --recursive ~/workspace
```

| 軸 | 何を測るか | キーインサイト |
|---|---|---|
| **Complementarity** | Role多様性（Architect, Anchor, Cleaner, Producer, Specialist） | Producer一色のチームは16点。全Role揃えば100点 |
| **Growth Potential** | Growingメンバー + Builder/Cleanerの手本がいるか | 手本がなければ若手は育たない |
| **Sustainability** | リスク状態（Former, Silent, Fragile）の逆数 | チーム速度の隠れた足枷 |
| **Debt Balance** | Debt Cleanupの平均。50超えなら自浄作用あり | チームが負債を生むか、片付けるか |
| **Productivity Density** | 1人あたりの生産量 + 少人数ボーナス | 「この人数でこのアウトプット」 |
| **Quality Consistency** | 平均品質 + 低分散 | 平均80でも40〜95に散らばるチームは健全ではない |
| **Risk Ratio** | Former/Silent/Fragileの割合 | 25%超えで警告。50%超えは危機 |

> 各軸の計算式は [Whitepaper](https://github.com/machuz/eis) を参照

---

## チーム分類 —— 銀河形態学

EISはチームを**5つの軸**で分類する。個人のトポロジーからボトムアップで導出される：

![チーム分類フロー](https://raw.githubusercontent.com/machuz/eis/main/docs/images/team-classification-flow.png?v=4)

| 軸 | 何から導出するか | 問い |
|---|---|---|
| **Structure** | Role分布 | チームにどんな構造的役割があるか？ |
| **Culture** | Style分布 | チームはどう仕事をしているか？ |
| **Phase** | State分布 | ライフサイクルのどこにいるか？ |
| **Risk** | 健全性指標 | どんなリスクを抱えているか？ |
| **Character** | 上4軸の複合 | このチームはどんな銀河か？ |

Characterは**銀河形態学**を使う——望遠鏡は銀河の形を記述するのであって、善し悪しを評価するのではないから：

| Character | 銀河 | 意味 |
|---|---|---|
| **Spiral** | 渦巻銀河 | 強い中心核 + 活発な星形成。設計と生産の両方が駆動 |
| **Elliptical** | 楕円銀河 | 成熟、安定、変化に強い。低エントロピー |
| **Starburst** | スターバースト銀河 | 爆発的成長。エネルギーは高いが、構造はまだ形成途中 |
| **Nebula** | 星雲 | 次世代エンジニアが育っている |
| **Irregular** | 不規則銀河 | 重力中心がない。高出力だが方向性なし |
| **Dwarf** | 矮小銀河 | 小さいが長寿。安定した品質 |
| **Collision** | 衝突銀河 | 構造的混乱。常に火消しに追われている |

> 天文学的な解説を含むガイド: [Galaxy Morphology Guide](https://orbit-d8x.pages.dev/galaxy-guide.html)

分類は**Impactで重み付け**される——Impact 90のArchitectは15のArchitectよりもはるかにチームの色を決める。強いシグナルを持つメンバーほど多くの文化を伝播させる。

---

## 成長モデル

EISのRole分類は3つの層にマッピングされる：

![成長モデル](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ch2-diagram-growth-model.png?v=4)

**実装層** → **安定化層** → **設計層**

- Survival上昇 → 実装層から安定化層へ登っている
- Design上昇 → 安定化層から設計層へ登っている
- DebtCleanup上昇 → チーム貢献の幅が広がっている

Growth Potentialが高いチームには、この階段を登れる環境がある——各層に手本がいる。手本がなければ、Growingメンバーは実装層で回り続ける。

![衰退モデル](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ch2-diagram-decline.png?v=4)

**BuilderやCleanerがいるチームは人の成長が速い。** 手本が存在すると、GrowingメンバーがActiveに遷移する速度がおよそ2倍になる。**Architectがいないチームは、時間とともに劣化する。**

---

## メンバーTier

gitに名前が出る全員が「チームメンバー」とは限らない。EISはメンバーを3層に分ける：

| Tier | 条件 | 使われる場所 |
|---|---|---|
| **Core** | `最近アクティブ && Impact >= 20` | 平均値、Density、Consistency |
| **Risk** | Former / Silent / Fragile | RiskRatio、分類 |
| **Peripheral** | それ以外 | カウントのみ |

ヘッダーは `4 core + 3 risk / 16 total` と表示される。ちょっと手伝った人は指標を希釈しない。Silentメンバーは検知される。

EISは**自動警告**も表示する——バス係数リスク、Silent蓄積、Gravity脆弱性、トップ貢献者集中。

![Team Warnings](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ch2-warnings.png?v=4)

---

## 実測結果

自社プロダクト（Backend 12リポ + Frontend 9リポ）に対して `eis team` を実行した結果：

**Backend — Spiral / Legacy-Heavy**：

- Core 4人で12リポを運用、Risk 3人（Silent 2 + Former 1）
- Architect + Anchor 2人 = AAR 0.50（健全レンジ）
- `Legacy-Heavy` フェーズ：衰退ではないが、歴史の重みが載っている

**Frontend — Starburst / Mature**：

- Core 6人、Risk 0人——全員がアクティブ
- Architect + Anchor在籍、Risk 0%
- Gravity警告が1件残るが、構造的には健全

**数字が物語を持ち始めた。** 「誰のシグナルが強いか」だけでなく「チームがどんな状態で、次に何が起きるか」が見える。

---

## 良い設計はコモンセンスを生む

BEがLegacy-Heavyなのは、前任のアーキテクトが退任したからだ。彼しか触っていなかったモジュールがいくつも残っている。

しかし、チームは崩壊していない。

なぜか。それらのモジュールが整理された設計のもとで作られていたからだ。完全なドキュメントも、完全な知識移転もなかった。しかし**コードの構造に埋め込まれた設計が、残ったエンジニアに十分な理解を与えた。**

強い設計は、人ではなく構造に知識を残す。強いチームはFormerメンバーのコードを徐々に自分たちのものに置き換えていき、Legacy-Heavyは自然と解消される。EISはその収束をSurvivalの推移として捉える。

---

## 使ってみる

```bash
❯ brew tap machuz/tap && brew install eis
❯ eis team --recursive ~/workspace

# JSON → AIに渡して深掘り分析
❯ eis team --format json --recursive ~/workspace | pbcopy
```

第1章は「この人はどんなエンジニアか？」に答える。
第2章は「このチームはどんな状態か？」に答える。

両方を組み合わせることで：採用（どのRoleが足りないか）、チーム編成（補完性の最大化）、1on1（Impact推移）、リスク管理（劣化の早期検知）。

すべてgit履歴から。アンケートも追加ツールも不要。

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-full.png?v=2)

**GitHub**: [eis](https://github.com/machuz/eis) — CLIツール、計算式、方法論すべてオープンソース。`brew tap machuz/tap && brew install eis` でインストール。

この記事が参考になったら：

[![Sponsor](https://img.shields.io/badge/Sponsor-❤-ea4aaa?logo=github&style=for-the-badge)](https://github.com/sponsors/machuz)

PayPay: `w_machu7`

---

### シリーズ

- [第1章：履歴だけでエンジニアの「戦闘力」を定量化する](https://ma2k8.hateblo.jp/entry/2026/03/11/153212)
- **第2章：個人のシグナルだけでは足りない：Git履歴からチームの健全性を観測する**（本記事）
- [第3章：Architectには流派がある：Git履歴が暴く進化の分岐モデル](https://ma2k8.hateblo.jp/entry/2026/03/14/135648)
- [第4章：Backend Architectは収束する：成仏という聖なる仕事](https://ma2k8.hateblo.jp/entry/2026/03/14/155124)
- [第5章：タイムライン：シグナルは嘘をつかないし、遠慮も映る](https://ma2k8.hateblo.jp/entry/2026/03/14/180329)
- [第6章：チームは進化する——タイムラインが暴く組織の法則](https://ma2k8.hateblo.jp/entry/2026/03/14/184223)
- [第7章：コードの宇宙を観測する](https://ma2k8.hateblo.jp/entry/2026/03/14/213413)
- [第8章：Engineering Relativity：なぜ同じエンジニアでもImpactが変わるのか](https://ma2k8.hateblo.jp/entry/2026/03/14/233602)
- [第9章：Origin：コード宇宙のビッグバン](https://ma2k8.hateblo.jp/entry/2026/03/15/054313)
- [第10章：Dark Matter：見えない重力](https://ma2k8.hateblo.jp/entry/2026/03/15/062608)
- [第11章：Entropy：宇宙は常に無秩序に向かう](https://ma2k8.hateblo.jp/entry/2026/03/15/062609)
- [第12章：Collapse：良いArchitectとBlack Hole Engineer](https://ma2k8.hateblo.jp/entry/2026/03/15/062610)
- [第13章：Cosmology of Code：コード宇宙論](https://ma2k8.hateblo.jp/entry/2026/03/15/062611)
- [第14章：Civilization：なぜ一部のコードベースだけが文明になるのか](https://ma2k8.hateblo.jp/entry/2026/03/15/215211)
- [第15章：AI Creates Stars, Not Gravity](https://ma2k8.hateblo.jp/entry/2026/03/15/221250)
- [最終章：The Engineers Who Shape Gravity：重力を作るエンジニアたち](https://ma2k8.hateblo.jp/entry/2026/03/15/231040)

---

← [第1章：履歴だけでエンジニアの「戦闘力」を定量化する](https://ma2k8.hateblo.jp/entry/2026/03/11/153212) | [第3章：Architectには流派がある →](https://ma2k8.hateblo.jp/entry/2026/03/14/135648)
