# git考古学 #0 —— 3分でわかるEngineering Impact Score

![Cover](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/hatena/cover-ch0.png?v=1)

*git log と git blame だけで、エンジニアの「戦闘力」が見える。*

---

## これは何か

**Engineering Impact Score（EIS、読み：ace）** は、Git履歴だけからエンジニアの技術的インパクトを定量化するOSSのCLIツールだ。

外部API不要。AIトークン不要。`git log` と `git blame` だけで動く。

```bash
brew tap machuz/tap && brew install eis
cd your-repo
eis
```

これだけで、こういう出力が得られる：

![Terminal Output](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/terminal-output.svg)

---

## なぜ作ったか

コミット数、PR数、変更行数——どれも測りやすいが、どれも本質を捉えていない。

タイポ修正もシステム全体の設計変更も「1 PR」。生成されたlockfileで数千行。コミット頻度は人による。

でもチームの中では、誰が強いか、みんなわかっている。

> 「あの人が書いたコードは残る」
> 「あの人はいつも触ってるけど、なぜか良くならない」

その直感を**数字にしたかった**。

---

## 望遠鏡の話

自分はこれまで何度か、強いエンジニアに「一緒に働こう」と声をかけてきた。ありがたいことに、来てくれた人が何人もいる。

彼らがなぜ来てくれたのか。技術スタックや報酬だけではないと思っている。

**「この人は自分の仕事をちゃんと見てくれる」**——そう感じてもらえたのではないか。

エンジニアにとって、自分の技術的な仕事が正しく観測されることは大きい。コミット数やPR数ではなく、**コードが残っているか、構造に貢献しているか、負債を片付けているか**——そういう本質を見てくれる目があること。

自分にはその目があった。少なくとも、そう自己認識している。

EISは、その**観測者の目を望遠鏡としてOSSにしたもの**だ。

誰でも使える。誰のチームでも覗ける。git履歴という、嘘のつけないレンズを通して。

---

## 構造を科学する

数学が強い。アルゴリズムが強い。言語仕様が強い。

これらにはアカデミックの世界で長年揉まれた理論がある。計算量理論、型理論、形式検証——数学的証明で正しさを保証できる、科学の土台がある。

ソフトウェアアーキテクチャにも学術的な試みは30年以上ある。Architecture Description Languagesや評価手法は提案されてきた。しかし**統一的な理論にはなっていない**。断片的で、実務に降りてきていない。

「良い設計とは何か」「このチームの構造は健全か」——こうした問いに対して、業界にはベストプラクティスや経験則はあっても、定量的な言葉が少ない。

そしてAIが大量にコードを書く時代が来た。

**コードを書く能力の価値は相対的に下がる。何より大事なのは構造だ。** どういう構造の上にコードを載せるか。その構造は変更に耐えるか。チームの知識はどこに蓄積されているか。

EISがその**構造を科学するための道具**になれたら嬉しい。

---

## 7つの軸

EISは7つの軸でスコアリングする。

| 軸 | 重み | 何を測るか |
|---|---|---|
| Production | 15% | 変更量 |
| Quality | 10% | 初回品質（fix率の低さ） |
| **Survival** | **25%** | **書いたコードが今も残っているか（時間減衰付き）** |
| Design | 20% | アーキテクチャファイルへの貢献 |
| Breadth | 10% | リポジトリ横断の活動 |
| Debt Cleanup | 15% | 他者が残した負債の清掃 |
| Indispensability | 5% | モジュール所有率（バスファクター） |

最も重要なのは**Survival**だ。書いたコードが半年後も残っているか。1年後も残っているか。

大量に書いても翌月に書き換えられるなら、それは強さではない。**残るコードを書ける人が強い。**

---

## 3軸のアーキタイプ

スコアだけではない。EISはエンジニアを3つの軸で分類する。

**Role** — 何を貢献するか
- Architect：構造を設計する人
- Anchor：品質を守る人
- Cleaner：負債を片付ける人
- Producer：量を生産する人
- Specialist：特定領域の専門家

**Style** — どう貢献するか
- Builder：作りながら設計する
- Resilient：壊されても再建する
- Churn：量は多いが残らない

**State** — ライフサイクル
- Former：退職したが資産が残っている
- Growing：まだ量は少ないが品質が高い
- Active：現在活動中

この分類から、チームの構造が見えてくる。

---

## たとえばこんなことがわかる

- **退職したArchitectのコードが今もコードベースの30%を占めている**（Former検出）
- **品質は高いが変更圧がないから残っているだけのコード**がある（Fragile検出）
- **チームにProducerがいない**——構造の上で量を生産する層が空白（Producer Vacuum）
- **Architect Bus Factor = 1**——設計者が一人に集中している

冷たいgit履歴から、こういう**チームの物語**が読み取れる。

---

## シリーズ目次

このブログシリーズ「git考古学」では、EISを使って実際のチームを分析し、何が見えたかを書いている。

1. **[履歴だけでエンジニアの「戦闘力」を定量化する](https://machu.hateblo.jp/)** — 7軸スコアリングの全設計
2. **チームトポロジー** — スコアからチーム構造が見える
3. **FE Architectは分岐する** — フロントエンドの進化モデル
4. **BE Architectは収束する** — バックエンドの進化モデル
5. **Infraは沈黙する** — インフラエンジニアの可視化問題
6. **Gravity Map** — コードの重力場を可視化する
7. **リスク検出** — 技術的リスクの定量化
8. **正規化の設計** — なぜハイブリッドスコアリングか
9. **ドメイン分離** — BE/FE/Infraを混ぜると汚染される
10. **時間減衰の設計** — Survivalの数学
11. **アーキタイプの設計** — 分類ロジックの全貌
12. **チーム指標** — チームの健康診断
13. **Robust Survival** — テスト済みコードの生存
14. **変更圧** — change pressureの定量化
15. **複数リポ分析** — 組織横断のスコアリング
16. **未来** — EISの次

---

## インストール

```bash
# Homebrew
brew tap machuz/tap && brew install eis

# Go
go install github.com/machuz/engineering-impact-score/cmd/eis@latest
```

**GitHub**: [engineering-impact-score](https://github.com/machuz/engineering-impact-score)

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/logo-full-zenn.png)
