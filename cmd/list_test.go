package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"fish-configurator/internal/fish"
	"fish-configurator/internal/ui"
)

// mockExecutor は fish.Executor のモック
type mockExecutor struct {
	executeFunc func(command string) (string, error)
	checkFunc   func() error
}

func (m *mockExecutor) ExecuteCommand(command string) (string, error) {
	if m.executeFunc != nil {
		return m.executeFunc(command)
	}
	return "", nil
}

func (m *mockExecutor) CheckAvailability() error {
	if m.checkFunc != nil {
		return m.checkFunc()
	}
	return nil
}

// mockPrompter は ui.Prompter のモック
type mockPrompter struct {
	choiceFunc  func(message string, choices []string) (string, error)
	stringFunc  func(message string) (string, error)
	confirmFunc func(message string) (bool, error)
}

func (m *mockPrompter) PromptChoice(message string, choices []string) (string, error) {
	if m.choiceFunc != nil {
		return m.choiceFunc(message, choices)
	}
	return "", nil
}

func (m *mockPrompter) PromptString(message string) (string, error) {
	if m.stringFunc != nil {
		return m.stringFunc(message)
	}
	return "", nil
}

func (m *mockPrompter) PromptConfirm(message string) (bool, error) {
	if m.confirmFunc != nil {
		return m.confirmFunc(message)
	}
	return false, nil
}

// インターフェース準拠の確認
var _ fish.Executor = (*mockExecutor)(nil)
var _ ui.Prompter = (*mockPrompter)(nil)

func TestListCommand_Alias(t *testing.T) {
	executor := &mockExecutor{
		executeFunc: func(command string) (string, error) {
			if command != "alias" {
				t.Errorf("expected command 'alias', got %q", command)
			}
			return "alias ll 'ls -la'\nalias gs 'git status'\n", nil
		},
	}
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			return "alias", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewListCommand(executor, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if errOut.String() != "" {
		t.Errorf("unexpected stderr output: %s", errOut.String())
	}

	// ヘッダーが含まれることを確認
	if !bytes.Contains([]byte(output), []byte("=== alias 一覧 ===")) {
		t.Errorf("expected header in output, got: %s", output)
	}
	// エントリが含まれることを確認
	if !bytes.Contains([]byte(output), []byte("alias ll 'ls -la'")) {
		t.Errorf("expected alias entry in output, got: %s", output)
	}
	if !bytes.Contains([]byte(output), []byte("alias gs 'git status'")) {
		t.Errorf("expected alias entry in output, got: %s", output)
	}
}

func TestListCommand_Abbr(t *testing.T) {
	executor := &mockExecutor{
		executeFunc: func(command string) (string, error) {
			if command != "abbr" {
				t.Errorf("expected command 'abbr', got %q", command)
			}
			return "abbr -a gco 'git checkout'\nabbr -a gp 'git push'\n", nil
		},
	}
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			return "abbr", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewListCommand(executor, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !bytes.Contains([]byte(output), []byte("=== abbr 一覧 ===")) {
		t.Errorf("expected header in output, got: %s", output)
	}
	if !bytes.Contains([]byte(output), []byte("abbr -a gco 'git checkout'")) {
		t.Errorf("expected abbr entry in output, got: %s", output)
	}
}

func TestListCommand_EmptyAlias(t *testing.T) {
	executor := &mockExecutor{
		executeFunc: func(command string) (string, error) {
			return "", nil
		},
	}
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			return "alias", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewListCommand(executor, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !bytes.Contains([]byte(output), []byte("aliasは登録されていません。")) {
		t.Errorf("expected empty message, got: %s", output)
	}
}

func TestListCommand_EmptyAbbr(t *testing.T) {
	executor := &mockExecutor{
		executeFunc: func(command string) (string, error) {
			return "\n", nil
		},
	}
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			return "abbr", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewListCommand(executor, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !bytes.Contains([]byte(output), []byte("abbrは登録されていません。")) {
		t.Errorf("expected empty message, got: %s", output)
	}
}

func TestListCommand_PromptError(t *testing.T) {
	executor := &mockExecutor{}
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			return "", fmt.Errorf("入力エラー")
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewListCommand(executor, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains(errOut.Bytes(), []byte("Error:")) {
		t.Errorf("expected error message on stderr, got: %s", errOut.String())
	}
}

func TestListCommand_ExecutorError(t *testing.T) {
	executor := &mockExecutor{
		executeFunc: func(command string) (string, error) {
			return "", fmt.Errorf("fish command failed")
		},
	}
	prompter := &mockPrompter{
		choiceFunc: func(message string, choices []string) (string, error) {
			return "alias", nil
		},
	}

	var out, errOut bytes.Buffer
	cmd := NewListCommand(executor, prompter, &out, &errOut)

	err := cmd.Execute(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errOutput := errOut.String()
	if !bytes.Contains([]byte(errOutput), []byte("Error: Fish Shell:")) {
		t.Errorf("expected fish error message on stderr, got: %s", errOutput)
	}
}
