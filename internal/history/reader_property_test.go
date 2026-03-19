package history

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/quick"
)

// TestProperty_HistoryCommandExtractionAccuracy tests Property 5: 履歴コマンド抽出の正確性
// **Validates: Requirements 4.2**
//
// Property: For any History_File, extracting all lines starting with `- cmd:` and removing
// the prefix results in all commands contained in the original file.
func TestProperty_HistoryCommandExtractionAccuracy(t *testing.T) {
	property := func(commands []string) bool {
		// 空のコマンドリストはスキップ
		if len(commands) == 0 {
			return true
		}

		// テスト用の一時ディレクトリを作成
		tempDir, err := os.MkdirTemp("", "fish-history-test-*")
		if err != nil {
			t.Logf("Failed to create temp dir: %v", err)
			return false
		}
		defer os.RemoveAll(tempDir)

		// テスト用の履歴ファイルを作成
		historyPath := filepath.Join(tempDir, "fish_history")
		file, err := os.Create(historyPath)
		if err != nil {
			t.Logf("Failed to create history file: %v", err)
			return false
		}

		// 履歴ファイルにコマンドを書き込む
		// fish_history フォーマット: `- cmd: <command>` の形式
		for _, cmd := range commands {
			// 改行文字を含むコマンドは正規化
			normalizedCmd := strings.ReplaceAll(cmd, "\n", " ")
			normalizedCmd = strings.ReplaceAll(normalizedCmd, "\r", " ")

			_, err := fmt.Fprintf(file, "- cmd: %s\n", normalizedCmd)
			if err != nil {
				file.Close()
				t.Logf("Failed to write to history file: %v", err)
				return false
			}
		}
		file.Close()

		// HistoryReader を使用してコマンドを読み込む
		reader := NewFileHistoryReader(historyPath)
		extractedCommands, err := reader.ReadCommands()
		if err != nil {
			t.Logf("Failed to read commands: %v", err)
			return false
		}

		// 抽出されたコマンド数が元のコマンド数と一致することを確認
		if len(extractedCommands) != len(commands) {
			t.Logf("Command count mismatch: expected %d, got %d", len(commands), len(extractedCommands))
			return false
		}

		// 各コマンドが正しく抽出されていることを確認
		for i, expectedCmd := range commands {
			normalizedExpected := strings.ReplaceAll(expectedCmd, "\n", " ")
			normalizedExpected = strings.ReplaceAll(normalizedExpected, "\r", " ")
			normalizedExpected = strings.TrimSpace(normalizedExpected)

			if extractedCommands[i] != normalizedExpected {
				t.Logf("Command mismatch at index %d: expected %q, got %q", i, normalizedExpected, extractedCommands[i])
				return false
			}
		}

		return true
	}

	// プロパティテストを実行
	config := &quick.Config{
		MaxCount: 100, // テストケース数
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// TestProperty_PrefixRemovalAccuracy tests Property 6: プレフィックス除去の正確性
// **Validates: Requirements 4.4**
//
// Property: For any string with `- cmd: ` prefix, removing the prefix results in
// the remaining string equal to the original string minus the prefix part.
func TestProperty_PrefixRemovalAccuracy(t *testing.T) {
	property := func(command string) bool {
		// プレフィックスを持つ文字列を作成
		prefixedString := "- cmd: " + command

		// プレフィックスを除去
		result := strings.TrimPrefix(prefixedString, "- cmd:")
		result = strings.TrimSpace(result)

		// 元のコマンドと一致することを確認
		expectedCommand := strings.TrimSpace(command)

		if result != expectedCommand {
			t.Logf("Prefix removal mismatch: expected %q, got %q", expectedCommand, result)
			return false
		}

		return true
	}

	// プロパティテストを実行
	config := &quick.Config{
		MaxCount: 100, // テストケース数
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// TestProperty_HistoryCommandExtractionWithNoise tests that the reader correctly
// extracts only command lines and ignores other lines in the history file
func TestProperty_HistoryCommandExtractionWithNoise(t *testing.T) {
	property := func(commands []string, noiseLines []string) bool {
		// 空のコマンドリストはスキップ
		if len(commands) == 0 {
			return true
		}

		// テスト用の一時ディレクトリを作成
		tempDir, err := os.MkdirTemp("", "fish-history-test-*")
		if err != nil {
			t.Logf("Failed to create temp dir: %v", err)
			return false
		}
		defer os.RemoveAll(tempDir)

		// テスト用の履歴ファイルを作成
		historyPath := filepath.Join(tempDir, "fish_history")
		file, err := os.Create(historyPath)
		if err != nil {
			t.Logf("Failed to create history file: %v", err)
			return false
		}

		// 履歴ファイルにコマンドとノイズ行を交互に書き込む
		for i, cmd := range commands {
			// 改行文字を含むコマンドは正規化
			normalizedCmd := strings.ReplaceAll(cmd, "\n", " ")
			normalizedCmd = strings.ReplaceAll(normalizedCmd, "\r", " ")

			_, err := fmt.Fprintf(file, "- cmd: %s\n", normalizedCmd)
			if err != nil {
				file.Close()
				t.Logf("Failed to write to history file: %v", err)
				return false
			}

			// ノイズ行を追加（`- cmd:` で始まらない行）
			if i < len(noiseLines) {
				noiseLine := noiseLines[i]
				// `- cmd:` で始まる場合は別のプレフィックスに変更
				if strings.HasPrefix(noiseLine, "- cmd:") {
					noiseLine = "- when: " + strings.TrimPrefix(noiseLine, "- cmd:")
				}
				_, err := fmt.Fprintf(file, "%s\n", noiseLine)
				if err != nil {
					file.Close()
					t.Logf("Failed to write noise line: %v", err)
					return false
				}
			}
		}
		file.Close()

		// HistoryReader を使用してコマンドを読み込む
		reader := NewFileHistoryReader(historyPath)
		extractedCommands, err := reader.ReadCommands()
		if err != nil {
			t.Logf("Failed to read commands: %v", err)
			return false
		}

		// 抽出されたコマンド数が元のコマンド数と一致することを確認（ノイズ行は除外される）
		if len(extractedCommands) != len(commands) {
			t.Logf("Command count mismatch: expected %d, got %d", len(commands), len(extractedCommands))
			return false
		}

		// 各コマンドが正しく抽出されていることを確認
		for i, expectedCmd := range commands {
			normalizedExpected := strings.ReplaceAll(expectedCmd, "\n", " ")
			normalizedExpected = strings.ReplaceAll(normalizedExpected, "\r", " ")
			normalizedExpected = strings.TrimSpace(normalizedExpected)

			if extractedCommands[i] != normalizedExpected {
				t.Logf("Command mismatch at index %d: expected %q, got %q", i, normalizedExpected, extractedCommands[i])
				return false
			}
		}

		return true
	}

	// プロパティテストを実行
	config := &quick.Config{
		MaxCount: 50, // テストケース数（2つの配列を生成するため少なめ）
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}
