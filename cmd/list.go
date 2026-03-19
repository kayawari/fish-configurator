package cmd

import (
	"fmt"
	"io"
	"strings"

	apperrors "fish-configurator/internal/errors"
	"fish-configurator/internal/fish"
	"fish-configurator/internal/ui"
)

// ListCommand は alias/abbr の一覧表示コマンド
type ListCommand struct {
	executor fish.Executor
	prompter ui.Prompter
	out      io.Writer
	errOut   io.Writer
}

// NewListCommand は新しい ListCommand を作成する
func NewListCommand(executor fish.Executor, prompter ui.Prompter, out io.Writer, errOut io.Writer) *ListCommand {
	return &ListCommand{
		executor: executor,
		prompter: prompter,
		out:      out,
		errOut:   errOut,
	}
}

// Execute は一覧表示コマンドを実行する
func (c *ListCommand) Execute(args []string) error {
	// aliasとabbrのどちらを表示するか選択を求める（要件 1.1）
	choice, err := c.prompter.PromptChoice("表示する種類を選択してください", []string{"alias", "abbr"})
	if err != nil {
		appErr := apperrors.NewValidationError("選択の取得に失敗しました", err)
		fmt.Fprintln(c.errOut, appErr.Error())
		return appErr
	}

	// 選択に応じてfishコマンドを実行（要件 1.2, 1.3）
	var fishCmd string
	switch choice {
	case "alias":
		fishCmd = "alias"
	case "abbr":
		fishCmd = "abbr"
	}

	output, err := c.executor.ExecuteCommand(fishCmd)
	if err != nil {
		appErr := apperrors.NewFishShellError(fmt.Sprintf("%sの一覧取得に失敗しました", choice), err)
		fmt.Fprintln(c.errOut, appErr.Error())
		return appErr
	}

	// 出力が空の場合は情報メッセージを表示（要件 1.6）
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		fmt.Fprintf(c.out, "%sは登録されていません。\n", choice)
		return nil
	}

	// 読みやすい形式で出力（要件 1.4）
	fmt.Fprintf(c.out, "=== %s 一覧 ===\n", choice)
	lines := strings.Split(trimmed, "\n")
	for _, line := range lines {
		if line != "" {
			fmt.Fprintf(c.out, "  %s\n", line)
		}
	}

	return nil
}
