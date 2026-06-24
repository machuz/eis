# OrbitLens Ace の中を歩く —— 天文台を一画面ずつ

![Cover](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/hatena/cover-ace-features.png?v=1)

*どの画面も、たった一つの問いに答える。ここでは、何が観えるようになるのか。*

---

[OrbitLens Ace](https://ace.orbitlens.io) は、OSSの git 望遠鏡である EIS の上に立つ天文台だ。望遠鏡が観測する。7軸、3軸トポロジー、JSON として。天文台は、その観測を、目に見える構造として読み返す。

これは見て回る記事だ。以下の各画面は、**公開OSSリポジトリ**を観測している様子で見せている。映っている構造は、誰でも読めるコードのものだ。それぞれについて、問いはひとつだけ。*何が観えるようになるのか。*

---

## 観測所ダッシュボード —— 7軸を一望する

![観測所ダッシュボード — 7軸シグナル（公開OSSリポの観測例）](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ace-observatory.png?v=1)

ダッシュボードは、光が最初に落ちる場所だ。観測されたすべてのエンジニアが、同じ7軸の上に並ぶ。Production、Catalysis、Survival、Design、Breadth、Debt Cleanup、Indispensability。貢献の「量」だけでなく、その「かたち」が一目で見える。

**何が観えるか:** 構造を形作った人と、量を生産した人の違いだ。1,890コミットで Survival がほぼゼロのエンジニアは、コミットグラフでは忙しく見えて、ここでは静かに見える。82コミットでも Design が天井に張りつくエンジニアなら、際立って見える。コミット数が均してしまう違いを、軸が保ち続ける。

---

## Star Detail —— レーダー、インサイト、構造的サマリー

![Star Detail — レーダー + 構造的サマリー（公開OSSリポの観測例）](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ace-star-detail.png?v=1)

望遠鏡を一つの星に向けると、7軸がレーダーに開く。その周りに、エンジニアのトポロジー分類（Role・Style・State）と、散文で書かれた **Structural Summary** が並ぶ。

立ち止まる価値があるのはサマリーだ。コンテキストのない数字は誤読を招く。Survival が低いのは設計が弱いからかもしれないし、レガシーコードを書き換えている最中だからかもしれない。サマリーはシグナルの場を丸ごと読み、実際に何が立っているかを描く。光が言葉になる。

**何が観えるか:** 「この人がどれだけ優秀か」ではなく、*このコード宇宙にどんな痕跡を残したか*だ。整合性を守る Cleaner は、量を生産する Producer とは違って読める。レーダーは、その違いを一つの数字に潰さず、そのまま見せる。

---

## モジュールトポロジー + 崩壊リスク —— 構造のどこが壊れかけているか

![モジュールトポロジー — 崩壊リスクとバスファクター（公開OSSリポの観測例）](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ace-module-topology.png?v=1)

望遠鏡は人だけでなく、人が存在する空間も観測する。すべてのモジュールが3つの軸の上に乗る。Coupling（境界品質）、Vitality（変更圧×生存力）、Ownership（知識分布）。

Ace は危険な組み合わせを読む。変更圧が高いのにコードが生き残らないモジュールは、構造的な時限爆弾だ。誰も触らないから残っているだけのモジュールは **Fragile** な要塞で、誰かが変えなければならない日まではきれいに見える。オーナーが去ったモジュールは **Orphaned**。バスファクターがすでにゼロになっている。

**何が観えるか:** システムのどこが壊れかけているか、それが危機になる前に。「このエンジニアが弱い」ではなく、「このモジュールはバスファクターが1で、それを抱えていた人はもういない」だ。

---

## 組織年表 —— コードベースが何をくぐってきたか

![組織年表 — 時系列の構造的イベント（公開OSSリポの観測例）](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ace-chronicle.png?v=1)

組織年表は Ace の芯であり、意図して**得点表ではない**。構造的なイベントを時間軸に沿って記録する。コードベースが生き延びたマイグレーション、ある subsystem を形作って去ったアーキテクト、オーナーが変わって Fragile になったモジュール。

記録するのは、コードベースが何をくぐってきたかであって、各人がどれだけ優秀かではない。この区別がすべてだ。得点表は人がゲームの仕方を覚えるもので、年表はチームが愛着を持つものだ。スコアは、近くで観たければそこにある。けれどそれは手に取るレンズであって、決して見出しにはならない。

**何が観えるか:** チーム自身の歴史が、見覚えを持てるくらいはっきりと書かれている。*評価ではなく観測。*

### Slack コネクタ —— `:orbitlens_chronicle:` リアクション

![Slack コネクタ — 年表リアクション（公開OSSリポの観測例）](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ace-slack-connector.png?v=1)

Ace だけが書く年表は、チームが実際に抱えている記録より薄くなる。だから年表にはコネクタがある。Slack のメッセージに `:orbitlens_chronicle:` でリアクションすると、その瞬間が時間軸の上に置かれる。つらいマイグレーションがついに着地した日、あるスレッドが決着させた判断。git は構造を記録する。コネクタは、git の届かない dark matter にチームが注釈を入れることを許す。

### 週次ダイジェスト —— コードベースを、週に一度

![週次ダイジェスト — その週の構造的イベント（公開OSSリポの観測例）](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ace-digest.png?v=1)

週に一度、Ace は直近の観測を読み、その週の構造的イベントを年表に置く。Fragile に踏み込んだモジュール、静かになったオーナー、ずれた生存比率。活動報告ではない。構造の中で何が変わったかの記録だ。

**何が観えるか:** コードベースのゆっくりした地殻変動を、チームが実際に頭に入れておけるリズムで。

---

## Ambient —— コードベースを、常に視界に

![Ambient モード — 常時表示（公開OSSリポの観測例）](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ace-ambient.png?v=1)

Ambient は、画面に立ち上げっぱなしにした天文台だ。朝会のための、あるいはオフィスの壁のための常時表示。構造を静かに視界に保つ。圧のかかっているモジュール、最近の年表エントリ、チームの重力場のかたち。調べるために開くダッシュボードではなく、ふと目をやって方向感覚を保つための空だ。

**何が観えるか:** 四半期に一度監査するものではなく、チームが隣で暮らす構造として。

---

## Gravity Certificate —— 旅をする痕跡

![Gravity Certificate — 構造的インパクトの可搬な記録（公開OSSリポの観測例）](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ace-certificate.png?v=1)

エンジニアの構造的インパクトは、コードベースの git 履歴の中に住む。そしてその人が去っても、そこに残り続ける。Gravity Certificate は、その観測を本人とともに旅させる。抱えていた重力、所有していたモジュール、形作ったアーキテクチャの記録を、履歴書で主張するのではなく git から観測したものとして。

これは慎重に設計してある。*コードが示すもの*の記録であり、一つの宇宙から観測したものだ。エンジニアリング能力の普遍的なランキングではない。一つのコードベースでの高い重力は、ローカルな観測にすぎない。証明書はまさにそう言うし、それ以上は言わない。

**何が観えるか:** ふだん見えないままになる、静かな構造の仕事だ。システムを支えていて、コミット数のリーダーボードには決して載らない種類の仕事。

---

## 全体の弧

望遠鏡が観測し、天文台が読む。ダッシュボードから証明書まで、貫く線は変わらない。**評価ではなく観測**だ。Ace はチームに、自分のコードベースが何をくぐってきたか、そして構造がどこで撓んでいるかを見せる。そしてスコアをレンズへと格下げする。記録が、ゲームの仕方を覚えるものではなく、チームが観たくなるものであり続けるように。

自分で望遠鏡を向ける:

```bash
brew install machuz/tap/eis
```

あるいは天文台を開く: [ace.orbitlens.io](https://ace.orbitlens.io)

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-full.png)

**GitHub**: [eis](https://github.com/machuz/eis) · **天文台**: [ace.orbitlens.io](https://ace.orbitlens.io) · **Library**: [library.orbitlens.io](https://library.orbitlens.io)

望遠鏡は無料で、オープンソース。ずっと。
