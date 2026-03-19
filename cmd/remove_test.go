package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"fish-configurator/internal/config"
)

func TestRemoveCommand_RemoveAlias(t *testing.T) {
	var removedType, removedName string
	mgr := &mockConfigManager{
		listEntriesFunc: func(entryType string) ([]config.Entry, error) {
			return []config.Entry{
				{Type: "alias", Name: "ll", Definition: "ls -la"},
				{Type: "alias", Name: "gs", Definition: "git status"},
			}, nil
		},
		removeEntryFunc: func(entryType, name string) error {
			removedType = entryType
			removedName = name
			return nil
		},
	}

	choiceCallCount := 0
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			choiceCallCount++
			if choiceCallCount == 1 {
				return "alias", nil
			}
			return "ll", nil
		},
		confirmFunc: func(message string) (bool, error) {
			return true, nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewRemoveCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if removedType != "alias" {
		t.Errorf("expected removed type 'alias', got %q", removedType)
	}
	if removedName != "ll" {
		t.Errorf("expected removed name 'll', got %q", removedName)
	}
	if !bytes.Contains(out.Bytes(), []byte("alias 'll' を削除しました。")) {
		t.Errorf("expected success message, got: %s", out.String())
	}
	if errOut.String() != "" {
		t.Errorf("unexpected stderr: %s", errOut.String())
	}
}

func TestRemoveCommand_RemoveAbbr(t *testing.T) {
	var removedType, removedName string
	mgr := &mockConfigManager{
		listEntriesFunc: func(entryType string) ([]config.Entry, error) {
			return []config.Entry{
				{Type: "abbr", Name: "gco", Definition: "git checkout"},
			}, nil
		},
		removeEntryFunc: func(entryType, name string) error {
			removedType = entryType
			removedName = name
			return nil
		},
	}

	choiceCallCount := 0
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			choiceCallCount++
			if choiceCallCount == 1 {
				return "abbr", nil
			}
			return "gco", nil
		},
		confirmFunc: func(message string) (bool, error) {
			return true, nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewRemoveCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if removedType != "abbr" {
		t.Errorf("expected removed type 'abbr', got %q", removedType)
	}
	if removedName != "gco" {
		t.Errorf("expected removed name 'gco', got %q", removedName)
	}
	if !bytes.Contains(out.Bytes(), []byte("abbr 'gco' を削除しました。")) {
		t.Errorf("expected success message, got: %s", out.String())
	}
}

func TestRemoveCommand_NoEntries(t *testing.T) {
	mgr := &mockConfigManager{
		listEntriesFunc: func(entryType string) ([]config.Entry, error) {
			return []config.Entry{}, nil
		},
	}

	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			return "alias", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewRemoveCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Contains(out.Bytes(), []byte("aliasは登録されていません。")) {
		t.Errorf("expected info message, got: %s", out.String())
	}
}

func TestRemoveCommand_CancelDeletion(t *testing.T) {
	removeCalled := false
	mgr := &mockConfigManager{
		listEntriesFunc: func(entryType string) ([]config.Entry, error) {
			return []config.Entry{
				{Type: "alias", Name: "ll", Definition: "ls -la"},
			}, nil
		},
		removeEntryFunc: func(entryType, name string) error {
			removeCalled = true
			return nil
		},
	}

	choiceCallCount := 0
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			choiceCallCount++
			if choiceCallCount == 1 {
				return "alias", nil
			}
			return "ll", nil
		},
		confirmFunc: func(message string) (bool, error) {
			return false, nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewRemoveCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if removeCalled {
		t.Error("RemoveEntry should not be called when user cancels")
	}
	if !bytes.Contains(out.Bytes(), []byte("削除をキャンセルしました。")) {
		t.Errorf("expected cancel message, got: %s", out.String())
	}
}

func TestRemoveCommand_PromptChoiceError(t *testing.T) {
	mgr := &mockConfigManager{}
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			return "", fmt.Errorf("入力エラー")
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewRemoveCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("Error:")) {
		t.Errorf("expected error on stderr, got: %s", errOut.String())
	}
}

func TestRemoveCommand_ListEntriesError(t *testing.T) {
	mgr := &mockConfigManager{
		listEntriesFunc: func(entryType string) ([]config.Entry, error) {
			return nil, fmt.Errorf("permission denied")
		},
	}

	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			return "alias", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewRemoveCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("エントリ一覧の取得に失敗しました")) {
		t.Errorf("expected list error, got: %s", errOut.String())
	}
}

func TestRemoveCommand_RemoveEntryError(t *testing.T) {
	mgr := &mockConfigManager{
		listEntriesFunc: func(entryType string) ([]config.Entry, error) {
			return []config.Entry{
				{Type: "alias", Name: "ll", Definition: "ls -la"},
			}, nil
		},
		removeEntryFunc: func(entryType, name string) error {
			return fmt.Errorf("write error")
		},
	}

	choiceCallCount := 0
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			choiceCallCount++
			if choiceCallCount == 1 {
				return "alias", nil
			}
			return "ll", nil
		},
		confirmFunc: func(message string) (bool, error) {
			return true, nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewRemoveCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("削除に失敗しました")) {
		t.Errorf("expected remove error, got: %s", errOut.String())
	}
}

func TestRemoveCommand_ConfirmError(t *testing.T) {
	mgr := &mockConfigManager{
		listEntriesFunc: func(entryType string) ([]config.Entry, error) {
			return []config.Entry{
				{Type: "alias", Name: "ll", Definition: "ls -la"},
			}, nil
		},
	}

	choiceCallCount := 0
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			choiceCallCount++
			if choiceCallCount == 1 {
				return "alias", nil
			}
			return "ll", nil
		},
		confirmFunc: func(message string) (bool, error) {
			return false, fmt.Errorf("confirm error")
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewRemoveCommand(mgr, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("確認の取得に失敗しました")) {
		t.Errorf("expected confirm error, got: %s", errOut.String())
	}
}
