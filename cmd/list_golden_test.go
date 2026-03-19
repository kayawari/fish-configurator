package cmd

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

// loadGolden はゴールデンファイルを読み込む。-update フラグが指定されている場合はファイルを更新する。
func loadGolden(t *testing.T, name string, actual string) string {
	t.Helper()
	golden := filepath.Join("testdata", name+".golden")

	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("failed to create testdata dir: %v", err)
		}
		if err := os.WriteFile(golden, []byte(actual), 0o644); err != nil {
			t.Fatalf("failed to update golden file %s: %v", golden, err)
		}
	}

	expected, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", golden, err)
	}
	return string(expected)
}

// Validates: Requirements 1.4
func TestListCommand_AliasOutput_Golden(t *testing.T) {
	executor := &mockExecutor{
		executeFunc: func(command string) (string, error) {
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

	if err := cmd.Execute(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	actual := out.String()
	expected := loadGolden(t, "list_alias", actual)
	if actual != expected {
		t.Errorf("output mismatch.\nGot:\n%s\nWant:\n%s", actual, expected)
	}
}

// Validates: Requirements 1.4
func TestListCommand_AbbrOutput_Golden(t *testing.T) {
	executor := &mockExecutor{
		executeFunc: func(command string) (string, error) {
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

	if err := cmd.Execute(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	actual := out.String()
	expected := loadGolden(t, "list_abbr", actual)
	if actual != expected {
		t.Errorf("output mismatch.\nGot:\n%s\nWant:\n%s", actual, expected)
	}
}

// Validates: Requirements 1.4
func TestListCommand_AliasEmpty_Golden(t *testing.T) {
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

	if err := cmd.Execute(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	actual := out.String()
	expected := loadGolden(t, "list_alias_empty", actual)
	if actual != expected {
		t.Errorf("output mismatch.\nGot:\n%s\nWant:\n%s", actual, expected)
	}
}

// Validates: Requirements 1.4
func TestListCommand_AbbrEmpty_Golden(t *testing.T) {
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

	if err := cmd.Execute(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	actual := out.String()
	expected := loadGolden(t, "list_abbr_empty", actual)
	if actual != expected {
		t.Errorf("output mismatch.\nGot:\n%s\nWant:\n%s", actual, expected)
	}
}

// Validates: Requirements 1.4
func TestListCommand_PromptError_Golden(t *testing.T) {
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

	actual := errOut.String()
	expected := loadGolden(t, "list_error_prompt", actual)
	if actual != expected {
		t.Errorf("stderr mismatch.\nGot:\n%s\nWant:\n%s", actual, expected)
	}
}

// Validates: Requirements 1.4
func TestListCommand_ExecutorError_Golden(t *testing.T) {
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

	actual := errOut.String()
	expected := loadGolden(t, "list_error_executor", actual)
	if actual != expected {
		t.Errorf("stderr mismatch.\nGot:\n%s\nWant:\n%s", actual, expected)
	}
}
