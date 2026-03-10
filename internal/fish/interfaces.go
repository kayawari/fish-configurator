package fish

// Executor は fish shell コマンドを実行し、結果を取得する
type Executor interface {
	ExecuteCommand(command string) (string, error)
	CheckAvailability() error
}

