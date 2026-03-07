package fish

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidateSyntax_ValidAlias は有効なaliasのシンタックスチェックをテストする
func TestValidateSyntax_ValidAlias(t *testing.T) {
	validator := NewFishValidator()
	
	// 有効なalias定義
	err := validator.ValidateSyntax("alias", "ll", "ls -la")
	if err != nil {
		t.Errorf("Expected no error for valid alias, got %v", err)
	}
}

// TestValidateSyntax_ValidAbbr は有効なabbrのシンタックスチェックをテストする
func TestValidateSyntax_ValidAbbr(t *testing.T) {
	validator := NewFishValidator()
	
	// 有効なabbr定義
	err := validator.ValidateSyntax("abbr", "gco", "git checkout")
	if err != nil {
		t.Errorf("Expected no error for valid abbr, got %v", err)
	}
}

// TestValidateSyntax_ComplexCommand は複雑なコマンドのシンタックスチェックをテストする
func TestValidateSyntax_ComplexCommand(t *testing.T) {
	validator := NewFishValidator()
	
	testCases := []struct {
		name       string
		entryType  string
		entryName  string
		definition string
	}{
		{
			name:       "pipe command",
			entryType:  "alias",
			entryName:  "count",
			definition: "ls -la | wc -l",
		},
		{
			name:       "command with options",
			entryType:  "alias",
			entryName:  "grep_color",
			definition: "grep --color=auto",
		},
		{
			name:       "command with redirection",
			entryType:  "abbr",
			entryName:  "save",
			definition: "echo test > output.txt",
		},
		{
			name:       "command with semicolon",
			entryType:  "alias",
			entryName:  "multi",
			definition: "echo first; echo second",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateSyntax(tc.entryType, tc.entryName, tc.definition)
			if err != nil {
				t.Errorf("Expected no error for %s, got %v", tc.name, err)
			}
		})
	}
}

// TestValidateSyntax_InvalidSyntax は無効な構文のエラーハンドリングをテストする
func TestValidateSyntax_InvalidSyntax(t *testing.T) {
	validator := NewFishValidator()
	
	testCases := []struct {
		name       string
		entryType  string
		entryName  string
		definition string
	}{
		{
			name:       "unclosed single quote",
			entryType:  "alias",
			entryName:  "bad",
			definition: "echo 'unclosed",
		},
		{
			name:       "unbalanced quotes",
			entryType:  "abbr",
			entryName:  "bad_balance",
			definition: "echo 'test\" mixed",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateSyntax(tc.entryType, tc.entryName, tc.definition)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tc.name)
				return
			}
			
			// エラーメッセージに "syntax validation failed" が含まれることを確認
			if !strings.Contains(err.Error(), "syntax validation failed") {
				t.Errorf("Expected error message to contain 'syntax validation failed', got %q", err.Error())
			}
		})
	}
}

// TestValidateSyntax_UnknownEntryType は未知のエントリタイプのエラーハンドリングをテストする
func TestValidateSyntax_UnknownEntryType(t *testing.T) {
	validator := NewFishValidator()
	
	// 未知のエントリタイプ
	err := validator.ValidateSyntax("unknown", "test", "echo test")
	if err == nil {
		t.Error("Expected error for unknown entry type, got nil")
	}
	
	// エラーメッセージに "unknown entry type" が含まれることを確認
	if !strings.Contains(err.Error(), "unknown entry type") {
		t.Errorf("Expected error message to contain 'unknown entry type', got %q", err.Error())
	}
}

// TestValidateSyntax_TempFileCleanup は一時ファイルの削除を確認するテストする
func TestValidateSyntax_TempFileCleanup(t *testing.T) {
	validator := NewFishValidator()
	
	// 一時ディレクトリのパスを取得
	tempDir := os.TempDir()
	
	// 実行前の一時ファイル数を取得
	beforeFiles, err := filepath.Glob(filepath.Join(tempDir, "fish-configurator-*.fish"))
	if err != nil {
		t.Fatalf("Failed to list temp files: %v", err)
	}
	beforeCount := len(beforeFiles)
	
	// ValidateSyntax を実行（成功ケース）
	err = validator.ValidateSyntax("alias", "test", "echo test")
	if err != nil {
		t.Fatalf("ValidateSyntax failed: %v", err)
	}
	
	// 実行後の一時ファイル数を取得
	afterFiles, err := filepath.Glob(filepath.Join(tempDir, "fish-configurator-*.fish"))
	if err != nil {
		t.Fatalf("Failed to list temp files: %v", err)
	}
	afterCount := len(afterFiles)
	
	// 一時ファイルが削除されていることを確認
	if afterCount != beforeCount {
		t.Errorf("Expected temp file to be cleaned up. Before: %d, After: %d", beforeCount, afterCount)
		t.Logf("Remaining files: %v", afterFiles)
	}
}

