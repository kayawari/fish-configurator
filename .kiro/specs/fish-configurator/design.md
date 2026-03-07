# 設計書

## 概要

fish-configuratorは、fish shell専用のalias（エイリアス）およびabbr（略語）を管理するCLIツールです。Go言語で実装され、標準パッケージを最大限活用し、外部依存を最小限に抑えます。ユーザーはインタラクティブなインターフェースを通じて、aliasとabbrの追加、削除、一覧表示、およびコマンド履歴からの選択を行うことができます。

## アーキテクチャ

### 全体構成

```
fish-configurator/
├── main.go              # エントリーポイント、CLIルーティング
├── cmd/
│   ├── list.go         # 一覧表示コマンド
│   ├── add.go          # 追加コマンド
│   ├── remove.go       # 削除コマンド
│   └── history.go      # 履歴選択コマンド
├── internal/
│   ├── config/
│   │   ├── manager.go  # 設定ファイル管理
│   │   └── parser.go   # 設定ファイル解析
│   ├── fish/
│   │   ├── executor.go # fish コマンド実行
│   │   └── validator.go # シンタックスチェック
│   ├── history/
│   │   └── reader.go   # 履歴ファイル読み込み
│   └── ui/
│       └── prompt.go   # ユーザー入力処理
└── go.mod
```

### 設計原則

1. **単一責任の原則**: 各パッケージは明確に定義された単一の責任を持つ
2. **依存性の逆転**: 具体的な実装ではなく、インターフェースに依存する
3. **標準パッケージ優先**: 外部ライブラリへの依存を最小限に抑える
4. **エラーハンドリング**: すべてのエラーを適切に処理し、ユーザーに明確なメッセージを提供する

## コンポーネントとインターフェース

### 1. CLIルーター（main.go）

**責務**: コマンドライン引数を解析し、適切なサブコマンドハンドラーにルーティングする

**インターフェース**:
```go
type Command interface {
    Execute(args []string) error
}
```

**実装**:
- `flag` パッケージを使用してサブコマンドを解析
- サブコマンドに応じて適切なハンドラーを呼び出す
- エラーハンドリングと終了コード管理

### 2. 設定ファイル管理（config/manager.go）

**責務**: Management_Fileの読み書き、エントリの追加・削除

**インターフェース**:
```go
type ConfigManager interface {
    Load() (*Config, error)
    Save(config *Config) error
    AddEntry(entryType, name, definition string) error
    RemoveEntry(entryType, name string) error
    ListEntries(entryType string) ([]Entry, error)
}

type Entry struct {
    Type       string // "alias" or "abbr"
    Name       string
    Definition string
}

type Config struct {
    Entries []Entry
}
```

**実装詳細**:
- Management_Fileのパス: `~/.config/fish/conf.d/fish-configurator.fish`
- ファイルが存在しない場合は自動的に新規作成する
- Config_Directoryが存在しない場合は、ディレクトリも自動的に作成する（要件5.7）
- ファイルの先頭に自動生成コメントを追加
- エントリの形式:
  - alias: `alias <name> '<definition>'`
  - abbr: `abbr -a <name> '<definition>'`

### 3. 設定ファイル解析（config/parser.go）

**責務**: Management_Fileを解析してエントリを抽出

**インターフェース**:
```go
type Parser interface {
    Parse(content string) ([]Entry, error)
}
```

**実装詳細**:
- 正規表現を使用してaliasとabbrの行を識別
- alias パターン: `^alias\s+(\S+)\s+'([^']+)'`
- abbr パターン: `^abbr\s+-a\s+(\S+)\s+'([^']+)'`
- コメント行と空行をスキップ

### 4. Fish コマンド実行（fish/executor.go）

**責務**: fish shellコマンドを実行し、結果を取得

**インターフェース**:
```go
type Executor interface {
    ExecuteCommand(command string) (string, error)
    CheckAvailability() error
}
```

**実装詳細**:
- `os/exec` パッケージを使用
- `fish -c "<command>"` 形式でコマンドを実行
- 標準出力と標準エラー出力をキャプチャ
- タイムアウト処理（5秒）

