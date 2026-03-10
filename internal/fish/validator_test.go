package fish

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidateFile_ValidAlias は有効なaliasのシンタックスチェックをテストする
func TestValidateFile_ValidAlias(t *testing.T) {
	validator := NewFishValidator()
	
	// テスト用の一時ファイルを作成
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.fish")
	
	// 有効なalias定義を書き込む
	content := "alias ll 'ls -la'\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	// シンタックスチェックを実行
	err := validator.ValidateFile(testFile)
	if err != nil {
		t.Errorf("Expected no error for valid alias, got %v", err)
	}
}

// TestValidateFile_ValidAbbr は有効なabbrのシンタックスチェックをテストする
func TestValidateFile_ValidAbbr(t *testing.T) {
	validator := NewFishValidator()
	
	// テスト用の一時ファイルを作成
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.fish")
	
	// 有効なabbr定義を書き込む
	content := "abbr -a gco 'git checkout'\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	// シンタックスチェックを実行
	err := validator.ValidateFile(testFile)
	if err != nil {
		t.Errorf("Expected no error for valid abbr, got %v", err)
	}
}

// TestValidateFile_ComplexCommand は複雑なコマンドのシンタックスチェックをテストする
func TestValidateFile_ComplexCommand(t *testing.T) {
	validator := NewFishValidator()
	
	testCases := []struct {
		name    string
		content string
	}{
		{
			name:    "pipe command",
			content: "alias count 'ls -la | wc -l'\n",
		},
		{
			name:    "command with options",
			content: "alias grep_color 'grep --color=auto'\n",
		},
		{
			name:    "command with redirection",
			content: "abbr -a save 'echo test > output.txt'\n",
		},
		{
			name:    "command with semicolon",
			content: "alias multi 'echo first; echo second'\n",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// テスト用の一時ファイルを作成
			tempDir := t.TempDir()
			testFile := filepath.Join(tempDir, "test.fish")
			
			if err := os.WriteFile(testFile, []byte(tc.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}
			
			err := validator.ValidateFile(testFile)
			if err != nil {
				t.Errorf("Expected no error for %s, got %v", tc.name, err)
			}
		})
	}
}

// TestValidateFile_InvalidSyntax は無効な構文のエラーハンドリングをテストする
func TestValidateFile_InvalidSyntax(t *testing.T) {
	validator := NewFishValidator()
	
	testCases := []struct {
		name    string
		content string
	}{
		{
			name:    "unclosed single quote",
			content: "alias bad 'echo 'unclosed'\n",
		},
		{
			name:    "unbalanced quotes",
			content: "abbr -a bad_balance 'echo 'test\" mixed'\n",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// テスト用の一時ファイルを作成
			tempDir := t.TempDir()
			testFile := filepath.Join(tempDir, "test.fish")
			
			if err := os.WriteFile(testFile, []byte(tc.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}
			
			err := validator.ValidateFile(testFile)
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

// TestValidateFile_MultipleEntries は複数のエントリを含むファイルのシンタックスチェックをテストする
func TestValidateFile_MultipleEntries(t *testing.T) {
	validator := NewFishValidator()
	
	// テスト用の一時ファイルを作成
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.fish")
	
	// 複数のエントリを含む内容
	content := `# Aliases
alias ll 'ls -la'
alias gs 'git status'

# Abbreviations
abbr -a gco 'git checkout'
abbr -a gp 'git push'
`
	
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	// シンタックスチェックを実行
	err := validator.ValidateFile(testFile)
	if err != nil {
		t.Errorf("Expected no error for multiple entries, got %v", err)
	}
}

// TestValidateFile_EmptyFile は空のファイルのシンタックスチェックをテストする
func TestValidateFile_EmptyFile(t *testing.T) {
	validator := NewFishValidator()
	
	// テスト用の一時ファイルを作成
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.fish")
	
	// 空のファイルを作成
	if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	// シンタックスチェックを実行（空のファイルは有効）
	err := validator.ValidateFile(testFile)
	if err != nil {
		t.Errorf("Expected no error for empty file, got %v", err)
	}
}

// TestValidateFile_SpecialCharacters は特殊文字を含む定義のシンタックスチェックをテストする
func TestValidateFile_SpecialCharacters(t *testing.T) {
	validator := NewFishValidator()
	
	testCases := []struct {
		name    string
		content string
	}{
		{
			name:    "dollar sign",
			content: "alias var 'echo $HOME'\n",
		},
		{
			name:    "asterisk",
			content: "alias all 'ls *.txt'\n",
		},
		{
			name:    "question mark",
			content: "alias single 'ls ?.txt'\n",
		},
		{
			name:    "brackets",
			content: "abbr -a range 'ls [a-z].txt'\n",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// テスト用の一時ファイルを作成
			tempDir := t.TempDir()
			testFile := filepath.Join(tempDir, "test.fish")
			
			if err := os.WriteFile(testFile, []byte(tc.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}
			
			err := validator.ValidateFile(testFile)
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

// TestValidateFile_NonExistentFile は存在しないファイルのエラーハンドリングをテストする
func TestValidateFile_NonExistentFile(t *testing.T) {
	validator := NewFishValidator()
	
	// 存在しないファイルパス
	nonExistentFile := "/tmp/non-existent-file-12345.fish"
	
	err := validator.ValidateFile(nonExistentFile)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}
