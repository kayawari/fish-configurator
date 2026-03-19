package history

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileHistoryReader は fish_history ファイルを読み込み、コマンド行を抽出する実装
type FileHistoryReader struct {
	historyPath string
}

// NewFileHistoryReader は新しい FileHistoryReader を作成する
// historyPath が空の場合、デフォルトのパス (~/.local/share/fish/fish_history) を使用する
func NewFileHistoryReader(historyPath string) *FileHistoryReader {
	if historyPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// エラーの場合はデフォルトパスを使用
			historyPath = "~/.local/share/fish/fish_history"
		} else {
			historyPath = filepath.Join(homeDir, ".local", "share", "fish", "fish_history")
		}
	}
	return &FileHistoryReader{
		historyPath: historyPath,
	}
}

// ReadCommands は fish_history ファイルからコマンドを読み込む
// `- cmd:` で始まる行を抽出し、プレフィックスを除去してコマンド文字列を返す
func (r *FileHistoryReader) ReadCommands() ([]string, error) {
	// ファイルを開く
	file, err := os.Open(r.historyPath)
	if err != nil {
		return nil, fmt.Errorf("履歴ファイルを開けません: %w", err)
	}
	defer file.Close()

	var commands []string
	scanner := bufio.NewScanner(file)

	// 行単位で読み込み
	for scanner.Scan() {
		line := scanner.Text()

		// `- cmd:` で始まる行を抽出
		if strings.HasPrefix(line, "- cmd:") {
			// プレフィックス `- cmd: ` を除去してコマンド文字列を取得
			command := strings.TrimPrefix(line, "- cmd:")
			// 先頭の空白を除去
			command = strings.TrimSpace(command)
			commands = append(commands, command)
		}
	}

	// スキャンエラーをチェック
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("履歴ファイルの読み込み中にエラーが発生しました: %w", err)
	}

	return commands, nil
}
