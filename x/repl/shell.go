package repl

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lonng/nano/log"
	"gopkg.in/abiosoft/ishell.v2"
)

// Repl start a shell for user
func Repl(opts ...Option) {
	for _, option := range opts {
		option(opt)
	}
	initClientPool()
	initLocalHandler()

	shell := ishell.New()
	logger = NewCliLogger(shell)
	log.SetLogger(logger)
	configure(shell)

	log.Println("Nano REPL Client")

	registerHelp(shell)

	registerInfo(shell)
	registerConnect(shell)
	registerDisconnect(shell)
	registerRequest(shell)
	registerNotify(shell)
	registerConnectStatus(shell)
	registerHistory(shell)
	registerClearHistory(shell)
	registerSetSerializer(shell)
	registerSetCommand(shell)
	registerDelCommand(shell)
	registerUpload(shell)
	registerList(shell)
	registerDownload(shell)
	registerRemove(shell)
	registerUse(shell)
	registerAlias(shell)
	registerLoginCli(shell)
	registerSync(shell)

	err := autoLoginCli(shell)
	if err != nil {
		panic(err)
	}

	err = loadCurrentSet(shell)
	if err != nil {
		shell.Println(err, ", please run use to load command set")
	}

	shell.Run()
}

func registerConnectStatus(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "connectStatus",
		Help: "return current connect status",
		Func: func(c *ishell.Context) {
			connStat := connectStatus()
			if connStat {
				c.Printf("connected\n")
			} else {
				c.Printf("not connected\n")
			}
		},
	})
}

func registerInfo(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "info",
		Func: func(c *ishell.Context) {
			err := info()
			if err != nil {
				c.Err(err)
			}
		},
		Help: "show info of all nodes",
	})
}

// registerHelp 重新注册help命令
func registerHelp(shell *ishell.Shell) {
	shell.DeleteCmd("help")
	shell.AddCmd(&ishell.Cmd{
		Name: "help",
		Func: func(c *ishell.Context) {
			helpText := c.HelpText()
			splits := strings.Split(helpText, "\n")
			for _, split := range splits {
				if !strings.Contains(split, "    [custom]") {
					c.Println(split)
				}
			}

			c.Println("Aliases:")
			for _, cmd := range c.Cmds() {
				if len(cmd.Aliases) > 0 {
					c.Println("  ", cmd.Name, "		", cmd.Aliases)
				}
			}

			c.Println("")
			c.Println("")

			c.Println("Custom commands:")
			for _, split := range splits {
				if strings.Contains(split, "    [custom]") {
					c.Println(split)
				}
			}
		},
		Help: "display help",
	})
}

// registerConnect 注册connect命令
func registerConnect(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "connect",
		Help: "connects to server",
		Func: func(c *ishell.Context) {
			var addr string
			if len(c.Args) == 0 {
				c.Print("address: ")
				addr = c.ReadLine()
			} else {
				addr = c.Args[0]
			}

			if err := connect(addr, func(route string, data []byte) {
				p := fmt.Printf
				if _, err := p("server->%s:%s\n", route, string(data)); err != nil {
					panic(err)
				}
			}); err != nil {
				c.Err(err)
			}
		},
	})
}

// registerRequest 注册request命令
func registerRequest(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "request",
		Help: "makes a request to server",
		Func: func(c *ishell.Context) {
			err := request(c.RawArgs[1:])
			if err != nil {
				c.Err(err)
			}
		},
	})
}

// registerNotify 注册notify命令
func registerNotify(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "notify",
		Help: "makes a notify to server",
		Func: func(c *ishell.Context) {
			err := notify(c.RawArgs[1:])
			if err != nil {
				c.Err(err)
			}
		},
	})
}

// registerDisconnect 注册disconnect命令
func registerDisconnect(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "disconnect",
		Help: "disconnects from server",
		Func: func(c *ishell.Context) {
			disconnect()
			c.Println("disconnected")
		},
	})
}

// registerHistory 注册history命令 查看最近100条命令记录
func registerHistory(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "history",
		Func: func(c *ishell.Context) {
			err := history(c.RawArgs[1:])
			if err != nil {
				c.Err(err)
			}
		},
		Help: "history show latest commands executed(default 100 lines)",
	})
}

// registerClearHistory 注册清除历史命令
func registerClearHistory(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "clearHistory",
		Func: func(c *ishell.Context) {
			err := clearHistory()
			if err != nil {
				c.Err(err)
			}
		},
		Help: "clear history",
	})
}

// registerSetSerializer 注册设置编码方式命令
func registerSetSerializer(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "setSerializer",
		Func: func(c *ishell.Context) {
			err := setSerializer(c.RawArgs[1])
			if err != nil {
				c.Err(err)
			}
		},
		Help: "set serializer. [options]: JSON, Protobuf",
	})
}