### 5. シンタックスチェック（fish/validator.go）

**責務**: fish shellのシンタックスチェックを実行

**インターフェース**:
```go
type Validator interface {
    ValidateSyntax(entryType, name, definition string) error
}
```

**実装詳細**:
- 一時ファイルを作成（`os.CreateTemp`）
- 一時ファイルに定義を書き込む
- `fish -n <tempfile>` を実行してシンタックスチェック
- 一時ファイルを削除（defer）
- エラーがあればエラーメッセージを返す

### 6. 履歴ファイル読み込み（history/reader.go）

**責務**: fish_historyファイルを読み込み、コマンド行を抽出

**インターフェース**:
```go
type HistoryReader interface {
    ReadCommands() ([]string, error)
}
```

**実装詳細**:
- History_Fileのパス: `~/.local/share/fish/fish_history`
- `bufio.Scanner` を使用して行単位で読み込み
- `- cmd:` で始まる行を抽出
- プレフィックス `- cmd: ` を除去してコマンド文字列を取得

### 7. ユーザー入力処理（ui/prompt.go）

**責務**: ユーザーからの入力を受け取り、検証する

**インターフェース**:
```go
type Prompter interface {
    PromptString(message string) (string, error)
    PromptChoice(message string, choices []string) (string, error)
    PromptConfirm(message string) (bool, error)
}
```

**実装詳細**:
- `bufio.Reader` を使用して標準入力から読み込み
- 入力の前後の空白を除去（`strings.TrimSpace`）
- 空白のみの入力を検証
- 選択肢の検証

### 8. Fzf統合（history/fzf.go）

**責務**: fzfプロセスを起動し、ユーザーの選択を取得

**インターフェース**:
```go
type FzfSelector interface {
    Select(items []string) (string, error)
    CheckAvailability() error
}
```

**実装詳細**:
- `os/exec` パッケージを使用してfzfプロセスを起動
- 標準入力経由でアイテムを渡す
- 標準出力から選択結果を取得
- fzfが利用できない場合はエラーを返す
- ユーザーがキャンセルした場合（終了コード130）を処理

## データモデル

### Entry

```go
type Entry struct {
    Type       string // "alias" or "abbr"
    Name       string // エントリの名前
    Definition string // エントリの定義内容
}
```

### Config

```go
type Config struct {
    Entries []Entry // すべてのエントリ
}
```

### Management_Fileフォーマット

```fish
# このファイルは fish-configurator によって自動生成されます
# 手動で編集しないでください

# Aliases
alias ll 'ls -la'
alias gs 'git status'

# Abbreviations
abbr -a gco 'git checkout'
abbr -a gp 'git push'
```

## 正確性プロパティ

プロパティとは、システムのすべての有効な実行において真であるべき特性や動作のことです。プロパティは、人間が読める仕様と機械で検証可能な正確性保証の橋渡しとなります。これらのプロパティは、ユニットテストとゴールデンテストで検証されます。


### プロパティ 1: 入力検証の一貫性

*任意の* 名前または定義内容の入力に対して、それが空文字列または空白文字のみで構成される場合、システムはエラーメッセージを表示して操作を拒否する

**検証: 要件 2.5, 2.6**

### プロパティ 2: エントリ追加の正確性

*任意の* 有効なエントリ（種類、名前、定義）に対して、シンタックスチェックが成功した場合、そのエントリをManagement_Fileに追加すると、ファイルを再読み込みした際にそのエントリが存在する

**検証: 要件 2.10, 4.10**

### プロパティ 3: 既存エントリの不変性

*任意の* Management_Fileと新しいエントリに対して、エントリを追加または削除する操作を行った場合、操作対象以外のすべてのエントリは変更されない

**検証: 要件 2.11, 3.6**

### プロパティ 4: エントリ削除の正確性

*任意の* Management_Fileと削除対象エントリに対して、そのエントリを削除すると、ファイルを再読み込みした際にそのエントリが存在しない

