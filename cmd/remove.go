package cmd

import (
	"fmt"
	"io"

	"fish-configurator/internal/config"
	"fish-configurator/internal/ui"
)

// RemoveCommand は alias/abbr の削除コマンド
type RemoveCommand struct {
	configManager config.ConfigManager
	prompter      ui.Prompter
	out           io.Writer
	errOut        io.Writer
}

// NewRemoveCommand は新しい RemoveCommand を作成する
func NewRemoveCommand(configManager config.ConfigManager, prompter ui.Prompter, out io.Writer, errOut io.Writer) *RemoveCommand {
	return &RemoveCommand{
		configManager: configManager,
		prompter:      prompter,
		out:           out,
		errOut:        errOut,
	}
}

// Execute は削除コマンドを実行する
func (c *RemoveCommand) Execute(args []string) error {
	// aliasとabbrのどちらを削除するか選択を求める（要件 3.1）
	entryType, err := c.prompter.PromptChoice("削除する種類を選択してください", []string{"alias", "abbr"})
	if err != nil {
		fmt.Fprintf(c.errOut, "Error: 選択の取得に失敗しました: %v\n", err)
		return err
	}

	// Management_File から該当する種類のエントリを取得（要件 3.2）
	entries, err := c.configManager.ListEntries(entryType)
	if err != nil {
		fmt.Fprintf(c.errOut, "Error: File System: エントリ一覧の取得に失敗しました: %v\n", err)
		return err
	}

	// エントリが存在しない場合は情報メッセージを表示して終了（要件 3.9）
	if len(entries) == 0 {
		fmt.Fprintf(c.out, "%sは登録されていません。\n", entryType)
		return nil
	}

	// エントリ名のリストを作成してユーザーに選択させる（要件 3.3）
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name
	}

	selectedName, err := c.prompter.PromptChoice("削除する項目を選択してください", names)
	if err != nil {
		fmt.Fprintf(c.errOut, "Error: 選択の取得に失敗しました: %v\n", err)
		return err
	}

	// 選択されたエントリが存在するか確認（要件 3.8）
	found := false
	for _, entry := range entries {
		if entry.Name == selectedName {
			found = true
			break
		}
	}
	if !found {
		fmt.Fprintf(c.errOut, "Warning: %s '%s' はManagement_Fileに存在しません。\n", entryType, selectedName)
		return fmt.Errorf("%s '%s' not found", entryType, selectedName)
	}

	// 確認プロンプトを表示（要件 3.4）
	confirmed, err := c.prompter.PromptConfirm(fmt.Sprintf("%s '%s' を削除しますか？", entryType, selectedName))
	if err != nil {
		fmt.Fprintf(c.errOut, "Error: 確認の取得に失敗しました: %v\n", err)
		return err
	}

	if !confirmed {
		fmt.Fprintf(c.out, "削除をキャンセルしました。\n")
		return nil
	}

	// Management_File から該当エントリを削除（要件 3.5, 3.6）
	if err := c.configManager.RemoveEntry(entryType, selectedName); err != nil {
		fmt.Fprintf(c.errOut, "Error: File System: %s '%s' の削除に失敗しました: %v\n", entryType, selectedName, err)
		return err
	}

	// 成功メッセージを表示（要件 3.7）
	fmt.Fprintf(c.out, "%s '%s' を削除しました。\n", entryType, selectedName)
	return nil
}