// registerSetCommand 注册setCommand的命令
func registerSetCommand(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name:    "setCommand",
		Aliases: []string{"set"},
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(errors.New("setCommand requires at least 1 arguments"))
				return
			}
			cmdName := c.Args[0]
			if isBuiltinCommand(cmdName) {
				c.Err(errors.New("can not set builtin commands"))
				return
			}
			allArgs := []string{cmdName}
			for {
				c.Print("$ ")
				cmdStr := c.ReadLine()
				cmdStr = strings.TrimSpace(cmdStr)
				if len(strings.Trim(cmdStr, ";")) > 0 {
					allArgs = append(allArgs, strings.Trim(cmdStr, ";"))
				}
				if cmdStr[len(cmdStr)-1] == ';' {
					break
				}
			}

			err := setCommand(shell, allArgs)
			if err != nil {
				c.Err(err)
			}
		},
		Help: `set custom command, use enter to split commands.`,
		LongHelp: `eg:
>>> setCommand test
$ connect :20000
$ disconnect
$ ;`,
	})
}

// registerDelCommand 注册delCommand命令
func registerDelCommand(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name:    "delCommand",
		Aliases: []string{"del"},
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(errors.New("delCommand requires at least 1 arguments"))
				return
			}
			cmdName := strings.TrimSpace(c.Args[0])
			if isBuiltinCommand(cmdName) {
				c.Err(errors.New("can not delete builtin commands"))
				return
			}
			shell.DeleteCmd(cmdName)
			err := delCommand(cmdName)
			if err != nil {
				c.Err(err)
			}
		},
		Help: "delete a custom command, eg:delCommand test",
	})
}

// registerUpload 注册upload命令
func registerUpload(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name:    "upload",
		Aliases: []string{"up"},
		Func: func(c *ishell.Context) {
			if len(c.Args) < 2 {
				c.Err(errors.New("upload requires at least 2 arguments"))
				return
			}
			err := upload(c.Args[0], c.Args[1])
			if err != nil {
				c.Err(err)
			}
		},
		Help: "upload local command set to remote, usage:upload localfilename remotefilename",
	})
}

// registerList 注册list命令 包含下面的子命令
func registerList(shell *ishell.Shell) {
	listCmd := &ishell.Cmd{
		Name: "list",
		Func: func(c *ishell.Context) {
			err := listAll()
			if err != nil {
				c.Err(err)
			}
		},
		Help: "list all command sets in local/account/remote, usage:list local/account/remote",
	}
	listCmd.AddCmd(&ishell.Cmd{
		Name: "local",
		Func: func(c *ishell.Context) {
			err := listLocal()
			if err != nil {
				c.Err(err)
			}
		},
		Help: "show local command sets",
	})
	listCmd.AddCmd(&ishell.Cmd{
		Name: "remote",
		Func: func(c *ishell.Context) {
			err := listRemote()
			if err != nil {
				c.Err(err)
			}
		},
		Help: "show remote command sets",
	})
	listCmd.AddCmd(&ishell.Cmd{
		Name: "account",
		Func: func(c *ishell.Context) {
			err := listAccount()
			if err != nil {
				c.Err(err)
			}
		},
		Help: "show command sets bound to current account",
	})
	listCmd.AddCmd(&ishell.Cmd{
		Name: "all",
		Func: func(c *ishell.Context) {
			err := listAll()
			if err != nil {
				c.Err(err)
			}
		},
		Help: "show all command sets",
	})
	shell.AddCmd(listCmd)
}

// registerDownload 注册download命令
func registerDownload(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name:    "download",
		Aliases: []string{"down"},
		Func: func(c *ishell.Context) {
			if len(c.Args) < 2 {
				c.Err(errors.New("download requires at least 2 arguments"))
				return
			}
			err := download(c.Args[0], c.Args[1])
			if err != nil {
				c.Err(err)
			}
		},
		Help: "download remote command set to local, usage:download remotefilename localfilename",
	})
}

// registerRemove registers remove command
func registerRemove(shell *ishell.Shell) {
	removeCmd := &ishell.Cmd{
		Name: "remove",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(errors.New("remove requires at least 1 argument"))
				return
			}
			name := c.Args[0]
			err := removeLocal(name)
			if err != nil {
				c.Err(err)
			}
			err = clearCurrentSet(name)
			if err != nil {
				c.Err(err)
			}
			logger.Printf("successfully removed %s in local command set\n", name)
		},
		Help: "remove a command set in local, usage:remove name",
	}
	removeCmd.AddCmd(&ishell.Cmd{
		Name: "local",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(errors.New("remove local requires at least 1 argument"))
				return
			}
			name := c.Args[0]
			err := removeLocal(name)
			if err != nil {
				c.Err(err)
			}
			err = clearCurrentSet(name)
			if err != nil {
				c.Err(err)
			}
			logger.Printf("successfully removed %s in local command set\n", name)
		},
		Help: "remove a command set in local, usage:remove local name",
	})

	removeCmd.AddCmd(&ishell.Cmd{
		Name: "remote",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(errors.New("remove remote requires at least 1 argument"))
				return
			}
			name := c.Args[0]
			err := removeRemote(name)
			if err != nil {
				c.Err(err)
			}
			err = clearCurrentSet(name)
			if err != nil {
				c.Err(err)
			}
			logger.Printf("successfully removed %s in remote command set\n", name)
		},
		Help: "remove a command set in remote, usage:remove remote name",
	})

	removeCmd.AddCmd(&ishell.Cmd{
		Name: "account",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(errors.New("remove account requires at least 1 argument"))
				return
			}
			name := c.Args[0]
			err := removeAccount(name)
			if err != nil {
				c.Err(err)
			}
			err = clearCurrentSet(name)
			if err != nil {
				c.Err(err)
			}
			logger.Printf("successfully removed %s in account command set\n", name)
		},
		Help: "remove a command set in account, usage:remove account name",
	})

	shell.AddCmd(removeCmd)
}

