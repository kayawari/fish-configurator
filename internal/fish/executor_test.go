package fish

import (
	"strings"
	"testing"
	"time"
)

// TestExecuteCommand_ValidCommand は有効なコマンドの実行をテストする
func TestExecuteCommand_ValidCommand(t *testing.T) {
	executor := NewFishExecutor()
	
	// 簡単なechoコマンドを実行
	output, err := executor.ExecuteCommand("echo 'hello world'")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// 出力を検証（改行を除去）
	output = strings.TrimSpace(output)
	if output != "hello world" {
		t.Errorf("Expected 'hello world', got %q", output)
	}
}

// TestExecuteCommand_MultipleCommands は複数のコマンドの実行をテストする
func TestExecuteCommand_MultipleCommands(t *testing.T) {
	executor := NewFishExecutor()
	
	// セミコロンで区切られた複数のコマンド
	output, err := executor.ExecuteCommand("echo 'first'; echo 'second'")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// 両方の出力が含まれることを確認
	if !strings.Contains(output, "first") {
		t.Errorf("Expected output to contain 'first', got %q", output)
	}
	if !strings.Contains(output, "second") {
		t.Errorf("Expected output to contain 'second', got %q", output)
	}
}

// TestExecuteCommand_InvalidCommand は無効なコマンドのエラーハンドリングをテストする
func TestExecuteCommand_InvalidCommand(t *testing.T) {
	executor := NewFishExecutor()
	
	// 存在しないコマンドを実行
	_, err := executor.ExecuteCommand("nonexistent_command_12345")
	if err == nil {
		t.Error("Expected error for invalid command, got nil")
	}
	
	// エラーメッセージに "fish command failed" が含まれることを確認
	if !strings.Contains(err.Error(), "fish command failed") {
		t.Errorf("Expected error message to contain 'fish command failed', got %q", err.Error())
	}
}

// TestExecuteCommand_SyntaxError はシンタックスエラーのエラーハンドリングをテストする
func TestExecuteCommand_SyntaxError(t *testing.T) {
	executor := NewFishExecutor()
	
	// シンタックスエラーのあるコマンド
	_, err := executor.ExecuteCommand("echo 'unclosed quote")
	if err == nil {
		t.Error("Expected error for syntax error, got nil")
	}
	
	// エラーメッセージに "fish command failed" が含まれることを確認
	if !strings.Contains(err.Error(), "fish command failed") {
		t.Errorf("Expected error message to contain 'fish command failed', got %q", err.Error())
	}
}

// TestExecuteCommand_Timeout はタイムアウト処理をテストする
func TestExecuteCommand_Timeout(t *testing.T) {
	executor := NewFishExecutor()
	
	// タイムアウトを短く設定（テスト用）
	executor.timeout = 100 * time.Millisecond
	
	// 長時間実行されるコマンド（sleepコマンド）
	_, err := executor.ExecuteCommand("sleep 10")
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
	
	// エラーメッセージに "timed out" が含まれることを確認
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected error message to contain 'timed out', got %q", err.Error())
	}
}

// TestExecuteCommand_EmptyCommand は空のコマンドの実行をテストする
func TestExecuteCommand_EmptyCommand(t *testing.T) {
	executor := NewFishExecutor()
	
	// 空のコマンドを実行
	output, err := executor.ExecuteCommand("")
	if err != nil {
		t.Fatalf("Expected no error for empty command, got %v", err)
	}
	
	// 出力が空であることを確認
	output = strings.TrimSpace(output)
	if output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}
}

// TestExecuteCommand_CapturesStderr は標準エラー出力のキャプチャをテストする
func TestExecuteCommand_CapturesStderr(t *testing.T) {
	executor := NewFishExecutor()
	
	// 標準エラー出力にメッセージを出力するコマンド
	output, err := executor.ExecuteCommand("echo 'error message' >&2")
	
	// エラーは発生しないが、出力に標準エラー出力が含まれる
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// 標準エラー出力がキャプチャされていることを確認
	if !strings.Contains(output, "error message") {
		t.Errorf("Expected output to contain 'error message', got %q", output)
	}
}

// TestCheckAvailability_FishAvailable はfishが利用可能な場合のテストする
func TestCheckAvailability_FishAvailable(t *testing.T) {
	executor := NewFishExecutor()
	
	err := executor.CheckAvailability()
	if err != nil {
		t.Skipf("fish shell is not available on this system: %v", err)
	}
}

// TestCheckAvailability_ReturnsError はfishが利用できない場合のエラーをテストする
// 注: このテストは実際にfishが利用できない環境でのみ有効
func TestCheckAvailability_ReturnsError(t *testing.T) {
	t.Skip("This test requires fish to be unavailable, skipping in normal test runs")
	
	executor := NewFishExecutor()
	
	err := executor.CheckAvailability()
	if err == nil {
		t.Error("Expected error when fish is not available, got nil")
	}
	
	// エラーメッセージに "not available" が含まれることを確認
	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("Expected error message to contain 'not available', got %q", err.Error())
	}
}

// TestNewFishExecutor はコンストラクタをテストする
func TestNewFishExecutor(t *testing.T) {
	executor := NewFishExecutor()
	
	if executor == nil {
		t.Fatal("Expected non-nil executor")
	}
	
	// デフォルトのタイムアウトが5秒であることを確認
	expectedTimeout := 5 * time.Second
	if executor.timeout != expectedTimeout {
		t.Errorf("Expected timeout %v, got %v", expectedTimeout, executor.timeout)
	}
}

// TestExecuteCommand_ExitCode は非ゼロの終了コードをテストする
func TestExecuteCommand_ExitCode(t *testing.T) {
	executor := NewFishExecutor()
	
	// 非ゼロの終了コードを返すコマンド
	_, err := executor.ExecuteCommand("false")
	if err == nil {
		t.Error("Expected error for non-zero exit code, got nil")
	}
	
	// エラーメッセージに "fish command failed" が含まれることを確認
	if !strings.Contains(err.Error(), "fish command failed") {
		t.Errorf("Expected error message to contain 'fish command failed', got %q", err.Error())
	}
}
