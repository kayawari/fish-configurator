package history

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReadCommands_ValidHistoryFile tests reading commands from a valid history file
func TestReadCommands_ValidHistoryFile(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "fish-history-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// テスト用の履歴ファイルを作成
	historyPath := filepath.Join(tempDir, "fish_history")
	content := `- cmd: ls -la
  when: 1234567890
- cmd: git status
  when: 1234567891
- cmd: echo "hello world"
  when: 1234567892
`
	err = os.WriteFile(historyPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write history file: %v", err)
	}

	// HistoryReader を作成
	reader := NewFileHistoryReader(historyPath)

	// コマンドを読み込む
	commands, err := reader.ReadCommands()
	if err != nil {
		t.Fatalf("Failed to read commands: %v", err)
	}

	// 期待されるコマンド数を確認
	expectedCount := 3
	if len(commands) != expectedCount {
		t.Errorf("Expected %d commands, got %d", expectedCount, len(commands))
	}

	// 各コマンドの内容を確認
	expectedCommands := []string{
		"ls -la",
		"git status",
		`echo "hello world"`,
	}

	for i, expected := range expectedCommands {
		if i >= len(commands) {
			t.Errorf("Missing command at index %d", i)
			continue
		}
		if commands[i] != expected {
			t.Errorf("Command at index %d: expected %q, got %q", i, expected, commands[i])
		}
	}
}

// TestReadCommands_EmptyFile tests reading from an empty history file
func TestReadCommands_EmptyFile(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "fish-history-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 空の履歴ファイルを作成
	historyPath := filepath.Join(tempDir, "fish_history")
	err = os.WriteFile(historyPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to write history file: %v", err)
	}

	// HistoryReader を作成
	reader := NewFileHistoryReader(historyPath)

	// コマンドを読み込む
	commands, err := reader.ReadCommands()
	if err != nil {
		t.Fatalf("Failed to read commands: %v", err)
	}

	// 空のリストが返されることを確認
	if len(commands) != 0 {
		t.Errorf("Expected 0 commands, got %d", len(commands))
	}
}

// TestReadCommands_NoCommandLines tests reading from a file with no command lines
func TestReadCommands_NoCommandLines(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "fish-history-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// コマンド行を含まない履歴ファイルを作成
	historyPath := filepath.Join(tempDir, "fish_history")
	content := `- when: 1234567890
  paths:
    - /home/user
- when: 1234567891
  paths:
    - /tmp
`
	err = os.WriteFile(historyPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write history file: %v", err)
	}

	// HistoryReader を作成
	reader := NewFileHistoryReader(historyPath)

	// コマンドを読み込む
	commands, err := reader.ReadCommands()
	if err != nil {
		t.Fatalf("Failed to read commands: %v", err)
	}

	// 空のリストが返されることを確認
	if len(commands) != 0 {
		t.Errorf("Expected 0 commands, got %d", len(commands))
	}
}

// TestReadCommands_MixedContent tests reading from a file with mixed content
func TestReadCommands_MixedContent(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "fish-history-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 混在したコンテンツを持つ履歴ファイルを作成
	historyPath := filepath.Join(tempDir, "fish_history")
	content := `- cmd: first command
  when: 1234567890
  paths:
    - /home/user
- when: 1234567891
- cmd: second command
  when: 1234567892
- paths:
    - /tmp
- cmd: third command
`
	err = os.WriteFile(historyPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write history file: %v", err)
	}

	// HistoryReader を作成
	reader := NewFileHistoryReader(historyPath)

	// コマンドを読み込む
	commands, err := reader.ReadCommands()
	if err != nil {
		t.Fatalf("Failed to read commands: %v", err)
	}

	// 期待されるコマンド数を確認
	expectedCount := 3
	if len(commands) != expectedCount {
		t.Errorf("Expected %d commands, got %d", expectedCount, len(commands))
	}

	// 各コマンドの内容を確認
	expectedCommands := []string{
		"first command",
		"second command",
		"third command",
	}

	for i, expected := range expectedCommands {
		if i >= len(commands) {
			t.Errorf("Missing command at index %d", i)
			continue
		}
		if commands[i] != expected {
			t.Errorf("Command at index %d: expected %q, got %q", i, expected, commands[i])
		}
	}
}

// TestReadCommands_FileNotFound tests error handling when file doesn't exist
func TestReadCommands_FileNotFound(t *testing.T) {
	// 存在しないファイルパスを指定
	historyPath := "/nonexistent/path/fish_history"

	// HistoryReader を作成
	reader := NewFileHistoryReader(historyPath)

	// コマンドを読み込む（エラーが返されることを期待）
	_, err := reader.ReadCommands()
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestReadCommands_CommandsWithSpecialCharacters tests commands with special characters
func TestReadCommands_CommandsWithSpecialCharacters(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "fish-history-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 特殊文字を含むコマンドを持つ履歴ファイルを作成
	historyPath := filepath.Join(tempDir, "fish_history")
	content := `- cmd: echo "hello world"
- cmd: grep -r "pattern" /path/to/dir
- cmd: sed 's/old/new/g' file.txt
- cmd: awk '{print $1}' data.txt
- cmd: find . -name "*.go" -type f
`
	err = os.WriteFile(historyPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write history file: %v", err)
	}

	// HistoryReader を作成
	reader := NewFileHistoryReader(historyPath)

	// コマンドを読み込む
	commands, err := reader.ReadCommands()
	if err != nil {
		t.Fatalf("Failed to read commands: %v", err)
	}

	// 期待されるコマンド数を確認
	expectedCount := 5
	if len(commands) != expectedCount {
		t.Errorf("Expected %d commands, got %d", expectedCount, len(commands))
	}

	// 各コマンドの内容を確認
	expectedCommands := []string{
		`echo "hello world"`,
		`grep -r "pattern" /path/to/dir`,
		`sed 's/old/new/g' file.txt`,
		`awk '{print $1}' data.txt`,
		`find . -name "*.go" -type f`,
	}

	for i, expected := range expectedCommands {
		if i >= len(commands) {
			t.Errorf("Missing command at index %d", i)
			continue
		}
		if commands[i] != expected {
			t.Errorf("Command at index %d: expected %q, got %q", i, expected, commands[i])
		}
	}
}

// TestNewFileHistoryReader_DefaultPath tests default path resolution
func TestNewFileHistoryReader_DefaultPath(t *testing.T) {
	// デフォルトパスで HistoryReader を作成
	reader := NewFileHistoryReader("")

	// historyPath が設定されていることを確認
	if reader.historyPath == "" {
		t.Error("Expected historyPath to be set, got empty string")
	}

	// パスに fish_history が含まれることを確認
	if !contains(reader.historyPath, "fish_history") {
		t.Errorf("Expected historyPath to contain 'fish_history', got %q", reader.historyPath)
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