**検証: 要件 3.5**

### プロパティ 5: 履歴コマンド抽出の正確性

*任意の* History_Fileに対して、`- cmd:` で始まるすべての行を抽出し、プレフィックスを除去した結果は、元のファイルに含まれるすべてのコマンドと一致する

**検証: 要件 4.2**

### プロパティ 6: プレフィックス除去の正確性

*任意の* `- cmd: ` プレフィックスを持つ文字列に対して、プレフィックスを除去した結果は、元の文字列からプレフィックス部分を除いた残りの文字列と等しい

**検証: 要件 4.4**

### プロパティ 7: フォーマット正確性

*任意の* エントリ（種類、名前、定義）に対して、Management_Fileに書き込む際、aliasの場合は `alias <name> '<definition>'` 形式、abbrの場合は `abbr -a <name> '<definition>'` 形式を使用する

**検証: 要件 5.5, 5.6**

### プロパティ 8: シンタックス検証の正確性

*任意の* エントリ（種類、名前、定義）に対して、Management_Fileに書き込まれた内容は、fish shellのシンタックスチェック（`fish -n`）を通過する

**検証: 要件 5.4**

### プロパティ 9: エラー出力の一貫性

*任意の* エラー条件に対して、システムはすべてのエラーメッセージを標準エラー出力（stderr）に出力する

**検証: 要件 6.6**

### プロパティ 10: ファイル永続化の正確性

*任意の* エントリ追加または削除操作に対して、操作が成功した場合、変更は即座にManagement_Fileに反映され、ファイルの整合性が保証される

**検証: 要件 9.1, 9.2**

### プロパティ 11: ファイル保存の検証

*任意の* Management_Fileへの書き込み操作に対して、書き込みが完了した後、ファイルが正しく保存されたことを確認できる（ファイルが存在し、読み込み可能である）

**検証: 要件 9.3**

## エラーハンドリング

### エラーの種類

1. **ファイルシステムエラー**
   - Management_Fileの読み書き失敗
   - Config_Directoryの作成失敗
   - 一時ファイルの作成失敗
   - 対応: 具体的なエラーメッセージを標準エラー出力に表示し、適切な終了コードで終了

2. **Fish Shellエラー**
   - fish コマンドが利用できない
   - シンタックスチェック失敗
   - alias/abbr コマンドの実行失敗
   - 対応: エラー内容を標準エラー出力に表示し、適切な終了コードで終了

3. **入力検証エラー**
   - 空白のみの名前または定義
   - 無効なサブコマンド
   - 対応: どの入力が無効かを明示し、標準エラー出力に表示

4. **外部コマンドエラー**
   - fzfが利用できない
   - fzfでのキャンセル（終了コード130）
   - 対応: エラーメッセージを標準エラー出力に表示し、適切な終了コードで終了

5. **履歴ファイルエラー**
   - History_Fileが存在しない
   - History_Fileの読み込み失敗
   - 対応: エラーメッセージを標準エラー出力に表示し、適切な終了コードで終了

### 終了コード

- `0`: 成功
- `1`: 一般的なエラー
- `2`: 入力検証エラー
- `3`: ファイルシステムエラー
- `4`: 外部コマンドエラー

### エラーメッセージのフォーマット

すべてのエラーメッセージは以下の形式で標準エラー出力に出力されます：

```
Error: <エラーの種類>: <具体的な説明>
```

例：
```
Error: File System: Management_Fileの読み込みに失敗しました: permission denied
Error: Validation: 名前は空白のみで構成できません
Error: Fish Shell: fish コマンドが見つかりません
```

## テスト戦略

### テストアプローチ

このプロジェクトでは、ユニットテストとゴールデンテストを使用します：

- **ユニットテスト**: 特定の例、エッジケース、エラー条件、入力値の検証を行う
- **ゴールデンテスト**: 出力結果がコード変更後も担保されているかをチェックする

### ユニットテスト

ユニットテストは以下に焦点を当てます：

