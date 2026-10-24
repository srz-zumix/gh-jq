# instructions

指示が曖昧な場合は編集せずに曖昧な箇所を指摘してください。

ソースコード中のコメントは英語で記載

## README.md

* README は英語で記載します
* markdown の書式警告は修正してください
* Usage 例・説明文は日本語・英語の混在を避け、統一した言語で記載してください
* completion の詳細をヘルプに書かず、Installation の次のセクションに簡単な説明と詳細への案内を記載する
* コマンドの説明は下記コメントブロックの内容のように記述してください
* コマンドの Usage 記載時は、引数（expression, file...）やオプションの省略可否・デフォルト値も明記してください
* オプションの一覧は項目のタイトルでソートして記述します
* 排他関係にあるオプション（--save / --load / --clear など）や、優先されるオプション（--null-input は入力を読まない 等）は明記してください
* Usage 例のコマンド列は実際に動作する形式で記載してください（例: `gh api repos/{owner}/{repo} | gh jq '.name'`）
* コマンドの説明文（Long）は簡潔かつ具体的に、何ができるか・どんなオプションがあるかを明記してください
* コマンド・オプションの追加・削除・変更時は必ず README も更新し、実装と乖離させないでください

<!--
### コマンドの機能

```sh
コマンド列
```

コマンドの Long の説明
-->

## コーディング規約

* fmt.Errorf: error strings should not end with punctuation or newlines (ST1005) go-staticcheck
* gh extension 用開発の共通パッケージは github.com/srz-zumix/go-gh-extension/pkg/<path/to/dir> で import（例: completion）
* jq 式の評価は github.com/cli/go-gh/v2/pkg/jq を利用し、独自に jq 実装を持たない

### ソースコード全般

* ディレクトリ・ファイル構成は以下の責務分割に従うこと
  * main.go: cmd.Execute() を呼ぶのみ
  * cmd/: CLI コマンド定義・引数/フラグ処理・cobra.Command 生成のみを担当し、ビジネスロジックは持たない
    * root.go: ルートコマンド（`gh jq [expression] [file...]`）。入力ソース（stdin / ファイル / 保存データ / null）の選択と出力設定（TTY 判定によるインデント・色付け）を行い、internal/ のパッケージを呼び出す
    * completion.go: go-gh-extension の completion コマンドの登録
  * internal/: gh-jq 固有のロジック。外部から import されない
    * evaluate/: jq 評価前の入力前処理（--null-input, --slurp）と jq.EvaluateFormatted の呼び出し
    * store/: --save / --load / --clear で使用する JSON 入力の永続化
      * 保存先は os.UserConfigDir() 配下の gh-jq ディレクトリ。ディレクトリは 0700、ファイルは 0600 で作成する
      * Clear はファイルが存在しない場合もエラーにしない
  * version/: バージョン情報管理
* 新しい責務（ビジネスロジック）を追加する場合は cmd/ に置かず、internal/ 配下に責務ごとのパッケージを新設する
* コメントは英語で記載し、関数・構造体・パッケージの責務が明確になるよう記述する
* エラーはどういう操作をしようとしてエラーになったかメッセージに含め、%w でラップする
* テストコードは *_test.go で実装し、各責務ごとに配置する
  * cmd/ 配下にはテストを書きません。フラグのパースや排他制御は cobra/pflag が担保しているため、二重にテストしない
  * cmd/ にテストを書きたくなった場合は責務分割のシグナルです。検証したいロジックを internal/ 配下のパッケージに切り出し、そちらでテストする
  * ファイル入出力を伴うテストは t.TempDir() を使い、実ユーザーの設定ディレクトリを汚さないようにする。macOS の os.UserConfigDir() は XDG_CONFIG_HOME を参照しないため、環境変数ではなく関数変数の差し替え（internal/store の userConfigDir 等）で保存先を切り替える
* テストデータには実データ（ユーザー名・ID・ハッシュ・組織名・メールアドレス等）を使用せず、ユーザー提供の情報や実際の API レスポンスも転用しない。検証に必要な形式・制約を満たす架空の値を使用する
* コード重複は避け、共通処理は関数化・ユーティリティ化する
* Lint/Formatter（go fmt, go vet, golangci-lint 等）を通してからコミットする
* 依存関係の循環(import cycle)が発生しないよう注意する

### パッケージ詳細

#### cmd

* 新しいオプションやコマンドを作成する場合は既存の実装を参照し、書き方など踏襲してください
* オプションは基本的にパッケージ変数で受け取り、init() で登録します
* jq と同等の機能を持つオプションは jq と同じ long/short 名にしてください（例: --null-input / -n, --slurp / -s）
* 同時に指定できないオプションは MarkFlagsMutuallyExclusive で宣言してください
* RunE で処理を実装します
* Args で引数の検証をします
* フラグ値の妥当性判定や入力の前処理など、ライブラリが担保しない判断・処理は RunE に直接書かず関数に切り出してください
* サブコマンドを追加する場合は cobra.Command を return する New<コマンド名>Cmd 関数を定義し、親コマンドの .go ファイルで登録してください

## ビルド・テスト

* ビルドは `make build`、テストは `make test`、ローカルへのインストールは `make install` を使用します
* go-gh-extension を同時に修正する場合は `make go-work` で go.work を作成し、コミットしないでください（`make clean` で削除）

## GitHub Actions

* ワークフローで使用する action はコミット SHA で固定し、バージョンをコメントで併記してください（pinact）。ただし actions/* はメジャーバージョンタグ（例: @v6）で指定します（.pinact.yaml / zizmor.yaml で除外済み）
* permissions は最小権限とし、checkout では persist-credentials: false を指定してください（zizmor）
