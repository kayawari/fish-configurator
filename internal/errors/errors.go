package errors

import "fmt"

// ExitCode はプロセスの終了コードを表す
const (
	ExitSuccess          = 0
	ExitGeneralError     = 1
	ExitValidationError  = 2
	ExitFileSystemError  = 3
	ExitExternalCmdError = 4
)

// ErrorCategory はエラーの種類を表す
type ErrorCategory string

const (
	CategoryGeneral     ErrorCategory = "General"
	CategoryValidation  ErrorCategory = "Validation"
	CategoryFileSystem  ErrorCategory = "File System"
	CategoryFishShell   ErrorCategory = "Fish Shell"
	CategoryExternalCmd ErrorCategory = "External Command"
)

// AppError はアプリケーション固有のエラーを表す
type AppError struct {
	Category ErrorCategory
	Message  string
	Err      error
}

// Error は error インターフェースを実装する
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("Error: %s: %s: %v", e.Category, e.Message, e.Err)
	}
	return fmt.Sprintf("Error: %s: %s", e.Category, e.Message)
}

// Unwrap は内部エラーを返す
func (e *AppError) Unwrap() error {
	return e.Err
}

// ExitCode はエラーカテゴリに基づいた終了コードを返す
func (e *AppError) ExitCode() int {
	switch e.Category {
	case CategoryValidation:
		return ExitValidationError
	case CategoryFileSystem:
		return ExitFileSystemError
	case CategoryFishShell, CategoryExternalCmd:
		return ExitExternalCmdError
	default:
		return ExitGeneralError
	}
}

// NewValidationError は入力検証エラーを作成する
func NewValidationError(message string, err error) *AppError {
	return &AppError{Category: CategoryValidation, Message: message, Err: err}
}

// NewFileSystemError はファイルシステムエラーを作成する
func NewFileSystemError(message string, err error) *AppError {
	return &AppError{Category: CategoryFileSystem, Message: message, Err: err}
}

// NewFishShellError は Fish Shell エラーを作成する
func NewFishShellError(message string, err error) *AppError {
	return &AppError{Category: CategoryFishShell, Message: message, Err: err}
}

// NewExternalCmdError は外部コマンドエラーを作成する
func NewExternalCmdError(message string, err error) *AppError {
	return &AppError{Category: CategoryExternalCmd, Message: message, Err: err}
}

// NewGeneralError は一般的なエラーを作成する
func NewGeneralError(message string, err error) *AppError {
	return &AppError{Category: CategoryGeneral, Message: message, Err: err}
}

// GetExitCode はエラーから終了コードを取得する
// AppError の場合はカテゴリに基づいた終了コードを返す
// それ以外のエラーの場合は ExitGeneralError を返す
func GetExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr.ExitCode()
	}
	return ExitGeneralError
}
