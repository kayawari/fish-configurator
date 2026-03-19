package cmd

import (
	"fmt"
	"io"
	"strings"

	"fish-configurator/internal/config"
	"fish-configurator/internal/ui"
)

// AddCommand は alias/abbr の追加コマンド
type AddCommand struct {
	configManager config.ConfigManager
	prompter      ui.Prompter
	out           io.Writer
	errOut        io.Writer
}

// NewAddCommand は新しい AddCommand を作成する
func NewAddCommand(configManager config.ConfigManager, prompter ui.Prompter, out io.Writer, errOut io.Writer) *AddCommand {
	return &AddCommand{
		configManager: configManager,
		prompter:      prompter,
		out:           out,
		errOut:        errOut,
	}
}

// Execute は追加コマンドを実行する
// args[0] には "alias" または "abbr" が指定される（要件 7.4, 7.5）
func (c *AddCommand) Execute(args []string) error {
	if len(args) == 0 {
		fmt.Fprintf(c.errOut, "Error: Validation: サブコマンド（alias または abbr）を指定してください\n")
		return fmt.Errorf("subcommand required: alias or abbr")
	}

	entryType := args[0]
	if entryType != "alias" && entryType != "abbr" {
		fmt.Fprintf(c.errOut, "Error: Validation: 無効なサブコマンドです: %s（alias または abbr を指定してください）\n", entryType)
		return fmt.Errorf("invalid subcommand: %s", entryType)
	}

	// 名前の入力を求める（要件 2.1, 2.2）
	name, err := c.prompter.PromptString(fmt.Sprintf("%sの名前を入力してください: ", entryType))
	if err != nil {
		fmt.Fprintf(c.errOut, "Error: Validation: 名前の入力が無効です: %v\n", err)
		return err
	}

	// 名前の空白チェック（要件 2.5）
	if strings.TrimSpace(name) == "" {
		fmt.Fprintf(c.errOut, "Error: Validation: 名前は空白のみで構成できません\n")
		return fmt.Errorf("name cannot be empty or whitespace only")
	}

	// 定義内容の入力を求める（要件 2.3）
	definition, err := c.prompter.PromptString(fmt.Sprintf("%sの定義を入力してください: ", entryType))
	if err != nil {
		fmt.Fprintf(c.errOut, "Error: Validation: 定義の入力が無効です: %v\n", err)
		return err
	}

	// 定義の空白チェック（要件 2.6）
	if strings.TrimSpace(definition) == "" {
		fmt.Fprintf(c.errOut, "Error: Validation: 定義は空白のみで構成できません\n")
		return fmt.Errorf("definition cannot be empty or whitespace only")
	}

	// 重複チェック（要件 6.4）
	existingConfig, err := c.configManager.Load()
	if err != nil {
		fmt.Fprintf(c.errOut, "Error: File System: 設定の読み込みに失敗しました: %v\n", err)
		return err
	}

	for _, entry := range existingConfig.Entries {
		if entry.Type == entryType && entry.Name == name {
			// 重複が見つかった場合、上書き確認を求める
			overwrite, err := c.prompter.PromptConfirm(
				fmt.Sprintf("警告: %s '%s' は既に存在します。上書きしますか？", entryType, name),
			)
			if err != nil {
				fmt.Fprintf(c.errOut, "Error: 確認の取得に失敗しました: %v\n", err)
				return err
			}
			if !overwrite {
				fmt.Fprintf(c.out, "追加をキャンセルしました。\n")
				return nil
			}
			// 上書きする場合は既存エントリを削除
			if err := c.configManager.RemoveEntry(entryType, name); err != nil {
				fmt.Fprintf(c.errOut, "Error: File System: 既存エントリの削除に失敗しました: %v\n", err)
				return err
			}
			break
		}
	}

	// Management_File にエントリを追加（要件 2.7, 2.8, 2.9, 2.10, 2.11）
	// ConfigManager.AddEntry が内部でシンタックスチェックとロールバックを処理する
	if err := c.configManager.AddEntry(entryType, name, definition); err != nil {
		fmt.Fprintf(c.errOut, "Error: %sの追加に失敗しました: %v\n", entryType, err)
		return err
	}

	// 成功メッセージを表示（要件 2.12）
	fmt.Fprintf(c.out, "%s '%s' を追加しました。\n", entryType, name)
	return nil
}
