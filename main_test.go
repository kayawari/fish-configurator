package main

import (
	"fmt"
	"testing"

	apperrors "fish-configurator/internal/errors"
)

// --- classifyError のテスト ---

// TestClassifyError_NilError は nil エラーの場合に ExitSuccess を返すことをテストする
func TestClassifyError_NilError(t *testing.T) {
	code := classifyError(nil)
	if code != apperrors.ExitSuccess {
		t.Errorf("Expected ExitSuccess (%d), got %d", apperrors.ExitSuccess, code)
	}
}

// TestClassifyError_ValidationErrors はバリデーションエラーの分類をテストする
func TestClassifyError_ValidationErrors(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{"validation error", apperrors.NewValidationError("invalid input", nil)},
		{"empty name", apperrors.NewValidationError("名前は空白のみで構成できません", nil)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := classifyError(tc.err)
			if code != apperrors.ExitValidationError {
				t.Errorf("Expected ExitValidationError (%d), got %d", apperrors.ExitValidationError, code)
			}
		})
	}
}

// TestClassifyError_FileSystemErrors はファイルシステムエラーの分類をテストする
func TestClassifyError_FileSystemErrors(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{"file read error", apperrors.NewFileSystemError("設定の読み込みに失敗しました", nil)},
		{"file write error", apperrors.NewFileSystemError("ファイルの書き込みに失敗しました", nil)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := classifyError(tc.err)
			if code != apperrors.ExitFileSystemError {
				t.Errorf("Expected ExitFileSystemError (%d), got %d", apperrors.ExitFileSystemError, code)
			}
		})
	}
}

// TestClassifyError_ExternalCmdErrors は外部コマンドエラーの分類をテストする
func TestClassifyError_ExternalCmdErrors(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{"fish shell error", apperrors.NewFishShellError("fish コマンドが見つかりません", nil)},
		{"external cmd error", apperrors.NewExternalCmdError("fzfが利用できません", nil)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := classifyError(tc.err)
			if code != apperrors.ExitExternalCmdError {
				t.Errorf("Expected ExitExternalCmdError (%d), got %d", apperrors.ExitExternalCmdError, code)
			}
		})
	}
}

// TestClassifyError_GeneralError は分類されないエラーが ExitGeneralError になることをテストする
func TestClassifyError_GeneralError(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{"general error", apperrors.NewGeneralError("something went wrong", nil)},
		{"plain error", fmt.Errorf("unexpected error occurred")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := classifyError(tc.err)
			if code != apperrors.ExitGeneralError {
				t.Errorf("Expected ExitGeneralError (%d), got %d", apperrors.ExitGeneralError, code)
			}
		})
	}
}

// --- run 関数のテスト ---

// TestRun_NoArgs はサブコマンドなしで ExitValidationError を返すことをテストする（要件 7.7）
func TestRun_NoArgs(t *testing.T) {
	code := run([]string{})
	if code != apperrors.ExitValidationError {
		t.Errorf("Expected ExitValidationError (%d) for no args, got %d", apperrors.ExitValidationError, code)
	}
}

// TestRun_HelpSubcommand はヘルプサブコマンドが ExitSuccess を返すことをテストする（要件 7.6）
func TestRun_HelpSubcommand(t *testing.T) {
	testCases := []struct {
		name string
		arg  string
	}{
		{"help", "help"},
		{"-h flag", "-h"},
		{"--help flag", "--help"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := run([]string{tc.arg})
			if code != apperrors.ExitSuccess {
				t.Errorf("Expected ExitSuccess (%d) for %q, got %d", apperrors.ExitSuccess, tc.arg, code)
			}
		})
	}
}

// TestRun_VersionSubcommand はバージョンサブコマンドが ExitSuccess を返すことをテストする
func TestRun_VersionSubcommand(t *testing.T) {
	testCases := []struct {
		name string
		arg  string
	}{
		{"version", "version"},
		{"-v flag", "-v"},
		{"--version flag", "--version"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := run([]string{tc.arg})
			if code != apperrors.ExitSuccess {
				t.Errorf("Expected ExitSuccess (%d) for %q, got %d", apperrors.ExitSuccess, tc.arg, code)
			}
		})
	}
}

// TestRun_InvalidSubcommand は無効なサブコマンドが ExitValidationError を返すことをテストする（要件 7.7）
func TestRun_InvalidSubcommand(t *testing.T) {
	testCases := []struct {
		name string
		arg  string
	}{
		{"unknown command", "invalid"},
		{"typo", "lst"},
		{"empty-like", "---"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := run([]string{tc.arg})
			if code != apperrors.ExitValidationError {
				t.Errorf("Expected ExitValidationError (%d) for %q, got %d", apperrors.ExitValidationError, tc.arg, code)
			}
		})
	}
}

// TestExitCodes_Constants は終了コード定数の値が正しいことをテストする
func TestExitCodes_Constants(t *testing.T) {
	if apperrors.ExitSuccess != 0 {
		t.Errorf("Expected ExitSuccess = 0, got %d", apperrors.ExitSuccess)
	}
	if apperrors.ExitGeneralError != 1 {
		t.Errorf("Expected ExitGeneralError = 1, got %d", apperrors.ExitGeneralError)
	}
	if apperrors.ExitValidationError != 2 {
		t.Errorf("Expected ExitValidationError = 2, got %d", apperrors.ExitValidationError)
	}
	if apperrors.ExitFileSystemError != 3 {
		t.Errorf("Expected ExitFileSystemError = 3, got %d", apperrors.ExitFileSystemError)
	}
	if apperrors.ExitExternalCmdError != 4 {
		t.Errorf("Expected ExitExternalCmdError = 4, got %d", apperrors.ExitExternalCmdError)
	}
}
