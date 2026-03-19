package cmd

import (
	"fmt"
	"io"
	"strings"

	"fish-configurator/internal/config"
	apperrors "fish-configurator/internal/errors"
	"fish-configurator/internal/history"
	"fish-configurator/internal/ui"
)

// HistoryCommand は履歴からalias/abbrを作成するコマンド
type HistoryCommand struct {
	historyReader history.HistoryReader
	fzfSelector   history.FzfSelector
	configManager config.ConfigManager
	prompter      ui.Prompter
	out           io.Writer
	errOut        io.Writer
}

// NewHistoryCommand は新しい HistoryCommand を作成する
func NewHistoryCommand(
	historyReader history.HistoryReader,
	fzfSelector history.FzfSelector,
	configManager config.ConfigManager,
	prompter ui.Prompter,
	out io.Writer,
	errOut io.Writer,
) *HistoryCommand {
	return &HistoryCommand{
		historyReader: historyReader,
		fzfSelector:   fzfSelector,
		configManager: configManager,
		prompter:      prompter,
		out:           out,
		errOut:        errOut,
	}
}

// Execute は履歴選択コマンドを実行する
func (c *HistoryCommand) Execute(args []string) error {
	// fzf の利用可能性をチェック（要件 4.13）
	if err := c.fzfSelector.CheckAvailability(); err != nil {
		appErr := apperrors.NewExternalCmdError("fzfが利用できません", err)
		fmt.Fprintln(c.errOut, appErr.Error())
		return appErr
	}

	// History_File を読み込む（要件 4.1, 4.2, 4.12）
	commands, err := c.historyReader.ReadCommands()
	if err != nil {
		appErr := apperrors.NewFileSystemError("履歴ファイルの読み込みに失敗しました", err)
		fmt.Fprintln(c.errOut, appErr.Error())
		return appErr
	}

	// fzf でコマンドを選択（要件 4.3, 4.4, 4.14）
	selected, err := c.fzfSelector.Select(commands)
	if err != nil {
		appErr := apperrors.NewExternalCmdError("fzfでの選択に失敗しました", err)
		fmt.Fprintln(c.errOut, appErr.Error())
		return appErr
	}

	// 選択されたコマンドから `- cmd: ` プレフィックスを除去（要件 4.4）
	// HistoryReader が既にプレフィックスを除去しているが、念のため確認
	selected = strings.TrimPrefix(selected, "- cmd: ")
	selected = strings.TrimSpace(selected)

	// alias と abbr のどちらを作成するか選択を求める（要件 4.5）
	entryType, err := c.prompter.PromptChoice("作成する種類を選択してください", []string{"alias", "abbr"})
	if err != nil {
		appErr := apperrors.NewValidationError("選択の取得に失敗しました", err)
		fmt.Fprintln(c.errOut, appErr.Error())
		return appErr
	}

	// 名前の入力を求める（要件 4.6）
	name, err := c.prompter.PromptString(fmt.Sprintf("%sの名前を入力してください: ", entryType))
	if err != nil {
		appErr := apperrors.NewValidationError(fmt.Sprintf("名前の入力が無効です: %v", err), nil)
		fmt.Fprintln(c.errOut, appErr.Error())
		return appErr
	}

	// 名前の空白チェック
	if strings.TrimSpace(name) == "" {
		appErr := apperrors.NewValidationError("名前は空白のみで構成できません", nil)
		fmt.Fprintln(c.errOut, appErr.Error())
		return appErr
	}

	// Management_File にエントリを追加（要件 4.7, 4.8, 4.9, 4.10）
	// ConfigManager.AddEntry が内部でシンタックスチェックを処理する
	if err := c.configManager.AddEntry(entryType, name, selected); err != nil {
		appErr := apperrors.NewFileSystemError(fmt.Sprintf("%sの追加に失敗しました", entryType), err)
		fmt.Fprintln(c.errOut, appErr.Error())
		return appErr
	}

	// 成功メッセージを表示（要件 4.11）
	fmt.Fprintf(c.out, "%s '%s' を追加しました。\n", entryType, name)
	return nil
}