1. **入力値の検証**:
   - 空文字列の検証
   - 空白のみの文字列の検証
   - 有効な入力の検証
   - 特殊文字を含む入力の検証

2. **特定の例**:
   - 特定のサブコマンドの実行
   - 特定のフォーマットの検証
   - 特定のエラー条件

3. **エッジケース**:
   - fish shellが利用できない場合
   - fzfが利用できない場合
   - Management_Fileが存在しない場合
   - 空のManagement_File
   - シンタックスエラーのあるエントリ

4. **統合ポイント**:
   - fish コマンドの実行
   - fzfプロセスの起動
   - ファイルの読み書き

**ユニットテストの例**:

```go
func TestAddAlias_ValidInput(t *testing.T) {
    manager := NewConfigManager()
    
    err := manager.AddEntry("alias", "ll", "ls -la")
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    
    config, err := manager.Load()
    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }
    
    found := false
    for _, entry := range config.Entries {
        if entry.Name == "ll" && entry.Definition == "ls -la" {
            found = true
            break
        }
    }
    
    if !found {
        t.Error("Expected entry not found in config")
    }
}

func TestAddAlias_EmptyName(t *testing.T) {
    manager := NewConfigManager()
    
    err := manager.AddEntry("alias", "", "ls -la")
    if err == nil {
        t.Error("Expected error for empty name, got nil")
    }
}

func TestAddAlias_WhitespaceOnlyName(t *testing.T) {
    manager := NewConfigManager()
    
    testCases := []string{
        " ",
        "  ",
        "\t",
        "\n",
        "   \t  ",
    }
    
    for _, tc := range testCases {
        err := manager.AddEntry("alias", tc, "ls -la")
        if err == nil {
            t.Errorf("Expected error for whitespace-only name %q, got nil", tc)
        }
    }
}

func TestRemoveEntry_PreservesOthers(t *testing.T) {
    manager := NewConfigManager()
    
    // 複数のエントリを追加
    manager.AddEntry("alias", "ll", "ls -la")
    manager.AddEntry("alias", "gs", "git status")
    manager.AddEntry("abbr", "gco", "git checkout")
    
    // 1つを削除
    err := manager.RemoveEntry("alias", "ll")
    if err != nil {
        t.Fatalf("Failed to remove entry: %v", err)
    }
    
    // 他のエントリが保持されていることを確認
    config, err := manager.Load()
    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }
    
    if len(config.Entries) != 2 {
        t.Errorf("Expected 2 entries, got %d", len(config.Entries))
    }
    
    for _, entry := range config.Entries {
        if entry.Name == "ll" {
            t.Error("Deleted entry still exists")
        }
    }
}
```

### ゴールデンテスト

ゴールデンテストは、出力結果を期待値ファイル（ゴールデンファイル）と比較して、コード変更後も出力が一貫していることを確認します。

**ゴールデンテストの対象**:
- Management_Fileの生成内容
- エラーメッセージの出力
- ヘルプメッセージの出力
- 一覧表示の出力フォーマット

**ゴールデンテストの実装**:

```go
func TestGenerateManagementFile_Golden(t *testing.T) {
    manager := NewConfigManager()
    
    // テストデータを追加
    manager.AddEntry("alias", "ll", "ls -la")
    manager.AddEntry("alias", "gs", "git status")
    manager.AddEntry("abbr", "gco", "git checkout")
    manager.AddEntry("abbr", "gp", "git push")
    
    // Management_Fileの内容を取得
    content, err := manager.GetFileContent()
    if err != nil {
        t.Fatalf("Failed to get file content: %v", err)
    }
    
    // ゴールデンファイルのパス
    goldenFile := "testdata/management_file.golden"
    
    // ゴールデンファイルを読み込む
    expected, err := os.ReadFile(goldenFile)
    if err != nil {
        t.Fatalf("Failed to read golden file: %v", err)
    }
    
    // 内容を比較
    if content != string(expected) {
        t.Errorf("Output does not match golden file.\nGot:\n%s\n\nExpected:\n%s", content, expected)
    }
}

func TestListCommand_Output_Golden(t *testing.T) {
    // テスト用のManagement_Fileを準備
    setupTestConfig(t)
    
    // listコマンドを実行して出力をキャプチャ
    var buf bytes.Buffer
    cmd := NewListCommand(&buf)
    err := cmd.Execute([]string{"alias"})
    if err != nil {
        t.Fatalf("Failed to execute list command: %v", err)
    }
    
    output := buf.String()
    
    // ゴールデンファイルのパス
    goldenFile := "testdata/list_output.golden"
    
    // ゴールデンファイルを読み込む
    expected, err := os.ReadFile(goldenFile)
    if err != nil {
        t.Fatalf("Failed to read golden file: %v", err)
    }
    
    // 内容を比較
    if output != string(expected) {
        t.Errorf("Output does not match golden file.\nGot:\n%s\n\nExpected:\n%s", output, expected)
    }
}

func TestErrorMessages_Golden(t *testing.T) {
    testCases := []struct {
        name     string
        testFunc func() string
        golden   string
    }{
        {
            name: "empty_name_error",
            testFunc: func() string {
                manager := NewConfigManager()
                err := manager.AddEntry("alias", "", "ls -la")
                return err.Error()
            },
            golden: "testdata/error_empty_name.golden",
        },
        {
            name: "whitespace_name_error",
            testFunc: func() string {
                manager := NewConfigManager()
                err := manager.AddEntry("alias", "   ", "ls -la")
                return err.Error()
            },
            golden: "testdata/error_whitespace_name.golden",
        },
        {
            name: "syntax_error",
            testFunc: func() string {
                manager := NewConfigManager()
                err := manager.AddEntry("alias", "test", "invalid syntax '")
                return err.Error()
            },
            golden: "testdata/error_syntax.golden",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            output := tc.testFunc()
            
            // ゴールデンファイルを読み込む
            expected, err := os.ReadFile(tc.golden)
            if err != nil {
                t.Fatalf("Failed to read golden file: %v", err)
            }
            
            // 内容を比較
            if output != string(expected) {
                t.Errorf("Output does not match golden file.\nGot:\n%s\n\nExpected:\n%s", output, expected)
            }
        })
    }
}
```

### テストカバレッジ目標

- 全体のコードカバレッジ: 80%以上
- 重要なビジネスロジック: 90%以上
- エラーハンドリングパス: 100%

### モックとスタブ

外部依存（fish shell、fzf）のテストには、インターフェースベースのモックを使用します：

```go
type FishExecutor interface {
    ExecuteCommand(command string) (string, error)
    CheckAvailability() error
}

type MockFishExecutor struct {
    ExecuteFunc func(command string) (string, error)
    CheckFunc   func() error
}

func (m *MockFishExecutor) ExecuteCommand(command string) (string, error) {
    if m.ExecuteFunc != nil {
        return m.ExecuteFunc(command)
    }
    return "", nil
}

func (m *MockFishExecutor) CheckAvailability() error {
    if m.CheckFunc != nil {
        return m.CheckFunc()
    }
    return nil
}
```

## 実装の注意事項

### パフォーマンス

- ファイル操作は最小限に抑える
- 大きなHistory_Fileの読み込みは行単位で処理（`bufio.Scanner`）
- fzfプロセスはストリーミングでデータを渡す

### セキュリティ

- ユーザー入力は常に検証する
- シェルインジェクションを防ぐため、コマンド引数は適切にエスケープする
- 一時ファイルは適切なパーミッション（0600）で作成する
- 一時ファイルは必ず削除する（defer）

### 保守性

- 各パッケージは明確に定義された責任を持つ
- インターフェースを使用して依存性を管理する
- エラーメッセージは明確で具体的にする
- コードコメントは日本語で記述する

### 拡張性

- 新しいサブコマンドの追加が容易
- 新しいエントリタイプ（aliasとabbr以外）のサポートが可能
- 異なる設定ファイルフォーマットへの対応が可能
