---
title: "個人シグナルの先へ：チームの健全性をGit履歴から観測する"
---

*個人のシグナルは「誰のシグナルが強いか」を教えてくれる。チームの健全性は「来四半期もこのチームが強いかどうか」を教えてくれる。*

![チーム構造とヘルスレーダー](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ch2-iconic.png?v=4)

## なぜ個人シグナルだけでは不十分か

全員がImpact 80超えのチームが必ずしも強いとは限らない。全員がProducerなら、アーキテクチャを触る人がいない。負債を片付ける人もいない。出荷速度は凄まじいが、コードは3ヶ月で腐る。

逆に、平均50台でもArchitectが1人、Cleanerがいて、Growingの若手が育っているチームは強い。半年後にはもっと強くなる。

**強いチームは個人シグナルの総和ではない。構成と補完性が重要だ。**

---

## なぜ売上では技術組織の強さが見えないか

「売上が伸びているから技術組織は大丈夫」——危険な思い込みだ。売上が測っているのは**プロダクト・マーケット・フィット**であって、**エンジニアリングの健全性**ではない。

売上は車の速度。エンジニアリングの健全性はエンジンの状態。エンジンが壊れかけていても、下り坂なら速度は出る。

Git履歴には、売上が捉えられないシグナルがある：

- **コードの耐久性** —— 同じ機能を毎四半期書き直していないか？
- **技術的負債** —— 機能1つ追加するたびにバグ修正が2つ発生していないか？
- **バス係数** —— 1人辞めたらどのモジュールが死ぬか？

**売上が伸びていても、Survival低下 + Debt増加 + Bus Factor集中が同時進行していれば、スケールした瞬間に組織は崩壊する。**

---

## チームヘルス7軸

`eis team` は個人シグナルをチームレベルの健全性に集約する：

```bash
❯ eis team --recursive ~/workspace
```

| 軸 | 何を測るか | 要点 |
|---|---|---|
| **Complementarity** | Roleの多様性（Architect, Anchor, Cleaner, Producer, Specialist） | Producer一色のチームは16。全種揃うと100 |
| **Growth Potential** | Growingメンバー + Builder/Cleanerのロールモデル | ロールモデルなしでは若手は育たない |
| **Sustainability** | リスク状態（Former, Silent, Fragile）の逆数 | チーム速度の隠れた足かせ |
| **Debt Balance** | Debt Cleanup平均。50超 = チーム全体が負債を減らしている | 自浄傾向 |
| **Productivity Density** | 人あたり生産量 + 少人数ボーナス | 「この人数でこのアウトプット」 |
| **Quality Consistency** | 平均品質 + 低分散 | 平均80でもレンジ95-40なら健全ではない |
| **Risk Ratio** | Former/Silent/Fragile の割合 | 25%超で要注意。50%超で危機的 |

