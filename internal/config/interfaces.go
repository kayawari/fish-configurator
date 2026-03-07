package config

// Entry は alias または abbr のエントリを表す
type Entry struct {
	Type       string // "alias" or "abbr"
	Name       string
	Definition string
}

// Config は設定ファイルの内容を表す
type Config struct {
	Entries []Entry
}

// ConfigManager は Management_File の読み書き、エントリの追加・削除を管理する
type ConfigManager interface {
	Load() (*Config, error)
	Save(config *Config) error
	AddEntry(entryType, name, definition string) error
	RemoveEntry(entryType, name string) error
	ListEntries(entryType string) ([]Entry, error)
}

// Parser は Management_File を解析してエントリを抽出する
type Parser interface {
	Parse(content string) ([]Entry, error)
}
