package fish

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/quick"
)

// TestProperty_SyntaxValidationAccuracy tests Property 8: シンタックス検証の正確性
// **Validates: Requirements 5.4**
//
// Property: For any entry (type, name, definition), when written to Management_File,
// the content passes fish shell's syntax check (fish -n).
func TestProperty_SyntaxValidationAccuracy(t *testing.T) {
	// fish が利用可能かチェック
	if err := exec.Command("fish", "--version").Run(); err != nil {
		t.Skip("fish shell is not available, skipping property test")
	}

	// プロパティ関数: エントリをManagement_Fileに書き込むと、fish -n でシンタックスチェックが通る
	property := func(entryType string, name string, definition string) bool {
		// 入力を正規化して有効な値にする
		entryType = normalizeEntryType(entryType)
		name = normalizeName(name)
		definition = normalizeDefinition(definition)

		// 空白のみの入力はスキップ
		if strings.TrimSpace(name) == "" || strings.TrimSpace(definition) == "" {
			return true
		}

		// テスト用の一時ディレクトリを作成
		tempDir, err := os.MkdirTemp("", "fish-validator-test-*")
		if err != nil {
			t.Logf("Failed to create temp dir: %v", err)
			return false
		}
		defer os.RemoveAll(tempDir)

		// Management_File形式でファイルを作成
		testFilePath := filepath.Join(tempDir, "test-config.fish")
		
		// エントリタイプに応じて適切な形式で書き込む
		var content string
		if entryType == "alias" {
			content = fmt.Sprintf("alias %s '%s'\n", name, definition)
		} else if entryType == "abbr" {
			content = fmt.Sprintf("abbr -a %s '%s'\n", name, definition)
		} else {
			t.Logf("Unknown entry type: %s", entryType)
			return false
		}

		// ファイルに書き込む
		err = os.WriteFile(testFilePath, []byte(content), 0644)
		if err != nil {
			t.Logf("Failed to write file: %v", err)
			return false
		}

		// fish -n <file> でシンタックスチェックを実行
		cmd := exec.Command("fish", "-n", testFilePath)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Logf("Syntax validation failed for entry (type=%s, name=%s, definition=%s): %s",
				entryType, name, definition, string(output))
			return false
		}

		return true
	}

	// プロパティテストを実行
	config := &quick.Config{
		MaxCount: 100, // 100回のランダムテストを実行
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// normalizeEntryType は entryType を "alias" または "abbr" に正規化する
func normalizeEntryType(s string) string {
	// 文字列の最初の文字を使って決定
	if len(s) == 0 || s[0]%2 == 0 {
		return "alias"
	}
	return "abbr"
}

// normalizeName は name を有効な識別子に正規化する
func normalizeName(s string) string {
	if len(s) == 0 {
		return "test"
	}
	
	// 英数字とアンダースコアのみを保持
	var builder strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			builder.WriteRune(r)
		}
	}
	
	result := builder.String()
	if result == "" {
		return "test"
	}
	
	// 最大長を制限
	if len(result) > 20 {
		result = result[:20]
	}
	
	return result
}

// normalizeDefinition は definition を有効なコマンドに正規化する
func normalizeDefinition(s string) string {
	if len(s) == 0 {
		return "echo test"
	}
	
	// シングルクォートをエスケープ（fish shellの構文に合わせる）
	s = strings.ReplaceAll(s, "'", "\\'")
	
	// 改行を削除
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	
	// 最大長を制限
	if len(s) > 50 {
		s = s[:50]
	}
	
	// 空白のみの場合はデフォルト値を返す
	if strings.TrimSpace(s) == "" {
		return "echo test"
	}
	
	return s
}
