package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultConfigManager は ConfigManager インターフェースのデフォルト実装です
type DefaultConfigManager struct {
	filePath string
	parser   Parser
}

// NewConfigManager は新しい ConfigManager インスタンスを作成します
func NewConfigManager() ConfigManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// フォールバック: 環境変数から取得
		homeDir = os.Getenv("HOME")
	}

	filePath := filepath.Join(homeDir, ".config", "fish", "conf.d", "fish-configurator.fish")

	return &DefaultConfigManager{
		filePath: filePath,
		parser:   NewParser(),
	}
}

// Load は Management_File を読み込んで Config を返します
func (m *DefaultConfigManager) Load() (*Config, error) {
	// ファイルが存在しない場合は空の Config を返す
	if _, err := os.Stat(m.filePath); os.IsNotExist(err) {
		return &Config{Entries: []Entry{}}, nil
	}

	content, err := os.ReadFile(m.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read management file: %w", err)
	}

	entries, err := m.parser.Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse management file: %w", err)
	}

	return &Config{Entries: entries}, nil
}

// Save は Config を Management_File に保存します
func (m *DefaultConfigManager) Save(config *Config) error {
	// Config_Directory が存在しない場合は作成
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// ファイル内容を生成
	var builder strings.Builder
	builder.WriteString("# このファイルは fish-configurator によって自動生成されます\n")
	builder.WriteString("# 手動で編集しないでください\n\n")

	// Aliases セクション
	hasAlias := false
	for _, entry := range config.Entries {
		if entry.Type == "alias" {
			if !hasAlias {
				builder.WriteString("# Aliases\n")
				hasAlias = true
			}
			builder.WriteString(fmt.Sprintf("alias %s '%s'\n", entry.Name, entry.Definition))
		}
	}

	if hasAlias {
		builder.WriteString("\n")
	}

	// Abbreviations セクション
	hasAbbr := false
	for _, entry := range config.Entries {
		if entry.Type == "abbr" {
			if !hasAbbr {
				builder.WriteString("# Abbreviations\n")
				hasAbbr = true
			}
			builder.WriteString(fmt.Sprintf("abbr -a %s '%s'\n", entry.Name, entry.Definition))
		}
	}

	// ファイルに書き込み
	if err := os.WriteFile(m.filePath, []byte(builder.String()), 0644); err != nil {
		return fmt.Errorf("failed to write management file: %w", err)
	}

	return nil
}

// AddEntry は新しいエントリを追加します
func (m *DefaultConfigManager) AddEntry(entryType, name, definition string) error {
	// 入力検証
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name cannot be empty or whitespace only")
	}
	if strings.TrimSpace(definition) == "" {
		return fmt.Errorf("definition cannot be empty or whitespace only")
	}

	// 現在の設定を読み込む
	config, err := m.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 新しいエントリを追加
	config.Entries = append(config.Entries, Entry{
		Type:       entryType,
		Name:       name,
		Definition: definition,
	})

	// 保存
	if err := m.Save(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// RemoveEntry は指定されたエントリを削除します
func (m *DefaultConfigManager) RemoveEntry(entryType, name string) error {
	// 現在の設定を読み込む
	config, err := m.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// エントリを削除
	newEntries := []Entry{}
	found := false
	for _, entry := range config.Entries {
		if entry.Type == entryType && entry.Name == name {
			found = true
			continue
		}
		newEntries = append(newEntries, entry)
	}

	if !found {
		return fmt.Errorf("entry not found: %s %s", entryType, name)
	}

	config.Entries = newEntries

	// 保存
	if err := m.Save(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// ListEntries は指定された種類のエントリ一覧を返します
func (m *DefaultConfigManager) ListEntries(entryType string) ([]Entry, error) {
	// 現在の設定を読み込む
	config, err := m.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 指定された種類のエントリをフィルタリング
	var entries []Entry
	for _, entry := range config.Entries {
		if entry.Type == entryType {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}
