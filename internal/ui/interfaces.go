package ui

// Prompter はユーザーからの入力を受け取り、検証する
type Prompter interface {
	PromptString(message string) (string, error)
	PromptChoice(message string, choices []string) (string, error)
	PromptConfirm(message string) (bool, error)
}
