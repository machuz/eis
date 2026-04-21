# Engineering Impact Signal (EIS, pronounced "ace") — Claude Code Guide

## 言語ガイドライン

- ユーザーとのコミュニケーションは常に日本語で行う
- コード・コミットメッセージは英語

## 文章スタイル統一指標 (公開テキスト全般の執筆ルール)

OrbitLens 名義で書く公開テキスト (note / X / ブログ / README / docs の読ませる箇所) は、以下を**常に**守る。心理OS の思想と文章の形式を一致させる。

- **熱の方向が外向きの動詞を禁止**: 「動かす」「影響を与える」「届ける」「書き換える」「変える」等の external-directed 動詞を避け、「動く」「保つ」「気づく」「問う」「置く」の internal / observational 動詞を基本に
- **思想者型 (問いかけ) を基本線に**: 「〜だろうか」「〜かもしれない」「〜気がする」「〜ではないだろうか」で、読者が自分で気づける余白を残す
- **哲学者型 (断定・imperative) は必要な場所のみに留める**: 概念定義、結語の punchline、具体TIP などの構造上必要な箇所に限定
- **自分語り / credential flex を排除**: 「僕が EM/VP/CEO を経験して〜」等、主張と直接関係しない経歴は削る。earned な credential (主張と直結する事実) は OK
- **安っぽい比喩を避ける**: 「飛べない鷹は〜」系の手垢付きメタファーは使わない。自ブランド lexicon (「心理OSを明け渡す」「成功の棺」「観測器は届かない」等) を優先
- **原作 (引用元) を薄っぺらくしない**: 福本伸行作品等の引用時、原作の重み (矜持/純度/生死観) を崩さない
- **ターゲットを狭めすぎない**: 「リーダー向け」「EM向け」とラベリングし過ぎない。概念の射程 (役職問わず組織にいる全員) を尊重

詳細アンチパターン・OK例は `~/.claude/projects/.../memory/feedback_writing_voice.md` 参照。

## 作業前の必須事項

- このディレクトリで作業する際は、必ず最初に `docs/` 配下の `blog-en` から始まる記事をすべて読み、EISの設計思想・哲学を十分に理解した上で作業を行うこと
- 思想を理解せずにコードやドキュメントを変更してはならない

## プロジェクト概要

Git履歴のみからエンジニアの実質的なインパクトを定量化するCLIツール。
外部API・AIトークン不要。`git log` と `git blame` だけで7軸観測 + アーキタイプ分類を行う。

- **言語**: Go 1.25.0
- **依存**: fatih/color, rodaine/table, gopkg.in/yaml.v3（最小構成）
- **配布**: GitHub Releases + Homebrew (`brew install machuz/tap/eis`)

## ディレクトリ構造

```
cmd/eis/main.go                        # エントリポイント
internal/
├── cli/
│   ├── root.go                        # コマンドディスパッチ & help
│   └── analyze.go                     # 分析オーケストレーション
├── config/config.go                   # YAML設定パース & バリデーション
├── domain/detect.go                   # ドメイン自動検出 (BE/FE/Infra/FW + カスタム)
├── git/
│   ├── runner.go                      # git コマンド実行ラッパー
│   ├── log.go                         # git log --numstat パーサー
│   └── blame.go                       # git blame パーサー + worker pool
├── metric/
│   ├── metric.go                      # RawScores構造体
│   ├── production.go                  # 変更行数
│   ├── quality.go                     # fix率 (100 - fix%)
│   ├── survival.go                    # 時間減衰blame (exp(-days/tau))
│   ├── design.go                      # アーキテクチャファイル貢献
│   ├── debt.go                        # 負債清掃率
│   └── indispensability.go            # バスファクター（モジュール所有率）
├── scorer/
│   ├── scorer.go                      # Impact計算 & 重み付け
│   ├── normalize.go                   # 正規化
│   └── archetype.go                   # アーキタイプ分類（confidence付き）
└── output/
    ├── table.go                       # カラーテーブル出力
    ├── json.go                        # JSON出力
    └── csv.go                         # CSV出力
```

