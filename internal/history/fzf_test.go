package history

import (
	"os"
	"os/exec"
	"testing"
)

// TestCheckAvailability_FzfInstalled tests when fzf is available
func TestCheckAvailability_FzfInstalled(t *testing.T) {
	// fzf が実際にインストールされているかチェック
	_, err := exec.LookPath("fzf")
	if err != nil {
		t.Skip("fzf is not installed, skipping test")
	}

	selector := NewDefaultFzfSelector("")
	err = selector.CheckAvailability()
	if err != nil {
		t.Errorf("Expected no error when fzf is installed, got: %v", err)
	}
}

// TestCheckAvailability_FzfNotInstalled tests when fzf is not available
func TestCheckAvailability_FzfNotInstalled(t *testing.T) {
	// 存在しないコマンドを指定
	selector := NewDefaultFzfSelector("/nonexistent/fzf")
	err := selector.CheckAvailability()
	if err == nil {
		t.Error("Expected error when fzf is not available, got nil")
	}
}

// TestSelect_EmptyItems tests error handling for empty items list
func TestSelect_EmptyItems(t *testing.T) {
	selector := NewDefaultFzfSelector("")

	_, err := selector.Select([]string{})
	if err == nil {
		t.Error("Expected error for empty items list, got nil")
	}
}

// TestNewDefaultFzfSelector_DefaultPath tests default fzf path
func TestNewDefaultFzfSelector_DefaultPath(t *testing.T) {
	selector := NewDefaultFzfSelector("")

	if selector.fzfPath != "fzf" {
		t.Errorf("Expected default fzfPath to be 'fzf', got %q", selector.fzfPath)
	}
}

// TestNewDefaultFzfSelector_CustomPath tests custom fzf path
func TestNewDefaultFzfSelector_CustomPath(t *testing.T) {
	customPath := "/usr/local/bin/fzf"
	selector := NewDefaultFzfSelector(customPath)

	if selector.fzfPath != customPath {
		t.Errorf("Expected fzfPath to be %q, got %q", customPath, selector.fzfPath)
	}
}

// MockFzfSelector is a mock implementation of FzfSelector for testing
type MockFzfSelector struct {
	SelectFunc            func(items []string) (string, error)
	CheckAvailabilityFunc func() error
}

func (m *MockFzfSelector) Select(items []string) (string, error) {
	if m.SelectFunc != nil {
		return m.SelectFunc(items)
	}
	return "", nil
}

func (m *MockFzfSelector) CheckAvailability() error {
	if m.CheckAvailabilityFunc != nil {
		return m.CheckAvailabilityFunc()
	}
	return nil
}

// TestMockFzfSelector_Select tests the mock selector
func TestMockFzfSelector_Select(t *testing.T) {
	expectedSelection := "test command"
	mock := &MockFzfSelector{
		SelectFunc: func(items []string) (string, error) {
			if len(items) == 0 {
				t.Error("Expected non-empty items")
			}
			return expectedSelection, nil
		},
	}

	items := []string{"test command", "another command"}
	result, err := mock.Select(items)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result != expectedSelection {
		t.Errorf("Expected %q, got %q", expectedSelection, result)
	}
}

// TestMockFzfSelector_CheckAvailability tests the mock availability check
func TestMockFzfSelector_CheckAvailability(t *testing.T) {
	mock := &MockFzfSelector{
		CheckAvailabilityFunc: func() error {
			return nil
		},
	}

	err := mock.CheckAvailability()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestFzfIntegration_RealFzf tests integration with real fzf if available
// This test is skipped if fzf is not installed
func TestFzfIntegration_RealFzf(t *testing.T) {
	// fzf が実際にインストールされているかチェック
	_, err := exec.LookPath("fzf")
	if err != nil {
		t.Skip("fzf is not installed, skipping integration test")
	}

	// この統合テストは手動で実行する必要があるため、環境変数でスキップ可能にする
	if os.Getenv("RUN_FZF_INTEGRATION_TEST") != "1" {
		t.Skip("Skipping fzf integration test (set RUN_FZF_INTEGRATION_TEST=1 to run)")
	}

	selector := NewDefaultFzfSelector("")
	items := []string{
		"ls -la",
		"git status",
		"echo hello",
	}

	// 注意: この部分は実際にfzfを起動するため、手動でテストする必要がある
	// 自動テストでは実行されない
	_, err = selector.Select(items)

	// fzf がキャンセルされた場合はエラーが返される
	// これは正常な動作なので、エラーメッセージをチェックする
	if err != nil {
		t.Logf("fzf returned error (expected if cancelled): %v", err)
	}
}

// TestFzfSelector_ErrorHandling tests various error conditions
func TestFzfSelector_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		fzfPath     string
		items       []string
		expectError bool
	}{
		{
			name:        "Empty items",
			fzfPath:     "fzf",
			items:       []string{},
			expectError: true,
		},
		{
			name:        "Invalid fzf path",
			fzfPath:     "/nonexistent/fzf",
			items:       []string{"item1", "item2"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewDefaultFzfSelector(tt.fzfPath)
			_, err := selector.Select(tt.items)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}
