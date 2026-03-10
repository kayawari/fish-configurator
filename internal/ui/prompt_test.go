package ui

import (
	"strings"
	"testing"
)

func TestPromptString_ValidInput(t *testing.T) {
	input := "test input\n"
	prompter := NewConsolePrompter(strings.NewReader(input))

	result, err := prompter.PromptString("Enter something: ")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != "test input" {
		t.Errorf("Expected 'test input', got '%s'", result)
	}
}

func TestPromptString_TrimWhitespace(t *testing.T) {
	input := "  test input  \n"
	prompter := NewConsolePrompter(strings.NewReader(input))

	result, err := prompter.PromptString("Enter something: ")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != "test input" {
		t.Errorf("Expected 'test input', got '%s'", result)
	}
}

func TestPromptString_EmptyInput(t *testing.T) {
	input := "\n"
	prompter := NewConsolePrompter(strings.NewReader(input))

	_, err := prompter.PromptString("Enter something: ")
	if err == nil {
		t.Error("Expected error for empty input, got nil")
	}
}

func TestPromptString_WhitespaceOnly(t *testing.T) {
	testCases := []string{
		" \n",
		"  \n",
		"\t\n",
		"   \t  \n",
	}

	for _, tc := range testCases {
		prompter := NewConsolePrompter(strings.NewReader(tc))
		_, err := prompter.PromptString("Enter something: ")
		if err == nil {
			t.Errorf("Expected error for whitespace-only input %q, got nil", tc)
		}
	}
}

func TestPromptChoice_ValidChoice(t *testing.T) {
	input := "alias\n"
	prompter := NewConsolePrompter(strings.NewReader(input))

	result, err := prompter.PromptChoice("Select type", []string{"alias", "abbr"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != "alias" {
		t.Errorf("Expected 'alias', got '%s'", result)
	}
}

func TestPromptChoice_InvalidChoice(t *testing.T) {
	input := "invalid\n"
	prompter := NewConsolePrompter(strings.NewReader(input))

	_, err := prompter.PromptChoice("Select type", []string{"alias", "abbr"})
	if err == nil {
		t.Error("Expected error for invalid choice, got nil")
	}
}

func TestPromptChoice_EmptyChoices(t *testing.T) {
	input := "anything\n"
	prompter := NewConsolePrompter(strings.NewReader(input))

	_, err := prompter.PromptChoice("Select type", []string{})
	if err == nil {
		t.Error("Expected error for empty choices, got nil")
	}
}

func TestPromptChoice_WhitespaceOnly(t *testing.T) {
	input := "   \n"
	prompter := NewConsolePrompter(strings.NewReader(input))

	_, err := prompter.PromptChoice("Select type", []string{"alias", "abbr"})
	if err == nil {
		t.Error("Expected error for whitespace-only input, got nil")
	}
}

func TestPromptConfirm_Yes(t *testing.T) {
	testCases := []string{"y\n", "yes\n", "Y\n", "YES\n", "Yes\n"}

	for _, tc := range testCases {
		prompter := NewConsolePrompter(strings.NewReader(tc))
		result, err := prompter.PromptConfirm("Confirm?")
		if err != nil {
			t.Errorf("Expected no error for input %q, got %v", tc, err)
		}
		if !result {
			t.Errorf("Expected true for input %q, got false", tc)
		}
	}
}

func TestPromptConfirm_No(t *testing.T) {
	testCases := []string{"n\n", "no\n", "N\n", "NO\n", "No\n"}

	for _, tc := range testCases {
		prompter := NewConsolePrompter(strings.NewReader(tc))
		result, err := prompter.PromptConfirm("Confirm?")
		if err != nil {
			t.Errorf("Expected no error for input %q, got %v", tc, err)
		}
		if result {
			t.Errorf("Expected false for input %q, got true", tc)
		}
	}
}

func TestPromptConfirm_Invalid(t *testing.T) {
	testCases := []string{"maybe\n", "x\n", "1\n"}

	for _, tc := range testCases {
		prompter := NewConsolePrompter(strings.NewReader(tc))
		_, err := prompter.PromptConfirm("Confirm?")
		if err == nil {
			t.Errorf("Expected error for invalid input %q, got nil", tc)
		}
	}
}

func TestPromptConfirm_WhitespaceOnly(t *testing.T) {
	input := "   \n"
	prompter := NewConsolePrompter(strings.NewReader(input))

	_, err := prompter.PromptConfirm("Confirm?")
	if err == nil {
		t.Error("Expected error for whitespace-only input, got nil")
	}
}
