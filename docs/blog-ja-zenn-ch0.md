---
title: "git考古学 #0 —— 3分でわかるEngineering Impact Signal"
emoji: "🔭"
type: "tech"
topics: ["git", "engineering", "productivity"]
published: true
---

![Cover](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/hatena/cover-ch0.png?v=2)

*git log と git blame だけで、エンジニアの「戦闘力」が見える。*

---

## これは何か

**Engineering Impact Signal（EIS、読み：ace）** は、Git履歴だけからエンジニアの技術的インパクトを観測するOSSのCLIツールだ。

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

でもチームの中では、誰のシグナルが強いか、みんな感じ取っている。

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

## 誠実な人が勝つ構造を作りたい

日本のエンジニアの給与は、諸外国に比べて低い。

技術力が劣っているからではない。文化的な優しさと、自分の価値を主張しないことからきていると思っている。黙ってコードを書き、黙って構造を直し、黙って負債を片付ける——そういう仕事は**見えない**。見えないから、伝わらない。伝わらないから、その人の働き——魂——が、声のでかいやつに吸われてしまう。

そういう感覚を覚えたとき、自分はチームの磁場を歪ませるほどの抵抗をしてきた。

優秀でモノづくりに真摯に向き合い続ける人の仕事が、**ちゃんと見える世界にしたい**。

EISを作っていく中で、自分が本当にやりたかったことの輪郭が見えてきた。**誠実にモノづくりに向き合う人が勝つ構造を作ること**だ。その自己認識を得た瞬間、熱量が爆発して、短期間でこのツールをここまで具現化できた。

そして今、望遠鏡の次を考えている。

望遠鏡は宇宙を**観測する**。しかし観測だけでは、エンジニアの人生は変わらない。観測データを**解釈**し、その人に合った宇宙——つまりコードベース、チーム、組織——を**提案**し、その宇宙の中で安定する軌道を**提示する**。そこまでやって初めて、「誠実な人が勝つ構造」になる。

それがEISの次のステップだ。望遠鏡を天文台にする。

---

## 構造を科学する

数学が強い。アルゴリズムが強い。言語仕様が強い。

これらにはアカデミックの世界で長年揉まれた理論がある。計算量理論、型理論、形式検証——数学的証明で正しさを保証できる、科学の土台がある。

ソフトウェアアーキテクチャにも学術的な試みは30年以上ある。Architecture Description Languagesや評価手法は提案されてきた。しかし**統一的な理論にはなっていない**。断片的で、実務に降りてきていない。

「良い設計とは何か」「このチームの構造は健全か」——こうした問いに対して、業界にはベストプラクティスや経験則はあっても、定量的な言葉が少ない。

そしてAIが大量にコードを書く時代が来た。

**コードを書く能力の価値は相対的に下がる。何より大事なのは構造だ。** どういう構造の上にコードを載せるか。その構造は変更に耐えるか。チームの知識はどこに蓄積されているか。

EISがその**構造を科学するための道具**になれたら嬉しい。

そして今、EISは人だけでなく**モジュールそのもの**も観測する。そのために4つの指標を設計した：

| 指標 | 何を測るか |
|---|---|
| **Change Pressure** | モジュールへの変更頻度÷コード量。変更圧が高いほど構造的ストレスが大きい |
| **Co-change Coupling** | 同時に変更されるモジュールの組み合わせ。import文に現れない暗黙の結合を検出する |
| **Module Survival** | モジュール内コードの時間減衰付き生存率。書かれたコードが残っているか |
| **Ownership Fragmentation** | モジュールの知識がどう分布しているか。Shannon entropyで測定 |

これらの指標を組み合わせて、すべてのモジュールを3つの独立した軸——Coupling（境界品質）、Vitality（変更圧×生存力）、Ownership（知識分布）——で分類する。目に見えない構造的リスクが観測データに変わる：

- `Hub × Critical × Orphaned` — 暗黙結合の中心点で、変更圧が極端に高く、かつオーナー不在。最大リスク
- `Independent × Stable × Distributed` — 境界がきれいで、所有も健全。理想形

**望遠鏡は、星（エンジニア）だけでなく、星が存在する空間（モジュール）も観測する。**

---

## 7つの軸

EISは7つの軸で観測する。

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

シグナルだけではない。EISはエンジニアを3つの軸で分類する。

**Role** — 何を貢献するか
- Architect：構造を設計する人
- Anchor：品質を守る人
- Cleaner：負債を片付ける人
- Producer：量を生産する人
- Specialist：特定領域の専門家

