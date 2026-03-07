package config

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// DefaultParser は Parser インターフェースのデフォルト実装です
type DefaultParser struct {
	aliasRegex *regexp.Regexp
	abbrRegex  *regexp.Regexp
}

// NewParser は新しい Parser インスタンスを作成します
func NewParser() Parser {
	return &DefaultParser{
		aliasRegex: regexp.MustCompile(`^alias\s+(\S+)\s+'([^']+)'`),
		abbrRegex:  regexp.MustCompile(`^abbr\s+-a\s+(\S+)\s+'([^']+)'`),
	}
}

// Parse は設定ファイルの内容を解析してエントリのスライスを返します
func (p *DefaultParser) Parse(content string) ([]Entry, error) {
	var entries []Entry
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// コメント行と空行をスキップ
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// alias行の解析
		if matches := p.aliasRegex.FindStringSubmatch(line); matches != nil {
			entries = append(entries, Entry{
				Type:       "alias",
				Name:       matches[1],
				Definition: matches[2],
			})
			continue
		}

		// abbr行の解析
		if matches := p.abbrRegex.FindStringSubmatch(line); matches != nil {
			entries = append(entries, Entry{
				Type:       "abbr",
				Name:       matches[1],
				Definition: matches[2],
			})
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan content: %w", err)
	}

	return entries, nil
}
