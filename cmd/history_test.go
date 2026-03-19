package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"fish-configurator/internal/history"
)

// mockHistoryReader は history.HistoryReader のモック
type mockHistoryReader struct {
	readCommandsFunc func() ([]string, error)
}

func (m *mockHistoryReader) ReadCommands() ([]string, error) {
	if m.readCommandsFunc != nil {
		return m.readCommandsFunc()
	}
	return []string{}, nil
}

// mockFzfSelector は history.FzfSelector のモック
type mockFzfSelector struct {
	selectFunc            func(items []string) (string, error)
	checkAvailabilityFunc func() error
}

func (m *mockFzfSelector) Select(items []string) (string, error) {
	if m.selectFunc != nil {
		return m.selectFunc(items)
	}
	return "", nil
}

func (m *mockFzfSelector) CheckAvailability() error {
	if m.checkAvailabilityFunc != nil {
		return m.checkAvailabilityFunc()
	}
	return nil
}

// インターフェース準拠の確認
var _ history.HistoryReader = (*mockHistoryReader)(nil)
var _ history.FzfSelector = (*mockFzfSelector)(nil)

func TestHistoryCommand_Success(t *testing.T) {
	var addedType, addedName, addedDef string
	reader := &mockHistoryReader{
		readCommandsFunc: func() ([]string, error) {
			return []string{"git status", "ls -la", "docker ps"}, nil
		},
	}
	fzf := &mockFzfSelector{
		selectFunc: func(items []string) (string, error) {
			return "git status", nil
		},
	}
	mgr := &mockConfigManager{
		addEntryFunc: func(entryType, name, definition string) error {
			addedType = entryType
			addedName = name
			addedDef = definition
			return nil
		},
	}
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			return "alias", nil
		},
		stringFunc: func(message string) (string, error) {
			return "gs", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewHistoryCommand(reader, fzf, mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if addedType != "alias" {
		t.Errorf("expected type 'alias', got %q", addedType)
	}
	if addedName != "gs" {
		t.Errorf("expected name 'gs', got %q", addedName)
	}
	if addedDef != "git status" {
		t.Errorf("expected definition 'git status', got %q", addedDef)
	}
	if !bytes.Contains(out.Bytes(), []byte("alias 'gs' を追加しました。")) {
		t.Errorf("expected success message, got: %s", out.String())
	}
	if errOut.String() != "" {
		t.Errorf("unexpected stderr: %s", errOut.String())
	}
}

func TestHistoryCommand_AbbrSuccess(t *testing.T) {
	var addedType, addedName, addedDef string
	reader := &mockHistoryReader{
		readCommandsFunc: func() ([]string, error) {
			return []string{"docker ps", "git push"}, nil
		},
	}
	fzf := &mockFzfSelector{
		selectFunc: func(items []string) (string, error) {
			return "git push", nil
		},
	}
	mgr := &mockConfigManager{
		addEntryFunc: func(entryType, name, definition string) error {
			addedType = entryType
			addedName = name
			addedDef = definition
			return nil
		},
	}
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			return "abbr", nil
		},
		stringFunc: func(message string) (string, error) {
			return "gp", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewHistoryCommand(reader, fzf, mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if addedType != "abbr" {
		t.Errorf("expected type 'abbr', got %q", addedType)
	}
	if addedName != "gp" {
		t.Errorf("expected name 'gp', got %q", addedName)
	}
	if addedDef != "git push" {
		t.Errorf("expected definition 'git push', got %q", addedDef)
	}
	if !bytes.Contains(out.Bytes(), []byte("abbr 'gp' を追加しました。")) {
		t.Errorf("expected success message, got: %s", out.String())
	}
}

func TestHistoryCommand_FzfNotAvailable(t *testing.T) {
	reader := &mockHistoryReader{}
	fzf := &mockFzfSelector{
		checkAvailabilityFunc: func() error {
			return fmt.Errorf("fzf が見つかりません")
		},
	}
	mgr := &mockConfigManager{}
	prompter := &mockPrompter{}

	var out, errOut bytes.Buffer
	cmd := NewHistoryCommand(reader, fzf, mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("Error: External Command:")) {
		t.Errorf("expected fzf error on stderr, got: %s", errOut.String())
	}
}

func TestHistoryCommand_HistoryFileError(t *testing.T) {
	reader := &mockHistoryReader{
		readCommandsFunc: func() ([]string, error) {
			return nil, fmt.Errorf("履歴ファイルを開けません: no such file or directory")
		},
	}
	fzf := &mockFzfSelector{}
	mgr := &mockConfigManager{}
	prompter := &mockPrompter{}

	var out, errOut bytes.Buffer
	cmd := NewHistoryCommand(reader, fzf, mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("履歴ファイルの読み込みに失敗しました")) {
		t.Errorf("expected history file error, got: %s", errOut.String())
	}
}

func TestHistoryCommand_FzfSelectionCancelled(t *testing.T) {
	reader := &mockHistoryReader{
		readCommandsFunc: func() ([]string, error) {
			return []string{"git status"}, nil
		},
	}
	fzf := &mockFzfSelector{
		selectFunc: func(items []string) (string, error) {
			return "", fmt.Errorf("選択がキャンセルされました")
		},
	}
	mgr := &mockConfigManager{}
	prompter := &mockPrompter{}

	var out, errOut bytes.Buffer
	cmd := NewHistoryCommand(reader, fzf, mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("Error: External Command:")) {
		t.Errorf("expected fzf cancel error, got: %s", errOut.String())
	}
}

func TestHistoryCommand_PromptChoiceError(t *testing.T) {
	reader := &mockHistoryReader{
		readCommandsFunc: func() ([]string, error) {
			return []string{"git status"}, nil
		},
	}
	fzf := &mockFzfSelector{
		selectFunc: func(items []string) (string, error) {
			return "git status", nil
		},
	}
	mgr := &mockConfigManager{}
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			return "", fmt.Errorf("入力エラー")
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewHistoryCommand(reader, fzf, mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("選択の取得に失敗しました")) {
		t.Errorf("expected choice error, got: %s", errOut.String())
	}
}

func TestHistoryCommand_EmptyName(t *testing.T) {
	reader := &mockHistoryReader{
		readCommandsFunc: func() ([]string, error) {
			return []string{"git status"}, nil
		},
	}
	fzf := &mockFzfSelector{
		selectFunc: func(items []string) (string, error) {
			return "git status", nil
		},
	}
	mgr := &mockConfigManager{}
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			return "alias", nil
		},
		stringFunc: func(message string) (string, error) {
			return "   ", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewHistoryCommand(reader, fzf, mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("名前は空白のみで構成できません")) {
		t.Errorf("expected empty name error, got: %s", errOut.String())
	}
}

func TestHistoryCommand_AddEntryError(t *testing.T) {
	reader := &mockHistoryReader{
		readCommandsFunc: func() ([]string, error) {
			return []string{"invalid syntax '"}, nil
		},
	}
	fzf := &mockFzfSelector{
		selectFunc: func(items []string) (string, error) {
			return "invalid syntax '", nil
		},
	}
	mgr := &mockConfigManager{
		addEntryFunc: func(entryType, name, definition string) error {
			return fmt.Errorf("syntax validation failed")
		},
	}
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			return "alias", nil
		},
		stringFunc: func(message string) (string, error) {
			return "test", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewHistoryCommand(reader, fzf, mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("追加に失敗しました")) {
		t.Errorf("expected add entry error, got: %s", errOut.String())
	}
}

func TestHistoryCommand_NamePromptError(t *testing.T) {
	reader := &mockHistoryReader{
		readCommandsFunc: func() ([]string, error) {
			return []string{"git status"}, nil
		},
	}
	fzf := &mockFzfSelector{
		selectFunc: func(items []string) (string, error) {
			return "git status", nil
		},
	}
	mgr := &mockConfigManager{}
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			return "alias", nil
		},
		stringFunc: func(message string) (string, error) {
			return "", fmt.Errorf("入力は空白のみで構成できません")
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewHistoryCommand(reader, fzf, mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("名前の入力が無効です")) {
		t.Errorf("expected name prompt error, got: %s", errOut.String())
	}
}