**Style** — どう貢献するか
- Builder：作りながら設計する
- Resilient：壊されても再建する
- Rescue：他者の負債を救済する
- Churn：量は多いが残らない
- Mass：大量生産だがSurvivalが低い
- Balanced：全軸バランス型
- Spread：広く浅く触るが深さがない

**State** — ライフサイクル
- Former：退職したが資産が残っている
- Silent：活動もSurvivalも低い（経験者のみ検出）
- Fragile：変更圧がないから残っているだけ
- Growing：まだ量は少ないが品質が高い
- Active：現在活動中

この分類から、チームの構造が見えてくる。

---

## たとえばこんなことがわかる

- **退職したArchitectのコードが今もコードベースの30%を占めている**（Former検出）
- **品質は高いが変更圧がないから残っているだけのコード**がある（Fragile検出）
- **チームにProducerがいない**——構造の上で量を生産する層が空白（Producer Vacuum）
- **Architect Bus Factor = 1**——設計者が一人に集中している
- **136のOrphanedモジュール**——オーナーが離脱し、知識を持つ人がいない（Module Topology）
- **12のCriticalモジュール**——変更圧が高くコードが生存しない。構造的な時限爆弾（Module Topology）

冷たいgit履歴から、こういう**チームの物語**が読み取れる。そしてモジュールトポロジーは「誰のシグナルが強いか」だけでなく、**「どこが壊れかけているか」**を教えてくれる。

---

## OSSの宇宙でも検証した

望遠鏡が正しく機能するかを確かめる最良の方法は、**すでに構造が知られている宇宙を観測する**ことだ。

EISを**29のOSSリポジトリ、55,343人のエンジニア**に対して実行した。React、Kubernetes、Rails、Laravel、esbuild、Rust——誰もが構造を知っているプロジェクトだ。

観測結果は、コミュニティの直感と一致した。

- **esbuild**：Evan Wallaceが全軸100。重力集中度92.5%——「あれはEvanが一人で作った」という共通認識そのまま
- **Rails**：Design 35超のアーキテクトが6人。20年かけて設計権限を分散させた文明——DHH、Jeremy Kemper、Rafael Francaら
- **Laravel**：Taylor Otwellが100、他のTop10は全員Design 4未満——「Taylorの作品」という認識そのまま
- **React**：10年で5世代のアーキテクト交代——Paul O'Shannessy → Dan Abramov → Brian Vaughn → Sebastian Markbåge → Jorge Cabiedes。Gravity Mapはコミット数の多寡とは異なる力学を浮き彫りにした。最も象徴的なのはJorge Cabiedesだ。わずか**82コミット**でGravity 60——2,010人中の重力圏Top 3に浮上した。Design 100、Production 100、Debt Cleanup 96。一方、Dan Abramovは**1,890コミット**でGravity 51.1。コミット数は23倍だが、重力はJorgeが上回る。Brian Vaughn（1,627コミット、Gravity 61.6）とほぼ同等の重力を、20分の1のコミット量で生み出している。Jorgeのスタイル分類は**Builder**——設計し、構築し、片付けまでやる全方位型だ。2025年のGravity分布に突如現れ、Design 100はTop 5の中でJorge**だけ**が持つ数字だ（Sebastian Markbåge 7、Dan Abramov 10、Brian Vaughn 4、Andrew Clark 20）。コードベースの重力場はコミットの**量**ではなく**構造への関与の質**で決まる。82コミットでも、それがアーキテクチャファイルに集中し、残り続け、負債を片付けるコードであれば、1,890コミットより強い重力を発揮する。git logの行数をスクロールしても見えない力が、git blameの地層には刻まれている。もう一つ、時系列を辿ることで初めて見える物語がある。Sebastian Markbågeだ。2016年から2023年まで、Sebastianの重力は**常にそこにあったが、中心ではなかった**。68→72→33→60→66→63→63→38。Paul O'Shannessyの重力場、Dan Abramovの重力場、Brian Vaughnの重力場——時代ごとに重力の中心が移り変わる中で、Sebastianは常にTop 5の重力圏内にいながら、一度も重力の極点にはならなかった。そして2024年——Gravity **100**。重力場の中心が、静かにSebastianへ移った。何が起きたのか。Sebastianのロール分類は**Cleaner**だ。新機能を量産するProducerでも、設計図を描くArchitectでもない。**他者が残した構造的負債を片付け、コードベースの整合性を守り続ける人**。Debt Cleanup 54、Quality 96.2、そしてIndispensability **100**——Reactのコードベースで、Sebastianのコードなしには成立しないモジュール占有率が最大という意味だ。8年間の重力推移を並べると、派手な数字は一つもない。だが彼のコードは**残り続けた**（Survival 79.2）。Dan Abramovが1,890コミットでSurvival 0.1、Brian Vaughnが1,627コミットでSurvival 0.1——書き換えの波に飲まれてコードが消えていく中で、Sebastianの1,495コミットはSurvival 79.2で地層に刻まれ続けた。他のエンジニアが去り、コードが書き換えられるたびに、**消えなかったコードの比率**が相対的に上がっていく。2024年、閾値を超えた。静かに積み上げた地層が、重力場そのものになった瞬間だ。スナップショットでは見えない。年次の重力分布を時系列で辿って初めて、「8年間、重力圏の中心ではなかったCleanerが、コードベースの地殻変動を経て重力の極点になった」という物語が浮かび上がる。これがGravity Mapの時間軸を持つ強みだ
- **Kubernetes**：重力集中度0.8%。5,000人超のコミュニティに構造が分散

