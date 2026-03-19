package errors

import (
	"fmt"
	"testing"
)

func TestAppError_Error_WithWrappedError(t *testing.T) {
	inner := fmt.Errorf("permission denied")
	err := NewFileSystemError("設定の読み込みに失敗しました", inner)

	expected := "Error: File System: 設定の読み込みに失敗しました: permission denied"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestAppError_Error_WithoutWrappedError(t *testing.T) {
	err := NewValidationError("名前は空白のみで構成できません", nil)

	expected := "Error: Validation: 名前は空白のみで構成できません"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestAppError_Unwrap(t *testing.T) {
	inner := fmt.Errorf("original error")
	err := NewGeneralError("something failed", inner)

	if err.Unwrap() != inner {
		t.Errorf("expected unwrapped error to be %v, got %v", inner, err.Unwrap())
	}
}

func TestAppError_ExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected int
	}{
		{"validation", NewValidationError("test", nil), ExitValidationError},
		{"filesystem", NewFileSystemError("test", nil), ExitFileSystemError},
		{"fish shell", NewFishShellError("test", nil), ExitExternalCmdError},
		{"external cmd", NewExternalCmdError("test", nil), ExitExternalCmdError},
		{"general", NewGeneralError("test", nil), ExitGeneralError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.ExitCode() != tc.expected {
				t.Errorf("expected exit code %d, got %d", tc.expected, tc.err.ExitCode())
			}
		})
	}
}

func TestGetExitCode_NilError(t *testing.T) {
	if GetExitCode(nil) != ExitSuccess {
		t.Errorf("expected ExitSuccess for nil error")
	}
}

func TestGetExitCode_PlainError(t *testing.T) {
	err := fmt.Errorf("plain error")
	if GetExitCode(err) != ExitGeneralError {
		t.Errorf("expected ExitGeneralError for plain error")
	}
}

func TestGetExitCode_AppError(t *testing.T) {
	err := NewFileSystemError("test", nil)
	if GetExitCode(err) != ExitFileSystemError {
		t.Errorf("expected ExitFileSystemError for AppError")
	}
}

func TestErrorMessageFormat(t *testing.T) {
	// 要件 6.6: すべてのエラーメッセージは "Error: <type>: <description>" 形式
	tests := []struct {
		name     string
		err      *AppError
		contains string
	}{
		{"validation format", NewValidationError("invalid input", nil), "Error: Validation: invalid input"},
		{"filesystem format", NewFileSystemError("read failed", nil), "Error: File System: read failed"},
		{"fish shell format", NewFishShellError("not found", nil), "Error: Fish Shell: not found"},
		{"external cmd format", NewExternalCmdError("fzf error", nil), "Error: External Command: fzf error"},
		{"general format", NewGeneralError("unknown", nil), "Error: General: unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Error() != tc.contains {
				t.Errorf("expected %q, got %q", tc.contains, tc.err.Error())
			}
		})
	}
}
