package ui

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// ConsolePrompter は標準入力からユーザー入力を受け取る Prompter の実装
type ConsolePrompter struct {
	reader *bufio.Reader
}

// NewConsolePrompter は新しい ConsolePrompter を作成する
func NewConsolePrompter(input io.Reader) *ConsolePrompter {
	return &ConsolePrompter{
		reader: bufio.NewReader(input),
	}
}

// PromptString はユーザーに文字列の入力を求める
// 空白のみの入力はエラーとして扱う
func (p *ConsolePrompter) PromptString(message string) (string, error) {
	fmt.Print(message)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("入力の読み込みに失敗しました: %w", err)
	}

	// 前後の空白を除去
	trimmed := strings.TrimSpace(input)

	// 空白のみの入力を検証
	if trimmed == "" {
		return "", fmt.Errorf("入力は空白のみで構成できません")
	}

	return trimmed, nil
}

// PromptChoice はユーザーに選択肢から1つを選ばせる
// 選択肢に含まれない入力はエラーとして扱う
func (p *ConsolePrompter) PromptChoice(message string, choices []string) (string, error) {
	if len(choices) == 0 {
		return "", fmt.Errorf("選択肢が指定されていません")
	}

	fmt.Printf("%s [%s]: ", message, strings.Join(choices, "/"))

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("入力の読み込みに失敗しました: %w", err)
	}

	// 前後の空白を除去
	trimmed := strings.TrimSpace(input)

	// 空白のみの入力を検証
	if trimmed == "" {
		return "", fmt.Errorf("入力は空白のみで構成できません")
	}

	// 選択肢に含まれるか検証
	for _, choice := range choices {
		if trimmed == choice {
			return trimmed, nil
		}
	}

	return "", fmt.Errorf("無効な選択です: %s（有効な選択肢: %s）", trimmed, strings.Join(choices, ", "))
}

// PromptConfirm はユーザーに yes/no の確認を求める
// y, yes, Y, YES は true を返し、n, no, N, NO は false を返す
func (p *ConsolePrompter) PromptConfirm(message string) (bool, error) {
	fmt.Printf("%s [y/n]: ", message)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("入力の読み込みに失敗しました: %w", err)
	}

	// 前後の空白を除去
	trimmed := strings.TrimSpace(input)

	// 空白のみの入力を検証
	if trimmed == "" {
		return false, fmt.Errorf("入力は空白のみで構成できません")
	}

	// 小文字に変換して比較
	lower := strings.ToLower(trimmed)

	switch lower {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("無効な入力です: %s（y/yes または n/no を入力してください）", trimmed)
	}
}
