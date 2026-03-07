package fish

import (
	"fmt"
	"os"
	"os/exec"
)

// FishValidator は fish shell のシンタックスチェックを実行する実装
type FishValidator struct{}

// NewFishValidator は新しい FishValidator を作成する
func NewFishValidator() *FishValidator {
	return &FishValidator{}
}

// ValidateSyntax は fish shell のシンタックスチェックを実行する
func (v *FishValidator) ValidateSyntax(entryType, name, definition string) error {
	// 一時ファイルを作成
	tmpFile, err := os.CreateTemp("", "fish-configurator-*.fish")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	
	// 一時ファイルのパスを取得
	tmpPath := tmpFile.Name()
	
	// 一時ファイルを必ず削除する
	defer os.Remove(tmpPath)
	defer tmpFile.Close()
	
	// エントリタイプに応じて適切な形式で定義を書き込む
	var content string
	switch entryType {
	case "alias":
		content = fmt.Sprintf("alias %s '%s'\n", name, definition)
	case "abbr":
		content = fmt.Sprintf("abbr -a %s '%s'\n", name, definition)
	default:
		return fmt.Errorf("unknown entry type: %s", entryType)
	}
	
	// 一時ファイルに書き込む
	if _, err := tmpFile.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}
	
	// ファイルを閉じて内容を確実にディスクに書き込む
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}
	
	// fish -n <tempfile> でシンタックスチェックを実行
	cmd := exec.Command("fish", "-n", tmpPath)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("syntax validation failed: %s", string(output))
	}
	
	return nil
}