> 各軸の計算式：[Whitepaper](https://github.com/machuz/eis)

---

## チーム分類 —— 銀河形態学

EISは個人トポロジーからボトムアップで、チームを**5軸**で分類する：

![チーム分類フロー](https://raw.githubusercontent.com/machuz/eis/main/docs/images/team-classification-flow.png?v=4)

| 軸 | 導出元 | 問い |
|---|---|---|
| **Structure** | Role分布 | どんな構造的役割があるか |
| **Culture** | Style分布 | どういうやり方で仕事をしているか |
| **Phase** | State分布 | ライフサイクルのどこにいるか |
| **Risk** | 健全性指標 | どんなリスクを抱えているか |
| **Character** | 上4軸の複合 | 一言で言うとどんなチームか |

Characterは**銀河形態学**のメタファを使う。望遠鏡は銀河の形を記述するが、良し悪しは判断しない：

| Character | 銀河 | 意味 |
|---|---|---|
| **Spiral** | 渦巻銀河 | 強い中心核 + 活発な星形成。設計と生産が両立 |
| **Elliptical** | 楕円銀河 | 成熟、安定、変化耐性。低エントロピー |
| **Starburst** | スターバースト銀河 | 爆発的成長。エネルギー高、構造は形成途中 |
| **Nebula** | 星雲 | 次世代エンジニアが育つ場所 |
| **Irregular** | 不規則銀河 | 重力中心なし。高出力だが方向性がない |
| **Dwarf** | 矮小銀河 | 小さいが長寿。安定した品質 |
| **Collision** | 衝突銀河 | 構造的混乱。常に火消し |

> 天文学的な解説付きの完全ガイド：[Galaxy Morphology Guide](https://orbit-d8x.pages.dev/galaxy-guide.html)

分類は**Impactで重み付け**される。Impact 90のArchitectはImpact 15のArchitectより圧倒的にチームの色を決める。強いシグナルを持つメンバーほど多くの文化を伝播させる。

---

## 成長モデル

EISのRole分類は3つの層にマッピングされる：

![成長モデル](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ch2-diagram-growth-model.png?v=4)

**実装層** → **安定化層** → **設計層**

- Survival上昇 → 実装層から安定化層へ
- Design上昇 → 安定化層から設計層へ
- DebtCleanup上昇 → チーム貢献の幅が広がっている

Growth Potentialが高いチームには、この階段を登れる環境がある。各層にロールモデルがいるからだ。ロールモデルがなければ、Growingメンバーは実装層で回り続ける。

![衰退モデル](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ch2-diagram-decline.png?v=4)

**BuilderかCleanerが1人いるチームは人が育つ。** ロールモデルがいるとき、GrowingメンバーがActiveに遷移する速度はおよそ倍になる。**Architect不在のチームは時間とともに劣化する。**

---

## メンバーティア

git上に名前が出る全員が「チームメンバー」とは限らない。EISはメンバーを3層に分ける：

| Tier | 条件 | 用途 |
|---|---|---|
| **Core** | `RecentlyActive && Impact >= 20` | 平均、Density、Consistency |
| **Risk** | Former / Silent / Fragile | RiskRatio、分類 |
| **Peripheral** | それ以外 | カウントのみ |

ヘッダーは `4 core + 3 risk / 16 total` と表示される。一時的な貢献者は指標を歪めず、Silentなメンバーは検知される。

EISは**自動警告**も出力する。バス係数リスク、Silent蓄積、Gravity脆弱性、トップ貢献者集中。

![Team Warnings](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ch2-warnings.png?v=4)

---

## 実測結果

実際のプロダクト（Backend 12リポ + Frontend 9リポ）で `eis team` を実行した結果：

**Backend — Spiral / Legacy-Heavy**：

- Core 4人で12リポを運用、Risk 3人（Silent 2 + Former 1）
- Architect + Anchor 2人 = AAR 0.50（健全レンジ）
- `Legacy-Heavy`：衰退ではないが、歴史の重みが載っている

**Frontend — Starburst / Mature**：

- Core 6人、Risk 0人 —— 全員がアクティブ
- Architect + Anchor在籍、Risk 0%
- Gravity警告が1件残るが、構造的には健全

**数字が物語を持ち始めた。** 「誰のシグナルが強いか」だけでなく「チームがどんな状態で、次に何が起きるか」が見えるようになった。

---

## 良い設計は常識を生む

BEチームがLegacy-Heavyなのは、前任のアーキテクトが退任し、彼しか触っていなかったモジュールが残っているからだ。

しかしチームは崩壊していない。

なぜか。それらのモジュールが整理された設計のもとで作られていたからだ。完全なドキュメントも知識移転もない。だが**コードの構造に埋め込まれた設計が、残ったエンジニアに十分な理解を与えている。**

強い設計は人ではなく構造に知識を残す。強いチームでは、Former メンバーのコードが徐々に自分たちのコードへと置き換わっていき、Legacy-Heavy は時間とともに解消される。EISはその収束をSurvivalの推移として捉える。

---

## 使ってみる

```bash
❯ brew tap machuz/tap && brew install eis
❯ eis team --recursive ~/workspace

# JSON → AIに貼って深掘り
❯ eis team --format json --recursive ~/workspace | pbcopy
```

第1章は「この人はどんなエンジニアか」に答える。
第2章は「このチームはどんな状態か」に答える。

組み合わせれば：採用（どのRoleが足りないか）、チーム編成（補完性の最大化）、1on1（Impactの推移）、リスク管理（劣化の早期検知）。

すべてGit履歴から。サーベイ不要。追加ツール不要。

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-full.png)

**GitHub**: [eis](https://github.com/machuz/eis) — CLIツール、計算式、方法論すべてオープンソース。`brew tap machuz/tap && brew install eis` でインストール。
