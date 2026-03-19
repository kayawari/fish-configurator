package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"fish-configurator/internal/config"
)

// mockConfigManager は config.ConfigManager のモック
type mockConfigManager struct {
	loadFunc        func() (*config.Config, error)
	saveFunc        func(cfg *config.Config) error
	addEntryFunc    func(entryType, name, definition string) error
	removeEntryFunc func(entryType, name string) error
	listEntriesFunc func(entryType string) ([]config.Entry, error)
}

func (m *mockConfigManager) Load() (*config.Config, error) {
	if m.loadFunc != nil {
		return m.loadFunc()
	}
	return &config.Config{Entries: []config.Entry{}}, nil
}

func (m *mockConfigManager) Save(cfg *config.Config) error {
	if m.saveFunc != nil {
		return m.saveFunc(cfg)
	}
	return nil
}

func (m *mockConfigManager) AddEntry(entryType, name, definition string) error {
	if m.addEntryFunc != nil {
		return m.addEntryFunc(entryType, name, definition)
	}
	return nil
}

func (m *mockConfigManager) RemoveEntry(entryType, name string) error {
	if m.removeEntryFunc != nil {
		return m.removeEntryFunc(entryType, name)
	}
	return nil
}

func (m *mockConfigManager) ListEntries(entryType string) ([]config.Entry, error) {
	if m.listEntriesFunc != nil {
		return m.listEntriesFunc(entryType)
	}
	return []config.Entry{}, nil
}

// インターフェース準拠の確認
var _ config.ConfigManager = (*mockConfigManager)(nil)

