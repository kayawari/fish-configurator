package main

import (
	"fmt"
	"os"

	"fish-configurator/cmd"
	"fish-configurator/internal/config"
	apperrors "fish-configurator/internal/errors"
	"fish-configurator/internal/fish"
	"fish-configurator/internal/history"
	"fish-configurator/internal/ui"
)

// バージョン情報
const version = "0.1.0"

// printUsage はヘルプメッセージを表示する（要件 7.6）
func printUsage(w *os.File) {
	fmt.Fprintf(w, `fish-configurator - fish shell の alias/abbr 管理ツール

使い方:
  fish-configurator <サブコマンド> [オプション]

サブコマンド:
  list        alias/abbr の一覧を表示する
  add <type>  alias または abbr を追加する (type: alias, abbr)
  remove      alias/abbr を削除する
  history     コマンド履歴から alias/abbr を作成する
  help        このヘルプメッセージを表示する
  version     バージョン情報を表示する

例:
  fish-configurator list
  fish-configurator add alias
  fish-configurator add abbr
  fish-configurator remove
  fish-configurator history
`)
}

// classifyError はエラーから終了コードを判定する
func classifyError(err error) int {
	return apperrors.GetExitCode(err)
}

func run(args []string) int {
	// サブコマンドが指定されていない場合（要件 7.7）
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: Validation: サブコマンドが指定されていません\n\n")
		printUsage(os.Stderr)
		return apperrors.ExitValidationError
	}

	subcommand := args[0]
	subArgs := args[1:]

	// help と version は依存関係不要
	switch subcommand {
	case "help", "-h", "--help":
		printUsage(os.Stdout)
		return apperrors.ExitSuccess
	case "version", "-v", "--version":
		fmt.Fprintf(os.Stdout, "fish-configurator version %s\n", version)
		return apperrors.ExitSuccess
	}

	// 依存関係の構築
	executor := fish.NewFishExecutor()
	validator := fish.NewFishValidator()
	prompter := ui.NewConsolePrompter(os.Stdin)
	configManager := config.NewConfigManager(
		config.WithValidator(validator),
	)
	historyReader := history.NewFileHistoryReader("")
	fzfSelector := history.NewDefaultFzfSelector("")

	// コマンドの作成とルーティング（要件 7.1, 7.2, 7.3）
	var command cmd.Command
	switch subcommand {
	case "list":
		command = cmd.NewListCommand(executor, prompter, os.Stdout, os.Stderr)
	case "add":
		command = cmd.NewAddCommand(configManager, prompter, os.Stdout, os.Stderr)
	case "remove":
		command = cmd.NewRemoveCommand(configManager, prompter, os.Stdout, os.Stderr)
	case "history":
		command = cmd.NewHistoryCommand(historyReader, fzfSelector, configManager, prompter, os.Stdout, os.Stderr)
	default:
		// 無効なサブコマンド（要件 7.7）
		fmt.Fprintf(os.Stderr, "Error: Validation: 無効なサブコマンドです: %s\n\n", subcommand)
		printUsage(os.Stderr)
		return apperrors.ExitValidationError
	}

	// コマンドを実行（要件 6.5）
	if err := command.Execute(subArgs); err != nil {
		return classifyError(err)
	}

	return apperrors.ExitSuccess
}

func main() {
	os.Exit(run(os.Args[1:]))
}
