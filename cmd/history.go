package cmd

import (
	"fmt"
	"io"
	"strings"

	"fish-configurator/internal/config"
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
		fmt.Fprintf(c.errOut, "Error: External Command: %v\n", err)
		return err
	}

	// History_File を読み込む（要件 4.1, 4.2, 4.12）
	commands, err := c.historyReader.ReadCommands()
	if err != nil {
		fmt.Fprintf(c.errOut, "Error: File System: 履歴ファイルの読み込みに失敗しました: %v\n", err)
		return err
	}

	// fzf でコマンドを選択（要件 4.3, 4.4, 4.14）
	selected, err := c.fzfSelector.Select(commands)
	if err != nil {
		fmt.Fprintf(c.errOut, "Error: External Command: %v\n", err)
		return err
	}

	// 選択されたコマンドから `- cmd: ` プレフィックスを除去（要件 4.4）
	// HistoryReader が既にプレフィックスを除去しているが、念のため確認
	selected = strings.TrimPrefix(selected, "- cmd: ")
	selected = strings.TrimSpace(selected)

	// alias と abbr のどちらを作成するか選択を求める（要件 4.5）
	entryType, err := c.prompter.PromptChoice("作成する種類を選択してください", []string{"alias", "abbr"})
	if err != nil {
		fmt.Fprintf(c.errOut, "Error: 選択の取得に失敗しました: %v\n", err)
		return err
	}

	// 名前の入力を求める（要件 4.6）
	name, err := c.prompter.PromptString(fmt.Sprintf("%sの名前を入力してください: ", entryType))
	if err != nil {
		fmt.Fprintf(c.errOut, "Error: Validation: 名前の入力が無効です: %v\n", err)
		return err
	}

	// 名前の空白チェック
	if strings.TrimSpace(name) == "" {
		fmt.Fprintf(c.errOut, "Error: Validation: 名前は空白のみで構成できません\n")
		return fmt.Errorf("name cannot be empty or whitespace only")
	}

	// Management_File にエントリを追加（要件 4.7, 4.8, 4.9, 4.10）
	// ConfigManager.AddEntry が内部でシンタックスチェックを処理する
	if err := c.configManager.AddEntry(entryType, name, selected); err != nil {
		fmt.Fprintf(c.errOut, "Error: %sの追加に失敗しました: %v\n", entryType, err)
		return err
	}

	// 成功メッセージを表示（要件 4.11）
	fmt.Fprintf(c.out, "%s '%s' を追加しました。\n", entryType, name)
	return nil
}
