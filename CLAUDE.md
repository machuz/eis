# Engineering Impact Signal (EIS, pronounced "ace") — Claude Code Guide

## 言語ガイドライン

- ユーザーとのコミュニケーションは常に日本語で行う
- コード・コミットメッセージは英語

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
3. Fragile — Surv↑ + Prod↓ + Qual<70（変更圧がないから残っているだけ）
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

ブログ記事変更時に同期が必要なファイル:
- `docs/blog-ja-zenn-chX.md` を変更したら → `books/git-archaeology/chX.md` も必ず同期
- Zenn Book版はフロントマターが異なる（titleのみ、emoji/type/topics/publishedなし）
- Zenn Book版のリンクは相対パス（`ch1`, `ch2` 等）、docs版はフルURL
- dev.to / はてなはCIで自動デプロイされるが、Zenn BookはGitHub連携で直接反映

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

impact metric → **engineering risk detector** への進化:
- ~~`change_pressure = commits_touching_module / module_LOC` で変更圧を定量化~~ 実装済（v1.2.0〜）
- ~~`tested_survival` vs `untested_survival` でrobust code vs dormant codeを分離~~ 実装済（v2.0.0）
- **次**: Fragile Fortress アーキタイプ判定に `untested_survival` を組み込む（別PR）
- Module Topology の `Vitality=Fragile` 精密化も同時に
