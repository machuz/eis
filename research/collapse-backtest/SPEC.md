# Collapse × Conformance — Predictive-Validity Backtest (spec v1)

**目的**: gravity/survival を defect 予測（repowise の土俵）でなく、**崩壊（去ると死ぬ構造）**の予測に向け、その**予測妥当性を実証**する。堀の第一ゲート兼、repowise が churn では語れないデータ・ストーリー。

---

## 1. 仮説（No Escape Clauses）

> **崩壊リスク = 集中度 × 非適合度。**
> あるモジュールが (A) 少数の保持者に集中し、かつ (B) codebase の構造パターンから逸脱しているとき、その保持者が去ると、そのモジュールは**期待減衰を超えて死ぬ／凍結する**。
> 逆に、集中していても**適合している**モジュールは、去っても崩壊しない（＝ bus-factor 単体は過検出）。

検証する差分主張は2つ:
- **H1**: 非適合度は、bus-factor 単体を超えて崩壊を予測する（適合した単独所有は false alarm）。
- **H2**: 非適合度 × 集中度の**交互作用**が、churn-health（repowise 型）より崩壊をよく当てる。

---

## 2. 単位と記法

- 単位 = **モジュール** m（ディレクトリ級、eis module topology）。
- 観測時刻 T0 = 保持者 a の離脱直前（`T_dep − 3mo`、離脱前スローダウンのリークを避けるため 3ヶ月マージン）。
- 結末窓 = `[T_dep, T_dep + N]`、N = **12ヶ月**（primary）、6/18ヶ月で頑健性確認。
- すべての feature は **≤ T0** のデータのみ、結末は **> T_dep** のみ（リークなし、repowise の T0→次6ヶ月 と同型）。

---

## 3. 予測子（T0 で測る）

### A. 集中度（誰が保持するか）— eis の非ゲーム的 edge
- `owner_share(m,a)` = モジュール m の**生存加重 gravity** のうち著者 a が持つ割合（commit数/blame% でなく **time-decayed survival 加重**）。
- `bus_factor(m)` = 生存 gravity の 50% を覆う最小著者数。
- 集中シグナル = 高 owner_share / bus_factor=1。

### B. 非適合度（特異か）— 新シグナル、安い2つ

**S1: co-change scatter（最安・グラフ再利用）**
```
C(m) = m と閾値超で co-change するモジュール集合（lift ≥ τ かつ support ≥ k）
D(m) = m と import/依存辺を持つモジュール集合
scatter(m) = |C(m) \ D(m)| / |C(m)|      # 隠れ結合の割合
```
高 scatter = アーキテクチャに沿わない隠れ結合 = 非適合。

**S2: token-entropy naturalness（安い・rigorous・leave-one-out）**
```
LM = n-gram token モデル（n=4, add-k or Kneser-Ney）を
     git_sha(T0) の「m を除く」全 repo で学習          # leave-one-out 必須
naturalness(m) = m のソースの平均 cross-entropy（bits/token）
```
高 = 驚き = 特異。repo 内で percentile 正規化して比較可能に。多言語は per-file→行数加重集約。

> **backtest では S1/S2 を事前合成しない。** どちらが signal を担うか分析で見てから合成する。

---

## 4. 崩壊 proxy（結末 Y、`[T_dep, T_dep+N]` で測る）

崩壊 = 知識が転移されなかった。2つの failure mode を別々に測る:

**Y1: excess-death（primary・survival ネイティブ）**
- 離脱著者 a の**生存していた行**が、離脱後に **codebase の baseline half-life から期待される減衰を超えて**死ぬ率。
- `excess_death(m) = actual_death_rate(a's lines in m, [T_dep,T_dep+N]) − expected_decay(baseline_half_life)`
- 意味 = 「誰も理解できず rewrite/rip-out された」= 崩壊の中核署名。**eis の survival をそのまま結末に使える**のが美点。

**Y2: abandonment（凍結型 collapse）**
- 離脱後、m への commit 活動が **repo 全体トレンドを超えて**急落。
- `abandon(m) = activity_drop(m) − activity_drop(repo)`   # repo 全体の衰退を差し引く
- 意味 = 「怖くて誰も触らない」= もう一つの崩壊。

