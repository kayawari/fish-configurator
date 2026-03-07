# 要件定義書

## はじめに

fish shell専用のalias（エイリアス）およびabbr（略語）をインタラクティブに管理するCLIツールです。このツールは、fish shellの実行環境から動的に情報を取得し、コマンド履歴を解析して、ユーザーが効率的にエイリアスと略語を管理できるようにします。

## 用語集

- **System**: fish-configurator CLI ツール
- **Fish_Shell**: ユーザーが使用しているfish shellの実行環境
- **Alias**: fish shellで関数として実装されるコマンドの別名（`alias` コマンドで作成されるが、内部的には関数）
- **Abbr**: fish shellのabbrコマンドで定義される略語展開（abbreviation）
- **History_File**: `~/.local/share/fish/fish_history` に保存されているコマンド履歴ファイル
- **Config_Directory**: `~/.config/fish/conf.d/` ディレクトリ
- **Management_File**: Config_Directory内に生成される専用の設定ファイル（fish-configurator が管理）
- **Fzf**: コマンドラインのファジーファインダーツール（https://github.com/junegunn/fzf）

## 要件

### 要件 1: Alias/Abbr の一覧表示

**ユーザーストーリー:** ユーザーとして、現在定義されているaliasとabbrの一覧を確認したい。そうすることで、既存の設定を把握し、重複や不要な定義を見つけることができる。

#### 受入基準

1. WHEN ユーザーが一覧表示コマンドを実行する THEN THE System SHALL aliasとabbrのどちらを表示するか選択を求める
2. WHEN ユーザーがaliasを選択する THEN THE System SHALL `fish -c "alias"` コマンドを実行してaliasの一覧を取得する
3. WHEN ユーザーがabbrを選択する THEN THE System SHALL `fish -c "abbr"` コマンドを実行してabbreviationの一覧を取得する
4. WHEN 取得した情報を表示する THEN THE System SHALL 各エントリの名前と定義内容を読みやすい形式で出力する
5. IF Fish_Shellが利用できない THEN THE System SHALL エラーメッセージを表示して終了する
6. IF 選択した種類のエントリが存在しない THEN THE System SHALL 情報メッセージを表示する

### 要件 2: 新規Alias/Abbrの追加

**ユーザーストーリー:** ユーザーとして、新しいaliasまたはabbrを追加したい。そうすることで、よく使うコマンドを短縮形で実行できるようになる。

#### 受入基準

1. WHEN ユーザーが `add alias` サブコマンドを実行する THEN THE System SHALL aliasの名前入力を求める
2. WHEN ユーザーが `add abbr` サブコマンドを実行する THEN THE System SHALL abbrの名前入力を求める
3. WHEN ユーザーが名前を入力する THEN THE System SHALL 定義内容（値）の入力を求める
4. WHEN ユーザーが入力を完了する THEN THE System SHALL 入力内容の妥当性を検証する
5. WHEN 名前が空文字列または空白文字のみで構成される THEN THE System SHALL エラーメッセージを表示して追加を拒否する
6. WHEN 定義内容が空文字列または空白文字のみで構成される THEN THE System SHALL エラーメッセージを表示して追加を拒否する
7. WHEN 入力が妥当である THEN THE System SHALL fish shellのシンタックスチェックを実行する
8. WHEN シンタックスチェックを実行する THEN THE System SHALL 一時ファイルに定義を書き込み `fish -n <tempfile>` コマンドで検証する
9. IF シンタックスエラーが検出される THEN THE System SHALL エラー出力を表示して追加を中止する
10. WHEN シンタックスチェックが成功する THEN THE System SHALL Management_File に新しいエントリを追加する
11. WHEN Management_File に書き込む THEN THE System SHALL 既存の内容を保持したまま新しいエントリを追加する
12. WHEN 追加が成功する THEN THE System SHALL 成功メッセージを表示する

### 要件 3: Alias/Abbrの削除

**ユーザーストーリー:** ユーザーとして、不要になったaliasまたはabbrを削除したい。そうすることで、設定をクリーンに保つことができる。

