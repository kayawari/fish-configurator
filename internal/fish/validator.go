package fish

import (
	"fmt"
	"os/exec"
)

// FishValidator は fish shell のシンタックスチェックを実行する実装
type FishValidator struct{}

// NewFishValidator は新しい FishValidator を作成する
func NewFishValidator() *FishValidator {
	return &FishValidator{}
}

// ValidateFile は fish shell のシンタックスチェックを実行する
// 指定されたファイルに対して fish -n を実行してシンタックスをチェックする
func (v *FishValidator) ValidateFile(filePath string) error {
	// fish -n <file> でシンタックスチェックを実行
	cmd := exec.Command("fish", "-n", filePath)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("syntax validation failed: %s", string(output))
	}
	
	return nil
}