**合成 collapse ラベル** = `Y1 が上位 q% ` OR `Y2 が上位 q%`（q は感度分析）。分析では Y1/Y2 分離も報告。

---

## 5. Departure イベントの定義

離脱著者 a in repo r:
- **活動期**: T_dep 以前に ≥ 6ヶ月かつ ≥ K commit（K=20 目安）。
- **離脱**: 最終 commit T_dep 以降、**≥ 12ヶ月無 commit**。
- **保持**: a が owner_share ≥ 0.4 を持つモジュール m のみ対象（a が実質保持していた所）。

**交絡ガード**:
- repo 全体が死んだケースを除外（repo-level 活動が同時に急落 → Y2 は repo トレンド差し引きで吸収、Y1 は baseline 再計算）。
- sabbatical/復帰（12ヶ月以内に復帰）は離脱と見なさない。

---

## 6. Backtest プロトコル

1. **コーパス**: OSS repo 群（既取得分＋補充）。条件 = 十分な履歴・複数貢献者・観測可能な離脱。目標 = **離脱×モジュール ケース 100〜300**（repowise の 28 repo/112k commit に対抗できる N）。
2. 各 `(repo, a, m)` について T0 で feature（A, S1, S2 ＋ controls）、`[T_dep,T_dep+N]` で Y。
3. **モデル**:
   ```
   logit(collapse) ~ concentration + nonconformance
                     + concentration × nonconformance     # ← 本命の交互作用項
                     + controls(module_size, a_abs_gravity, module_age, language, repo_FE)
   ```
4. **ベースライン（超えるべき相手）**:
   - B0 = bus_factor=1 単体（集中のみ）
   - B1 = churn-health（recent churn × complexity、repowise 型）
5. **指標**: ROC AUC（交互作用モデル vs B0 vs B1）、**DeLong 検定**で AUC 差の有意性（repowise の 0.74/DeLong p<1e-9 と同じ土俵）。lift = ΔAUC。

---

## 7. 成功基準（何が出れば勝ちか）

- **H1**: 適合した単独所有モジュールは崩壊しない（B0 の false-alarm 率が高い）、非適合な単独所有は崩壊する。→ デモの「bus-factor 1 × idiosyncrasy」修正が実証される。
- **H2**: 交互作用モデルの AUC が **B0 を +Δ、B1（churn-health）を +Δ 上回る**（repowise は churn を +0.10 で超えた。同等以上を狙う）。
- 出れば: 「**去ると死ぬコードは、bus-factor でなく*非適合*が言い当てる**」＝ repowise が churn では絶対言えない finding。

---

## 8. コスト / 決定論

- feature は全て **T0 snapshot の線形計算**（履歴で回さない）。survival/ownership は eis 既存。scatter は topology グラフ再利用。naturalness は T0 で m を除く n-gram 1 パス（対象=離脱者保持モジュールのみ、全 repo でも O(tokens)）。
- 決定論: 固定 git_sha の n-gram/グラフカウント = W-02 準拠。observation であって generation でない = W-01 準拠（embedding は使わない）。
- Y は survival 曲線＋活動 log から。既取得 OSS 観測を再利用。

---

## 9. 段階（cheap-first、過剰投資回避）

1. **Phase 0**: S1(co-change scatter) ＋ 素朴 S2(token-entropy) だけを既取得 OSS に乗せ、H1/H2 を回す。
2. lift が出たら **Phase 1**: AST 形状 naturalness（tree-sitter）で S2 を精緻化。
3. **出なければ AST は作らない**。安い signal で thesis が死ぬなら、そこで方向転換。

---

## 10. 決めるべき選択点（本当に選択が要るのはここだけ）

1. **コーパス**: 既取得 OSS はどの repo・何本・離脱貢献者を含むか? 補充が要るか?（離脱ケース 100+ を確保したい）
2. **崩壊 proxy の主軸**: Y1(excess-death) を primary で確定でよいか? Y2(abandonment) は補助でよいか?
3. **離脱の閾値**: 活動期 ≥6mo/≥20commit・離脱 ≥12mo無 commit で確定でよいか?