#### 受入基準

1. WHEN ユーザーが `remove` サブコマンドを実行する THEN THE System SHALL aliasとabbrのどちらを削除するか選択を求める
2. WHEN ユーザーがaliasまたはabbrを選択する THEN THE System SHALL Management_File から該当する種類のエントリを解析して一覧表示する
3. WHEN 一覧を表示する THEN THE System SHALL ユーザーが削除対象を選択できるインタラクティブな方法を提供する
4. WHEN ユーザーが削除対象を選択する THEN THE System SHALL 確認メッセージを表示する
5. WHEN ユーザーが削除を確認する THEN THE System SHALL Management_File から該当エントリのみを削除する
6. WHEN Management_File から削除する THEN THE System SHALL 他のエントリを保持する
7. WHEN 削除が成功する THEN THE System SHALL 成功メッセージを表示する
8. IF 削除対象がManagement_Fileに存在しない THEN THE System SHALL 警告メッセージを表示する
9. IF Management_Fileに該当する種類のエントリが存在しない THEN THE System SHALL 情報メッセージを表示して終了する

### 要件 4: コマンド履歴からの選択機能

**ユーザーストーリー:** ユーザーとして、過去に実行したコマンドから直接aliasやabbrの定義を作成したい。そうすることで、よく使うコマンドを簡単に短縮形に変換できる。

#### 受入基準

1. WHEN ユーザーが `history` サブコマンドを実行する THEN THE System SHALL History_File を読み込む
2. WHEN History_File を読み込む THEN THE System SHALL `- cmd:` で始まる行を抽出してfzfに渡す
3. WHEN fzfにデータを渡す THEN THE System SHALL fzfプロセスを起動して標準入力経由でコマンド履歴を提供する
4. WHEN ユーザーがfzfでコマンドを選択する THEN THE System SHALL 選択されたコマンドから `- cmd: ` プレフィックスを除去する
5. WHEN コマンドが選択される THEN THE System SHALL aliasとabbrのどちらを作成するか選択を求める
6. WHEN ユーザーがaliasまたはabbrを選択する THEN THE System SHALL 名前の入力を求める
7. WHEN ユーザーが名前を入力する THEN THE System SHALL fish shellのシンタックスチェックを実行する
8. WHEN シンタックスチェックを実行する THEN THE System SHALL 一時ファイルに定義を書き込み `fish -n <tempfile>` コマンドで検証する
9. IF シンタックスエラーが検出される THEN THE System SHALL エラー出力を表示して追加を中止する
10. WHEN シンタックスチェックが成功する THEN THE System SHALL Management_File に新しいエントリを追加する
11. WHEN 追加が成功する THEN THE System SHALL 成功メッセージを表示する
12. IF History_File が存在しない THEN THE System SHALL エラーメッセージを表示して機能を終了する
13. IF fzfが利用できない THEN THE System SHALL エラーメッセージを表示して機能を終了する
14. IF fzfでコマンドが選択されなかった THEN THE System SHALL 処理を中止して終了する

### 要件 5: 設定ファイルの安全な管理

**ユーザーストーリー:** ユーザーとして、既存の `config.fish` を破壊せずにaliasとabbrを管理したい。そうすることで、手動で行った他の設定を保護できる。

#### 受入基準

1. THE System SHALL `config.fish` を直接編集しない
2. THE System SHALL Config_Directory 内に専用のManagement_File を作成する
3. WHEN Management_File を作成する THEN THE System SHALL ファイル名に `fish-configurator` を含める
4. WHEN Management_File に書き込む THEN THE System SHALL 有効なfish shellスクリプト構文を使用する
5. WHEN aliasを書き込む THEN THE System SHALL `alias <name> '<definition>'` 形式を使用する
6. WHEN abbrを書き込む THEN THE System SHALL `abbr -a <name> '<definition>'` 形式を使用する
7. IF Config_Directory が存在しない THEN THE System SHALL ディレクトリを作成する
8. WHEN ファイル操作を行う THEN THE System SHALL 適切なファイルパーミッションを設定する

