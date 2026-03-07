package history

// HistoryReader は fish_history ファイルを読み込み、コマンド行を抽出する
type HistoryReader interface {
	ReadCommands() ([]string, error)
}

// FzfSelector は fzf プロセスを起動し、ユーザーの選択を取得する
type FzfSelector interface {
	Select(items []string) (string, error)
	CheckAvailability() error
}