## 7軸観測

| 軸 | 重み | 正規化 | 概要 |
|---|---|---|---|
| Production | 15% | 絶対値（日次変更量/基準値） | 変更量 |
| Quality | 10% | 絶対値（100 - fix率） | 初回品質 |
| Survival | 25% | 相対値（ドメイン内） | 時間減衰blame生存 |
| Design | 20% | 相対値 | アーキテクチャファイル貢献 |
| Breadth | 10% | 相対値 | リポ横断（3コミット以上） |
| Debt Cleanup | 15% | 絶対値（0-100） | 他者の負債清掃率 |
| Indispensability | 5% | 相対値 | モジュール80%+所有 |

## 3軸エンジニアトポロジー（archetype.go）

v0.9.0で単一アーキタイプから3軸独立分類に移行。破壊的変更。

### Role（何を貢献するか）
1. Architect — Design↑ + RobustSurv↑ + Breadth○（Robust必須）
2. Anchor — Qual↑ + notLow(Prod)
3. Cleaner — Qual↑ + Surv↑ + Debt↑
4. Producer — notLow(Prod)
5. Specialist — Surv↑ + Breadth↓

### Style（どう貢献するか）
1. Builder — Prod↑ + Design↑ + Debt○
2. Resilient — Prod↑ + Surv↓ + RobustSurv○
3. Rescue — Prod↑ + Surv↓ + Debt↑
4. Churn — Prod-Surv gap≥30 + notLow(Prod) + Qual↓ + Surv↓
5. Mass — Prod↑ + Surv↓
6. Balanced — Impact≥30
7. Spread — Breadth↑ + Prod↓ + Surv↓ + Design↓

### State（ライフサイクルフェーズ）
1. Former — RawSurv↑ + Surv↓ + (Design↑ or Indisp↑)
2. Silent — Prod↓ + Surv↓ + Debt↓（commits≥100のみ）
3. Fragile — Indisp≥60 + Prod<40 のゲート通過時、以下の3段で判定:
   - Dormant≥80% ∧ Untested≥50% → 0.90-0.97（最強: 変更圧も保護もない化石）
   - Dormant≥80% のみ → 0.85-0.95（従来の判定）
   - Untested≥70% ∧ Surv≥50 → 0.80-0.90（変更圧データ欠如時のフォールバック）
4. Growing — Prod↓ + Qual↑
5. Active — 直近コミットあり

### 分類ロジック

- soft-match関数: `highness(v)`, `lowness(v)`, `notLow(v)` → 0.0〜1.0
- 最小信頼度閾値: 0.10（これ未満はフィルタ）
- 優先マージン: 0.15以内なら定義順（priority）で解決
- 各軸独立に最良マッチを返す（confidence付き）
- 該当なしは「—」

## ドキュメント更新対象

アーキタイプ追加・変更時に更新が必要なファイル（全4箇所）:

1. **README.md** — アーキタイプテーブル + 説明文
2. **PROMPT.md** — Claude分析プロンプト内のアーキタイプリスト
3. **docs/blog-en-devto.md** — 英語ブログ（テーブル + 説明 + 注釈）
4. **docs/blog-ja-hatena.md** — 日本語ブログ（テーブル + 説明 + 注釈）

出力フォーマット変更時:
- **docs/images/terminal-output.svg** — ターミナル出力のSVG
- **docs/images/archetypes-radar.svg** — レーダーチャートSVG

## リリースフロー

```bash
# 1. コミット & プッシュ
git add <files> && git commit -m "feat: ..." && git push origin main

# 2. タグ作成 & プッシュ
git tag v0.X.Y && git push origin v0.X.Y

# 3. goreleaser実行（ローカル）
GITHUB_TOKEN=$(gh auth token) goreleaser release --clean
# 注: HOMEBREW_TAP_GITHUB_TOKEN が未設定だとbrew更新だけ失敗する（バイナリは公開される）

# 4. Homebrew tap手動更新（HOMEBREW_TAP_GITHUB_TOKENがない場合）
cat dist/homebrew/eis.rb > /tmp/homebrew-tap/eis.rb
cd /tmp/homebrew-tap && git add eis.rb && git commit -m "update eis to v0.X.Y" && git push

# 5. ローカル更新
brew upgrade eis
# または
go install github.com/machuz/eis/cmd/eis@latest
```