### 要件 6: エラーハンドリングと検証

**ユーザーストーリー:** ユーザーとして、エラーが発生した場合に明確なメッセージを受け取りたい。そうすることで、問題を理解し、適切に対処できる。

#### 受入基準

1. WHEN ファイル操作が失敗する THEN THE System SHALL 具体的なエラーメッセージを表示する
2. WHEN Fish_Shell コマンドの実行が失敗する THEN THE System SHALL エラー内容を表示する
3. WHEN 無効な入力を受け取る THEN THE System SHALL どの入力が無効かを明示する
4. WHEN 重複する名前でエントリを追加しようとする THEN THE System SHALL 警告メッセージを表示して上書き確認を求める
5. WHEN システムエラーが発生する THEN THE System SHALL 適切な終了コードを返す
6. THE System SHALL すべてのエラーメッセージを標準エラー出力に出力する

### 要件 7: CLIインターフェース設計

**ユーザーストーリー:** ユーザーとして、直感的で使いやすいコマンドラインインターフェースを使用したい。そうすることで、学習コストを最小限に抑えて効率的に作業できる。

#### 受入基準

1. THE System SHALL サブコマンド形式のインターフェースを提供する
2. THE System SHALL `list`, `add`, `remove`, `history` の各サブコマンドをサポートする
3. THE System SHALL `add` サブコマンドに `alias` と `abbr` のサブサブコマンドをサポートする
4. WHEN ユーザーが `add alias` を実行する THEN THE System SHALL alias追加フローを開始する
5. WHEN ユーザーが `add abbr` を実行する THEN THE System SHALL abbr追加フローを開始する
6. WHEN ユーザーがヘルプを要求する THEN THE System SHALL 使用方法と利用可能なサブコマンドを表示する
7. WHEN 無効なサブコマンドが指定される THEN THE System SHALL エラーメッセージとヘルプ情報を表示する
8. WHEN インタラクティブな入力を求める THEN THE System SHALL 明確なプロンプトを表示する
9. WHEN 処理が完了する THEN THE System SHALL 結果を明確に伝えるメッセージを表示する
10. THE System SHALL 標準入力からの入力を受け付ける

### 要件 8: 外部依存の管理

**ユーザーストーリー:** 開発者として、必要最小限の外部依存で開発したい。そうすることで、ツールの保守性と移植性を高めることができる。

#### 受入基準

1. THE System SHALL Go言語の標準パッケージを優先的に使用する
2. THE System SHALL `os`, `flag`, `bufio`, `os/exec`, `text/template` などの標準パッケージを活用する
3. THE System SHALL `history` サブコマンドの実装にfzfを外部コマンドとして使用する
4. WHEN fzfを使用する THEN THE System SHALL `os/exec` パッケージを使用してfzfプロセスを起動する
5. WHEN YAMLを解析する必要がある THEN THE System SHALL 軽量な実装または標準パッケージで対応可能な方法を使用する
6. THE System SHALL Go言語の最新安定版でビルド可能である
7. THE System SHALL fzf以外の外部バイナリへの依存を避ける

### 要件 9: データの永続化

**ユーザーストーリー:** ユーザーとして、追加したaliasとabbrが永続的に保存されることを期待する。そうすることで、fish shellを再起動しても設定が維持される。

#### 受入基準

1. WHEN エントリを追加または削除する THEN THE System SHALL 変更を即座にManagement_File に反映する
2. WHEN Management_File に書き込む THEN THE System SHALL ファイルの整合性を保証する
3. WHEN 書き込みが完了する THEN THE System SHALL ファイルが正しく保存されたことを確認する
4. IF 書き込み中にエラーが発生する THEN THE System SHALL 元のファイル内容を保持する
5. THE System SHALL Management_File にコメントを含めて管理情報を記録する
6. WHEN Management_File を生成する THEN THE System SHALL ファイルの先頭に自動生成された旨のコメントを追加する
