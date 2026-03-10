package config

import (
	"os"
	"path/filepath"
	"testing"
)

// MockValidator は Validator インターフェースのモック実装
type MockValidator struct {
	ValidateFunc func(filePath string) error
}

func (m *MockValidator) ValidateFile(filePath string) error {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(filePath)
	}
	return nil
}

// TestAddEntry_WithValidator_Success はvalidatorが成功する場合のテスト
func TestAddEntry_WithValidator_Success(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "fish-configurator.fish")

	// モックvalidatorを作成（常に成功）
	validator := &MockValidator{
		ValidateFunc: func(filePath string) error {
			return nil
		},
	}

	// ConfigManagerを作成
	manager := NewConfigManager(
		WithFilePath(testFilePath),
		WithValidator(validator),
	).(*DefaultConfigManager)

	// エントリを追加
	err := manager.AddEntry("alias", "test", "echo test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// ファイルを再読み込みして確認
	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// エントリが存在することを確認
	found := false
	for _, entry := range config.Entries {
		if entry.Type == "alias" && entry.Name == "test" && entry.Definition == "echo test" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Entry not found after successful validation")
	}
}

// TestAddEntry_WithValidator_Failure はvalidatorが失敗する場合のテスト
func TestAddEntry_WithValidator_Failure(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "fish-configurator.fish")

	// 既存のエントリを追加
	manager := NewConfigManager(
		WithFilePath(testFilePath),
	).(*DefaultConfigManager)

	err := manager.AddEntry("alias", "existing", "echo existing")
	if err != nil {
		t.Fatalf("Failed to add existing entry: %v", err)
	}

	// モックvalidatorを作成（常に失敗）
	validator := &MockValidator{
		ValidateFunc: func(filePath string) error {
			return os.ErrInvalid
		},
	}

	// validatorを設定した新しいmanagerを作成
	manager = NewConfigManager(
		WithFilePath(testFilePath),
		WithValidator(validator),
	).(*DefaultConfigManager)

	// 新しいエントリを追加（失敗するはず）
	err = manager.AddEntry("alias", "bad", "echo 'unclosed")
	if err == nil {
		t.Fatal("Expected error when validator fails, got nil")
	}

	// ファイルを再読み込み
	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 既存のエントリは保持されていることを確認
	existingFound := false
	badFound := false
	for _, entry := range config.Entries {
		if entry.Type == "alias" && entry.Name == "existing" {
			existingFound = true
		}
		if entry.Type == "alias" && entry.Name == "bad" {
			badFound = true
		}
	}

	if !existingFound {
		t.Error("Existing entry was lost after validation failure")
	}

	if badFound {
		t.Error("Bad entry was added despite validation failure")
	}
}

// TestAddEntry_WithValidator_RollbackPreservesExisting はロールバックが既存エントリを保持することを確認
func TestAddEntry_WithValidator_RollbackPreservesExisting(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "fish-configurator.fish")

	// 複数の既存エントリを追加
	manager := NewConfigManager(
		WithFilePath(testFilePath),
	).(*DefaultConfigManager)

	entries := []struct {
		entryType  string
		name       string
		definition string
	}{
		{"alias", "ll", "ls -la"},
		{"alias", "gs", "git status"},
		{"abbr", "gco", "git checkout"},
	}

	for _, e := range entries {
		err := manager.AddEntry(e.entryType, e.name, e.definition)
		if err != nil {
			t.Fatalf("Failed to add entry %s: %v", e.name, err)
		}
	}

	// モックvalidatorを作成（常に失敗）
	validator := &MockValidator{
		ValidateFunc: func(filePath string) error {
			return os.ErrInvalid
		},
	}

	// validatorを設定した新しいmanagerを作成
	manager = NewConfigManager(
		WithFilePath(testFilePath),
		WithValidator(validator),
	).(*DefaultConfigManager)

	// 新しいエントリを追加（失敗するはず）
	err := manager.AddEntry("alias", "bad", "invalid syntax")
	if err == nil {
		t.Fatal("Expected error when validator fails, got nil")
	}

	// ファイルを再読み込み
	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// すべての既存エントリが保持されていることを確認
	if len(config.Entries) != len(entries) {
		t.Errorf("Expected %d entries, got %d", len(entries), len(config.Entries))
	}

	for _, expected := range entries {
		found := false
		for _, actual := range config.Entries {
			if actual.Type == expected.entryType &&
				actual.Name == expected.name &&
				actual.Definition == expected.definition {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Entry not found after rollback: %s %s", expected.entryType, expected.name)
		}
	}
}

// TestAddEntry_WithoutValidator はvalidatorなしでも動作することを確認
func TestAddEntry_WithoutValidator(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "fish-configurator.fish")

	// validatorなしのConfigManagerを作成
	manager := NewConfigManager(
		WithFilePath(testFilePath),
	).(*DefaultConfigManager)

	// エントリを追加
	err := manager.AddEntry("alias", "test", "echo test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// ファイルを再読み込みして確認
	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// エントリが存在することを確認
	found := false
	for _, entry := range config.Entries {
		if entry.Type == "alias" && entry.Name == "test" && entry.Definition == "echo test" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Entry not found when validator is not set")
	}
}
