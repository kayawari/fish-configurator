package history

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// DefaultFzfSelector は fzf プロセスを起動し、ユーザーの選択を取得する実装
type DefaultFzfSelector struct {
	fzfPath string
}

// NewDefaultFzfSelector は新しい DefaultFzfSelector を作成する
// fzfPath が空の場合、システムの PATH から fzf を検索する
func NewDefaultFzfSelector(fzfPath string) *DefaultFzfSelector {
	if fzfPath == "" {
		fzfPath = "fzf"
	}
	return &DefaultFzfSelector{
		fzfPath: fzfPath,
	}
}

// CheckAvailability は fzf が利用可能かどうかをチェックする
func (s *DefaultFzfSelector) CheckAvailability() error {
	_, err := exec.LookPath(s.fzfPath)
	if err != nil {
		return fmt.Errorf("fzf が見つかりません。fzf をインストールしてください: %w", err)
	}
	return nil
}

// Select は fzf を使用してアイテムを選択する
// items: 選択肢のリスト
// 戻り値: 選択されたアイテム、エラー
// ユーザーがキャンセルした場合（終了コード130）、エラーを返す
func (s *DefaultFzfSelector) Select(items []string) (string, error) {
	// fzf が利用可能かチェック
	if err := s.CheckAvailability(); err != nil {
		return "", err
	}

	// アイテムが空の場合はエラー
	if len(items) == 0 {
		return "", fmt.Errorf("選択肢が空です")
	}

	// fzf プロセスを起動
	cmd := exec.Command(s.fzfPath)

	// 標準入力にアイテムを渡す
	input := strings.Join(items, "\n")
	cmd.Stdin = strings.NewReader(input)

	// 標準出力をキャプチャ
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// 標準エラー出力をキャプチャ
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// コマンドを実行
	err := cmd.Run()
	if err != nil {
		// 終了コードをチェック
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()

			// 終了コード 130 はユーザーがキャンセルした場合
			if exitCode == 130 {
				return "", fmt.Errorf("選択がキャンセルされました")
			}

			// その他のエラー
			return "", fmt.Errorf("fzf の実行に失敗しました (終了コード: %d): %s", exitCode, stderr.String())
		}

		return "", fmt.Errorf("fzf の実行に失敗しました: %w", err)
	}

	// 選択結果を取得（末尾の改行を除去）
	selected := strings.TrimSpace(stdout.String())

	// 選択結果が空の場合はエラー
	if selected == "" {
		return "", fmt.Errorf("何も選択されませんでした")
	}

	return selected, nil
}