// TestValidateSyntax_TempFileCleanupOnError はエラー時の一時ファイル削除を確認するテストする
func TestValidateSyntax_TempFileCleanupOnError(t *testing.T) {
	validator := NewFishValidator()
	
	// 一時ディレクトリのパスを取得
	tempDir := os.TempDir()
	
	// 実行前の一時ファイル数を取得
	beforeFiles, err := filepath.Glob(filepath.Join(tempDir, "fish-configurator-*.fish"))
	if err != nil {
		t.Fatalf("Failed to list temp files: %v", err)
	}
	beforeCount := len(beforeFiles)
	
	// ValidateSyntax を実行（エラーケース）
	err = validator.ValidateSyntax("alias", "bad", "echo 'unclosed")
	if err == nil {
		t.Fatal("Expected error for invalid syntax, got nil")
	}
	
	// 実行後の一時ファイル数を取得
	afterFiles, err := filepath.Glob(filepath.Join(tempDir, "fish-configurator-*.fish"))
	if err != nil {
		t.Fatalf("Failed to list temp files: %v", err)
	}
	afterCount := len(afterFiles)
	
	// エラー時でも一時ファイルが削除されていることを確認
	if afterCount != beforeCount {
		t.Errorf("Expected temp file to be cleaned up even on error. Before: %d, After: %d", beforeCount, afterCount)
		t.Logf("Remaining files: %v", afterFiles)
	}
}

// TestValidateSyntax_EmptyDefinition は空の定義のシンタックスチェックをテストする
func TestValidateSyntax_EmptyDefinition(t *testing.T) {
	validator := NewFishValidator()
	
	// 空の定義（fishのシンタックスチェックは通るはず）
	err := validator.ValidateSyntax("alias", "empty", "")
	if err != nil {
		t.Errorf("Expected no error for empty definition, got %v", err)
	}
}

// TestValidateSyntax_SpecialCharacters は特殊文字を含む定義のシンタックスチェックをテストする
func TestValidateSyntax_SpecialCharacters(t *testing.T) {
	validator := NewFishValidator()
	
	testCases := []struct {
		name       string
		entryType  string
		entryName  string
		definition string
	}{
		{
			name:       "dollar sign",
			entryType:  "alias",
			entryName:  "var",
			definition: "echo $HOME",
		},
		{
			name:       "asterisk",
			entryType:  "alias",
			entryName:  "all",
			definition: "ls *.txt",
		},
		{
			name:       "question mark",
			entryType:  "alias",
			entryName:  "single",
			definition: "ls ?.txt",
		},
		{
			name:       "brackets",
			entryType:  "abbr",
			entryName:  "range",
			definition: "ls [a-z].txt",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateSyntax(tc.entryType, tc.entryName, tc.definition)
			if err != nil {
				t.Errorf("Expected no error for %s, got %v", tc.name, err)
			}
		})
	}
}

// TestNewFishValidator はコンストラクタをテストする
func TestNewFishValidator(t *testing.T) {
	validator := NewFishValidator()
	
	if validator == nil {
		t.Fatal("Expected non-nil validator")
	}
}

// TestValidateSyntax_MultipleValidations は複数回の検証をテストする
func TestValidateSyntax_MultipleValidations(t *testing.T) {
	validator := NewFishValidator()
	
	// 複数回の検証を実行
	for i := 0; i < 5; i++ {
		err := validator.ValidateSyntax("alias", "test", "echo test")
		if err != nil {
			t.Errorf("Validation %d failed: %v", i+1, err)
		}
	}
}
