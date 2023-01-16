package repl

const (
	cliHashKey    string = "cli_commands"
	cliAccountKey string = "cli_accounts"
)

const fmtPlaceHolder = "%s"

const (
	LOCAL = iota
	ACCOUNT
)

var (
	pClient        *Client
	disconnectedCh chan bool
	logger         *CliLogger
	currentAccount *Account
)

var builtinCommands = [...]string{
	"help",
	"connect",
	"disconnect",
	"request",
	"notify",
	"connectStatus",
	"history",
	"clearHistory",
	"setSerializer",
	"setCommand",
	"delCommand",
	"upload",
	"list",
	"download",
	"remove",
	"use",
	"alias",
	"loginCli",
	"sync",
}