さらに興味深い発見があった。**言語ファミリーによって重力の集中度が4.8倍違う。**

| 言語カテゴリ | 重力集中度 | 構造の物理 |
|---|---|---|
| Go（反フレームワーク文化） | 16.4% | 少数のアーキテクトに集中 |
| Rust / Scala（表現力型） | 6.7% | 型システムが構造を分散 |
| Rails / Laravel（FW依存） | 5.1% | フレームワークが構造を吸収 |
| C / C++（システム） | 3.4% | 最も分散 |

ここで大事なことを一つ。**どの構造が「正しい」かという話ではない。**

esbuildの92.5%集中は「悪い設計」ではない——一人の天才が全体を把握できるスケールでは、それが最適解かもしれない。Kubernetesの0.8%分散も「分散しているから優れている」わけではない——5,000人規模では分散が必然であり、それ自体が設計判断の結果だ。

EISが観測しているのは**構造の物理法則**であって、良し悪しの判定ではない。望遠鏡は銀河の形を記述する。渦巻銀河が楕円銀河より「優れている」とは言わない。

### Top 50：OSS宇宙で最も明るい星たち

29プロジェクト全体から、**Gravity（構造的影響力）で上位50人のエンジニア**をランキングした。

> [OSS Gravity Map — Top 50 Engineers](https://machuz.github.io/eis/research/oss-gravity-map/analysis/top50.html)

Salvatore Sanfilippo（Redis）、Alexey Milovidov（ClickHouse）、Ritchie Vink（Polars）——彼らの重力はスケールを振り切る。しかしより注目すべき発見は、**世界が名前を知らない440人のエンジニア**だった。カンファレンスで登壇しない。Twitterのフォロワーも多くない。しかしコードベースの重力場を辿ると、そこにいた——静かにアーキテクチャを支え続けていた。彼らを**Hidden Architects（隠れたアーキテクト）**と呼ぶ。

**異なる宇宙の比較について。** Gravityは*各リポジトリ内での相対的なシグナル*であり、リポジトリ間の絶対比較ではない。eslintでのJosh GoldbergのGravity 100とKubernetesでのJordan LiggittのGravity 77.3は**異なる宇宙からの観測**であり、直接比較はできない。これはまさにEngineering Relativity（第8章）だ。

ただし、歪みはGravityの構成によってある程度軽減される。3つの軸——モジュール所有率、設計関与率、横断性——はいずれも**比率ベースのシグナル**であり、絶対量ではない。50モジュールのプロジェクトで80%を持つことと500モジュールのプロジェクトで80%を持つことは、同じIndispensabilityシグナルとして記録される。このランキングが捉えているのは*自分の宇宙の重力場を形作った人*であり、「最も大きな宇宙で働いた人」ではない。

各銀河で最も明るい星をマッピングしていると考えてほしい。銀河の大きさは異なるが、どの銀河でも重力場を形作る星は観測できる。

> 詳細な分析結果：[OSS Gravity Map](https://machuz.github.io/engineering-impact-score/research/oss-gravity-map/analysis/oss-gravity-map-ja.html)

---

## これは何で「ない」か

> *We don't measure engineers. We reveal how software actually works.*
> *——エンジニアを測っているのではない。ソフトウェアが実際にどう動いているかを明らかにしている。*

このシリーズでは「戦闘力」という言葉を使っている。ドラゴンボールから借りたキャッチーなメタファーだが、危険な含意がある——エンジニアを一つの軸で序列化できる、という誤解だ。

**できない。EISもそれを目指していない。**

じゃあ何を測っているのか。シンプルだ。**このコードベースで、どれだけ作ったか。どれだけ影響を与えたか。そして、作ったものがどれだけ残っているか。** それだけ。「エンジニアとしてどれだけ優秀か」ではなく、「このコード宇宙にどんな痕跡を残したか」。

本当の意味での「優秀さ」は、**複数の宇宙に残した痕跡**によって初めて定量化できる。一つのコードベースでのImpactはローカルな観測に過ぎない。異なるコードベース、異なるチーム、異なるドメインで一貫して高いImpactを残せるなら——それは再現性のある重力だ。一つの銀河で輝く星と、自然の力そのものの違い。

いくつか大事なことを書いておく。

**EISが観測しているのは「このコードベースでのインパクト」であって、エンジニアとしての絶対的な能力ではない。** Impact 40は「このコードベースにおいて、その人のコードが残り、設計に影響を与え、負債を掃除している」ということ。別のコードベースに移れば、Impactは逆転するかもしれない。（これをエンジニアリング相対性理論と呼んでいる。）

**コンテキストなしのシグナルは危険。** Survivalが低いのは設計が悪いからかもしれないし、レガシーコードを書き換えている最中（Rescueスタイル）だからかもしれない。設計の悪いコードベースでImpactが高いのは「誰もリファクタできないコード」を書いているだけかもしれない。常にコンテキストとセットで解釈すべきだ。

**コード以外の貢献はgitに残らない。** コードレビューの質、メンタリング、ドキュメンテーション、心理的安全性、ドメイン知識——これらは極めて重要だが `git log` には痕跡を残さない。EISはgitが記録するものだけを捉える。これだけでエンジニアを完全に評価しようとするのは、有害で間違っている。

**監視ツールではない。** EISは望遠鏡——すでに存在する構造を明らかにするものであり、序列を作るものではない。理解と改善のためではなく、ランク付けと罰のために使われるなら、それはEISの失敗だ。

**時間減衰がゲーミングを防ぐ。** 忙しそうに見せる作業ではシグナルは強くならない——数ヶ月後にコードベースに残っているコードだけがカウントされる。負債清掃軸があることで、他者に仕事を発生させてImpactを上げることも構造的にできない。

望遠鏡は星の明るさを測る。どの星が存在に値するかを決めるものではない。

---

## シリーズ目次

このブログシリーズ「git考古学」では、EISを使って実際のチームを分析し、何が見えたかを書いている。

1. **[履歴だけでエンジニアの「戦闘力」を定量化する](https://zenn.dev/machuz/books/git-archaeology/viewer/ch1)** — 7軸観測の全設計
2. **[エンジニアの「戦闘力」から、チームの「構造力」へ](https://zenn.dev/machuz/books/git-archaeology/viewer/ch2)**
3. **[Architectには流派がある：Git履歴が暴く進化の分岐モデル](https://zenn.dev/machuz/books/git-archaeology/viewer/ch3)**
4. **[Backend Architectは収束する：成仏という聖なる仕事](https://zenn.dev/machuz/books/git-archaeology/viewer/ch4)**
5. **[タイムライン：シグナルは嘘をつかないし、遠慮も映る](https://zenn.dev/machuz/books/git-archaeology/viewer/ch5)**
6. **[チームは進化する——タイムラインが暴く組織の法則](https://zenn.dev/machuz/books/git-archaeology/viewer/ch6)**
7. **[コードの宇宙を観測する](https://zenn.dev/machuz/books/git-archaeology/viewer/ch7)**
8. **[Engineering Relativity：なぜ同じエンジニアでもImpactが変わるのか](https://zenn.dev/machuz/books/git-archaeology/viewer/ch8)**
9. **[Origin：コード宇宙のビッグバン](https://zenn.dev/machuz/books/git-archaeology/viewer/ch9)**
10. **[Dark Matter：見えない重力](https://zenn.dev/machuz/books/git-archaeology/viewer/ch10)**
11. **[Entropy：宇宙は常に無秩序に向かう](https://zenn.dev/machuz/books/git-archaeology/viewer/ch11)**
12. **[Collapse：良いArchitectとBlack Hole Engineer](https://zenn.dev/machuz/books/git-archaeology/viewer/ch12)**
13. **[Cosmology of Code：コード宇宙論](https://zenn.dev/machuz/books/git-archaeology/viewer/ch13)**
14. **[Civilization：なぜ一部のコードベースだけが文明になるのか](https://zenn.dev/machuz/books/git-archaeology/viewer/ch14)**
15. **[AI Creates Stars, Not Gravity](https://zenn.dev/machuz/books/git-archaeology/viewer/ch15)**
16. **[The Engineers Who Shape Gravity：重力を作るエンジニアたち](https://zenn.dev/machuz/books/git-archaeology/viewer/ch16)**

---

## インストール

```bash
# Homebrew
brew tap machuz/tap && brew install eis

# Go
go install github.com/machuz/eis/cmd/eis@latest
```

**GitHub**: [eis](https://github.com/machuz/eis)

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/logo-full-zenn.png)