// registerUse 注册use命令  在切换命令集的时候用到
func registerUse(shell *ishell.Shell) {
	useCmd := &ishell.Cmd{
		Name: "use",
		Func: func(c *ishell.Context) {
			// use 默认use local
			if len(c.Args) < 1 {
				c.Err(errors.New("use requires at least 1 argument"))
				return
			}
			name := c.Args[0]
			err := use(shell, name, LOCAL)
			if err != nil {
				c.Err(err)
				return
			}
			c.Printf("successfully change to command set %s\n", c.Args[0])
		},
		Help: "select a command set to load",
	}
	useNewFunc := func(c *ishell.Context) {
		if len(c.Args) < 1 {
			c.Err(errors.New("use new requires at least 1 argument"))
			return
		}
		err := useNewLocal(shell, strings.TrimSpace(c.Args[0]))
		if err != nil {
			c.Err(err)
		}
	}
	useNewCmd := &ishell.Cmd{
		Name: "new",
		Func: useNewFunc,
		Help: "create a new local command set",
	}
	useNewCmd.AddCmd(&ishell.Cmd{
		Name: "local",
		Func: useNewFunc,
		Help: "create a new local command set",
	})
	useNewCmd.AddCmd(&ishell.Cmd{
		Name: "account",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(errors.New("use new account requires at least 1 argument"))
				return
			}
			err := checkLoginCli()
			if err != nil {
				c.Err(err)
			}
			err = useNewAccount(shell, strings.TrimSpace(c.Args[0]))
			if err != nil {
				c.Err(err)
			}
		},
		Help: "create a new account command set",
	})
	useCmd.AddCmd(useNewCmd)
	useCmd.AddCmd(&ishell.Cmd{
		Name: "local",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(errors.New("use local requires at least 1 argument"))
				return
			}
			name := c.Args[0]
			err := use(shell, name, LOCAL)
			if err != nil {
				c.Err(err)
				return
			}
			c.Printf("successfully change to command set %s\n", c.Args[0])
		},
		Help: "use local command set",
	})
	useCmd.AddCmd(&ishell.Cmd{
		Name: "account",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(errors.New("use account requires at least 1 argument"))
				return
			}
			name := c.Args[0]
			err := use(shell, name, ACCOUNT)
			if err != nil {
				c.Err(err)
				return
			}
			c.Printf("successfully change to command set %s\n", c.Args[0])
		},
		Help: "use account command set",
	})
	shell.AddCmd(useCmd)
}

// registerAlias 注册alias命令  起别名时用到
func registerAlias(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "alias",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 2 {
				c.Err(errors.New("alias requires at least 2 arguments"))
				return
			}
			success, err := alias(shell, c.Args[0], c.Args[1])
			if err != nil {
				c.Err(err)
				return
			}
			if success {
				c.Printf("successfully add alias:%s for %s\n", c.Args[1], c.Args[0])
			} else {
				c.Printf("no command named %s\n", c.Args[0])
			}
		},
		Help: "set alias for command",
	})
}

// registerLoginCli 注册loginCli命令  在某台机器上第一次登陆cli时用到
func registerLoginCli(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name:    "loginCli",
		Aliases: []string{"lgc"},
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(errors.New("loginCli requires at lease 1 argument"))
				return
			}
			account := c.Args[0]
			err := loginCli(shell, account)
			if err != nil {
				c.Err(err)
			}
		},
		Help: "login cli account system",
	})
}

// registerSync 注册sync命令
func registerSync(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "sync",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 4 {
				c.Err(errors.New("sync requires at least 4 arguments"))
				return
			}
			src := c.Args[0]
			srcCmd := c.Args[1]
			dst := c.Args[2]
			dstCmd := c.Args[3]
			err := sync(shell, src, srcCmd, dst, dstCmd)
			if err != nil {
				c.Err(err)
			}
		},
		Help: "sync command from local/remote/account to local/remote/account",
	})
}
