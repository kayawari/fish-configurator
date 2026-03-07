package fish

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// FishExecutor は fish shell コマンドを実行する実装
type FishExecutor struct {
	timeout time.Duration
}

// NewFishExecutor は新しい FishExecutor を作成する
func NewFishExecutor() *FishExecutor {
	return &FishExecutor{
		timeout: 5 * time.Second,
	}
}

// ExecuteCommand は fish -c "<command>" を実行して結果を返す
func (e *FishExecutor) ExecuteCommand(command string) (string, error) {
	// タイムアウト付きコンテキストを作成
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	// fish -c "<command>" を実行
	cmd := exec.CommandContext(ctx, "fish", "-c", command)
	
	// 標準出力と標準エラー出力をキャプチャ
	output, err := cmd.CombinedOutput()
	
	// タイムアウトエラーをチェック
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("command execution timed out after %v", e.timeout)
	}
	
	if err != nil {
		return string(output), fmt.Errorf("fish command failed: %w", err)
	}
	
	return string(output), nil
}

// CheckAvailability は fish コマンドが利用可能かチェックする
func (e *FishExecutor) CheckAvailability() error {
	_, err := exec.LookPath("fish")
	if err != nil {
		return fmt.Errorf("fish shell is not available: %w", err)
	}
	return nil
}
