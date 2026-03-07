package cmd

// Command は CLI コマンドを表す
type Command interface {
	Execute(args []string) error
}