### GitHub Actionsリリース（タグプッシュで自動）

`.github/workflows/release.yml` がタグ `v*` で自動実行。
`HOMEBREW_TAP_GITHUB_TOKEN` シークレットが必要。

## テスト実行

```bash
go test ./...
```

## ポータル同期ルール

- eisリポの `research/` や `docs/` に公開HTMLページを追加・変更した場合、orbitリポの `pages/index.html`（ポータル）にもリンクを追加・更新すること
- ポータルの場所: `/Users/machu/Mydev/myGithub/orbit/pages/index.html`
- Cloudflare Pages（orbit-d8x.pages.dev）で自動デプロイされる

## 自律的に進めてよい作業

- archetype.goの修正 → ドキュメント4箇所の更新 → コミット → リリースまで一気通貫でOK
- SVGの更新
- ブログ記事の更新（日英両方）
- Homebrew tap更新

## 設計思想

- **ゲーム耐性**: 時間減衰survivalは忙しさでは膨らまない。残ったコードだけがカウントされる
- **コメント非計上**: コードファイル（Go/TS/Py/SQL等）のコメント行・空行は Production/Survival/Design/Debt から除外。コメント量産でスコアを水増しできない。`.md`/`.txt` 等の散文ファイルは論文・研究用途を想定して全行カウント（`internal/git/comment.go`）
- **テスト加重 Survival** (v2.0.0): テストに守られていないコード行の生存スコアは α=0.5 に減衰。`_test.go` 等の兄弟ペア or モジュール単位で判定（`internal/metric/test_detection.go`）。テスト文化不在リポもそのまま半減 — 望遠鏡は曲げない。SaaS 側で `tested_survival` / `untested_survival` を個別に読み取り可能。config `untested_survival_weight` で上書き可
- **ドメイン分離**: BE/FE/Infra/FW（+ カスタムドメイン）は別々に観測。混ぜると汚染される
- **ハイブリッド観測**: 絶対値（組織横断比較可能）+ 相対値（ドメイン内順位）
- **40点 = シニア**: 7軸で40+を出すのは意図的に厳しい基準

## EIS と ace-orbit の責務分担

- **EIS（CLI）** = 観測データの生成器。git log / git blame から7軸シグナル・3軸エンジニアトポロジー・3軸モジュールトポロジーを算出し、JSON で出力する
- **ace-orbit（SaaS）** = 観測データの解釈・推薦・予測を担う。Structural Summary、人×モジュール突合（Conway's Law検証）、時間系列リスク予測、アラートはすべて SaaS 側の機能

CLI に推薦ロジックや予測機能は入れない。CLI は望遠鏡、SaaS は天文台。

## 今後の方向性

impact metric → **engineering risk detector** への進化は v1.2.0 〜 v2.2.0 で段階完成:
- ~~`change_pressure = commits_touching_module / module_LOC` で変更圧を定量化~~ v1.2.0
- ~~`tested_survival` vs `untested_survival` でrobust code vs dormant codeを分離~~ v2.0.0
- ~~Fragile アーキタイプ判定に `untested_survival` を組み込む~~ v2.1.0
- ~~Module Topology に `Vitality=Fragile` 新設~~ v2.2.0

**Module Topology Vitality (6 levels)** (`internal/scorer/module_archetype.go`):
- **Stable** — 低変更圧 + テスト有り or coverage データ無し。健康な平衡
- **Fragile** (v2.2.0) — 低変更圧 + テスト率 < 10% + 生存コード有り。ただ触られてないだけの化石フォートレス
- **Warming** — 変更圧上昇中
- **Turbulent** — 高変更圧 + 低生存率
- **Critical** — 極端な Turbulent
- **Dead** — コミット無し + 所有者不在

次の方向性は未定。engineering risk detector は一旦完成した。