func TestAddCommand_AddAlias(t *testing.T) {
	var addedType, addedName, addedDef string
	mgr := &mockConfigManager{
		addEntryFunc: func(entryType, name, definition string) error {
			addedType = entryType
			addedName = name
			addedDef = definition
			return nil
		},
	}
	prompter := &mockPrompter{
		stringFunc: func(message string) (string, error) {
			if contains(message, "名前") {
				return "ll", nil
			}
			return "ls -la", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewAddCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute([]string{"alias"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if addedType != "alias" {
		t.Errorf("expected type 'alias', got %q", addedType)
	}
	if addedName != "ll" {
		t.Errorf("expected name 'll', got %q", addedName)
	}
	if addedDef != "ls -la" {
		t.Errorf("expected definition 'ls -la', got %q", addedDef)
	}
	if !bytes.Contains(out.Bytes(), []byte("alias 'll' を追加しました。")) {
		t.Errorf("expected success message, got: %s", out.String())
	}
	if errOut.String() != "" {
		t.Errorf("unexpected stderr: %s", errOut.String())
	}
}

func TestAddCommand_AddAbbr(t *testing.T) {
	var addedType, addedName, addedDef string
	mgr := &mockConfigManager{
		addEntryFunc: func(entryType, name, definition string) error {
			addedType = entryType
			addedName = name
			addedDef = definition
			return nil
		},
	}
	prompter := &mockPrompter{
		stringFunc: func(message string) (string, error) {
			if contains(message, "名前") {
				return "gco", nil
			}
			return "git checkout", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewAddCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute([]string{"abbr"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if addedType != "abbr" {
		t.Errorf("expected type 'abbr', got %q", addedType)
	}
	if addedName != "gco" {
		t.Errorf("expected name 'gco', got %q", addedName)
	}
	if addedDef != "git checkout" {
		t.Errorf("expected definition 'git checkout', got %q", addedDef)
	}
	if !bytes.Contains(out.Bytes(), []byte("abbr 'gco' を追加しました。")) {
		t.Errorf("expected success message, got: %s", out.String())
	}
}

func TestAddCommand_NoArgs(t *testing.T) {
	mgr := &mockConfigManager{}
	prompter := &mockPrompter{}

	var out, errOut bytes.Buffer
	cmd := NewAddCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute([]string{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("Error:")) {
		t.Errorf("expected error on stderr, got: %s", errOut.String())
	}
}

func TestAddCommand_InvalidSubcommand(t *testing.T) {
	mgr := &mockConfigManager{}
	prompter := &mockPrompter{}

	var out, errOut bytes.Buffer
	cmd := NewAddCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute([]string{"invalid"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("無効なサブコマンド")) {
		t.Errorf("expected invalid subcommand error, got: %s", errOut.String())
	}
}

func TestAddCommand_EmptyNameFromPrompt(t *testing.T) {
	mgr := &mockConfigManager{}
	prompter := &mockPrompter{
		stringFunc: func(message string) (string, error) {
			if contains(message, "名前") {
				return "", fmt.Errorf("入力は空白のみで構成できません")
			}
			return "ls -la", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewAddCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute([]string{"alias"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("名前の入力が無効です")) {
		t.Errorf("expected name validation error, got: %s", errOut.String())
	}
}

func TestAddCommand_EmptyDefinitionFromPrompt(t *testing.T) {
	mgr := &mockConfigManager{}
	prompter := &mockPrompter{
		stringFunc: func(message string) (string, error) {
			if contains(message, "名前") {
				return "ll", nil
			}
			return "", fmt.Errorf("入力は空白のみで構成できません")
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewAddCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute([]string{"alias"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("定義の入力が無効です")) {
		t.Errorf("expected definition validation error, got: %s", errOut.String())
	}
}

func TestAddCommand_DuplicateOverwrite(t *testing.T) {
	removeCalledWith := ""
	var addedDef string
	mgr := &mockConfigManager{
		loadFunc: func() (*config.Config, error) {
			return &config.Config{
				Entries: []config.Entry{
					{Type: "alias", Name: "ll", Definition: "ls -l"},
				},
			}, nil
		},
		removeEntryFunc: func(entryType, name string) error {
			removeCalledWith = name
			return nil
		},
		addEntryFunc: func(entryType, name, definition string) error {
			addedDef = definition
			return nil
		},
	}
	prompter := &mockPrompter{
		stringFunc: func(message string) (string, error) {
			if contains(message, "名前") {
				return "ll", nil
			}
			return "ls -la", nil
		},
		confirmFunc: func(message string) (bool, error) {
			return true, nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewAddCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute([]string{"alias"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if removeCalledWith != "ll" {
		t.Errorf("expected RemoveEntry called with 'll', got %q", removeCalledWith)
	}
	if addedDef != "ls -la" {
		t.Errorf("expected new definition 'ls -la', got %q", addedDef)
	}
	if !bytes.Contains(out.Bytes(), []byte("alias 'll' を追加しました。")) {
		t.Errorf("expected success message, got: %s", out.String())
	}
}

func TestAddCommand_DuplicateCancel(t *testing.T) {
	addCalled := false
	mgr := &mockConfigManager{
		loadFunc: func() (*config.Config, error) {
			return &config.Config{
				Entries: []config.Entry{
					{Type: "alias", Name: "ll", Definition: "ls -l"},
				},
			}, nil
		},
		addEntryFunc: func(entryType, name, definition string) error {
			addCalled = true
			return nil
		},
	}
	prompter := &mockPrompter{
		stringFunc: func(message string) (string, error) {
			if contains(message, "名前") {
				return "ll", nil
			}
			return "ls -la", nil
		},
		confirmFunc: func(message string) (bool, error) {
			return false, nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewAddCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute([]string{"alias"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if addCalled {
		t.Error("AddEntry should not be called when user cancels overwrite")
	}
	if !bytes.Contains(out.Bytes(), []byte("追加をキャンセルしました。")) {
		t.Errorf("expected cancel message, got: %s", out.String())
	}
}

func TestAddCommand_AddEntryError(t *testing.T) {
	mgr := &mockConfigManager{
		addEntryFunc: func(entryType, name, definition string) error {
			return fmt.Errorf("syntax validation failed: invalid syntax")
		},
	}
	prompter := &mockPrompter{
		stringFunc: func(message string) (string, error) {
			if contains(message, "名前") {
				return "test", nil
			}
			return "invalid '", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewAddCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute([]string{"alias"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("追加に失敗しました")) {
		t.Errorf("expected add failure error, got: %s", errOut.String())
	}
}

func TestAddCommand_LoadError(t *testing.T) {
	mgr := &mockConfigManager{
		loadFunc: func() (*config.Config, error) {
			return nil, fmt.Errorf("permission denied")
		},
	}
	prompter := &mockPrompter{
		stringFunc: func(message string) (string, error) {
			if contains(message, "名前") {
				return "ll", nil
			}
			return "ls -la", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewAddCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute([]string{"alias"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("設定の読み込みに失敗しました")) {
		t.Errorf("expected load error, got: %s", errOut.String())
	}
}

// contains はメッセージに指定文字列が含まれるかチェックするヘルパー
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

func TestAddCommand_WhitespaceOnlyName(t *testing.T) {
	mgr := &mockConfigManager{}
	prompter := &mockPrompter{
		stringFunc: func(message string) (string, error) {
			if contains(message, "名前") {
				return "   ", nil
			}
			return "ls -la", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewAddCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute([]string{"alias"})
	if err == nil {
		t.Fatal("expected error for whitespace-only name, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("名前は空白のみで構成できません")) {
		t.Errorf("expected whitespace name error, got: %s", errOut.String())
	}
}

func TestAddCommand_WhitespaceOnlyDefinition(t *testing.T) {
	mgr := &mockConfigManager{}
	prompter := &mockPrompter{
		stringFunc: func(message string) (string, error) {
			if contains(message, "名前") {
				return "ll", nil
			}
			return "   ", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewAddCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute([]string{"alias"})
	if err == nil {
		t.Fatal("expected error for whitespace-only definition, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("定義は空白のみで構成できません")) {
		t.Errorf("expected whitespace definition error, got: %s", errOut.String())
	}
}

func TestAddCommand_DuplicateConfirmError(t *testing.T) {
	mgr := &mockConfigManager{
		loadFunc: func() (*config.Config, error) {
			return &config.Config{
				Entries: []config.Entry{
					{Type: "alias", Name: "ll", Definition: "ls -l"},
				},
			}, nil
		},
	}
	prompter := &mockPrompter{
		stringFunc: func(message string) (string, error) {
			if contains(message, "名前") {
				return "ll", nil
			}
			return "ls -la", nil
		},
		confirmFunc: func(message string) (bool, error) {
			return false, fmt.Errorf("confirm prompt failed")
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewAddCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute([]string{"alias"})
	if err == nil {
		t.Fatal("expected error when confirm fails, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("確認の取得に失敗しました")) {
		t.Errorf("expected confirm error message, got: %s", errOut.String())
	}
}

func TestAddCommand_DuplicateRemoveError(t *testing.T) {
	mgr := &mockConfigManager{
		loadFunc: func() (*config.Config, error) {
			return &config.Config{
				Entries: []config.Entry{
					{Type: "alias", Name: "ll", Definition: "ls -l"},
				},
			}, nil
		},
		removeEntryFunc: func(entryType, name string) error {
			return fmt.Errorf("remove failed: permission denied")
		},
	}
	prompter := &mockPrompter{
		stringFunc: func(message string) (string, error) {
			if contains(message, "名前") {
				return "ll", nil
			}
			return "ls -la", nil
		},
		confirmFunc: func(message string) (bool, error) {
			return true, nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewAddCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute([]string{"alias"})
	if err == nil {
		t.Fatal("expected error when remove fails, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("既存エントリの削除に失敗しました")) {
		t.Errorf("expected remove error message, got: %s", errOut.String())
	}
}
