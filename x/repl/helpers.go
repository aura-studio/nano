package repl

import (
	"bufio"
	"bytes"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"fmt"

	"github.com/lonng/nano/log"
	"gopkg.in/abiosoft/ishell.v2"
)

const (
	// JSONArray means array struct [] in json
	JSONArray = iota
	// JSONMap means map struct {} in json
	JSONMap
	// JSONString means string "" in json
	JSONString
	// JSONInt means int in json
	JSONInt
	// JSONEnd means nothing valuable found
	JSONEnd
)

func client() error {
	if opt.IsWebSocket {
		log.Println("using websocket client")

		pClient = NewWebsocket(opt.WSPath)
	} else {
		log.Println(fmt.Sprintf("Using %s client", opt.SerializerName))
		client := NewClient()
		pClient = client
	}

	return nil
}

func tryConnect(addr string) error {
	if opt.IsWebSocket {
		// TODO: compatible with all http
		// if err := pClient.ConnectWS(addr); err != nil {
		// 	return err
		// }
	} else {
		if err := pClient.Connect(addr); err != nil {
			return err
		}
	}
	return nil
}

func readServerMessages(callback func(route string, data []byte)) {
	channel := pClient.MsgChannel()
	for {
		select {
		case <-disconnectedCh:
			close(disconnectedCh)
			return
		case m := <-channel:
			callback(m.Route, parseData(m.Data))
		}
	}
}

//setCurrentCommandSet sets current command set to name when use another command set
func setCurrentCommandSet(setName string) error {
	if err := removeFile(currentPath); err != nil {
		return err
	}
	f, err := newFile(currentPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(setName); err != nil {
		return err
	}

	return f.Sync()
}

// getCurrentCommandSet returns current command set name in use
func getCurrentCommandSet() (string, error) {
	if !exists(currentPath) {
		f, err := newFile(currentPath)
		if err != nil {
			return "", err
		}
		defer f.Close()
		if _, err := f.WriteString("default"); err != nil {
			panic(err)
		}
		if err := f.Sync(); err != nil {
			panic(err)
		}
		return "default", nil
	}
	f, err := readFile(currentPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	tmpBytes := make([]byte, 128)
	length, err := f.Read(tmpBytes)
	if err != nil {
		return "", err
	}
	current := string(tmpBytes[:length])
	return current, nil
}

// save custom command to local file
func saveCmdsToFile(cmdName string, cmds string) error {
	currentPath := getCurrentCmdFile()
	f, err := appendFile(currentPath)
	if err != nil {
		log.Println("open custom command file err:%v", err)
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("%s@%s\n", cmdName, cmds)); err != nil {
		panic(err)
	}
	return f.Sync()
}

// loadAlias loads alias
func loadAlias(shell *ishell.Shell, aliases map[string][]string) {
	allCmds := shell.Cmds()
	for cmdName, aliases := range aliases {
		for _, cmd := range allCmds {
			if cmd.Name == cmdName {
				cmd.Aliases = aliases
				break
			}
		}
	}
}

func configure(c *ishell.Shell) {
	c.SetHistoryPath(historyPath)
}

func nextChar(index int, data []byte) int {
	for _, b := range data[index+1:] {
		switch b {
		case '{':
			return JSONMap
		case '[':
			return JSONArray
		case '"':
			return JSONString
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			return JSONInt
		}
	}
	return JSONEnd
}

func mergeLineInMap(index int, data []byte) bool {
	bracketsDepth := 0
	elementCount := 0
	for _, b := range data[index+1:] {
		switch b {
		case '{':
			return false
		case '}':
			return elementCount <= 8
		case '[':
			bracketsDepth++
		case ']':
			bracketsDepth--
		case ':':
			if bracketsDepth == 0 {
				elementCount++
			}
		}
	}
	return false
}

func deleteNewLine(data []byte) string {
	intArrayFlag := false
	lineMapFlag := false
	var result bytes.Buffer
	for index, b := range data {
		switch b {
		case '[':
			if nextChar(index, data) == JSONInt {
				intArrayFlag = true
			}
			result.WriteByte(b)
		case ']':
			if intArrayFlag {
				intArrayFlag = false
			}
			result.WriteByte(b)
		case '{':
			if mergeLineInMap(index, data) {
				lineMapFlag = true
			}
			result.WriteByte(b)
		case '}':
			lineMapFlag = false
			result.WriteByte(b)
		case '\n', ' ':
			if !intArrayFlag && !lineMapFlag {
				result.WriteByte(b)
			}
		default:
			result.WriteByte(b)
		}
	}

	return result.String()
}

func parseData(data []byte) []byte {
	if opt.PrettyJSON {
		var m interface{}
		_ = json.Unmarshal(data, &m)
		data, _ = json.MarshalIndent(m, "", "    ")
	}

	// replace Code: xx to Code: xx//text for code
	//dataStr := string(data)
	dataStr := deleteNewLine(data)
	lines := strings.Split(dataStr, "\n")
	codePrefix := "    \"Code\":"
	reasonPrefix := "\"Reason\""
	for ix, line := range lines {
		if strings.HasPrefix(line, codePrefix) || strings.Contains(line, reasonPrefix) {
			colonIndex := strings.Index(line, ":")
			codeStr := line[colonIndex+1:]
			codeStr = strings.ReplaceAll(codeStr, " ", "")
			codeStr = strings.Trim(codeStr, ",")
			code, err := strconv.ParseInt(codeStr, 10, 64)
			if err != nil {
				continue
			}
			if strings.HasPrefix(line, codePrefix) && opt.ErrorReader != nil {
				message := opt.ErrorReader(code)
				if message != "" {
					lines[ix] = line + " // " + message
				}
			}
			if strings.Contains(line, reasonPrefix) && opt.ReasonReader != nil {
				message := opt.ReasonReader(code)
				if message != "" {
					lines[ix] = line + " // " + message
				}
			}
		}
	}
	dataStr = strings.Join(lines, "\n")
	data = []byte(dataStr)

	return data
}

func runCommandSequence(c *ishell.Context, cmds []string, args []string, repeat int64) error {
	argsIndex := 0
	for _, cmd := range cmds {
		placeHolderCount := strings.Count(cmd, fmtPlaceHolder)
		tmp := make([]interface{}, 0)
		for i:=argsIndex;i<argsIndex+placeHolderCount;i++ {
			if len(args) <= i {
				return fmt.Errorf("arguments not enough, please check command:%s arguments count", cmd)
			}
			tmp = append(tmp, args[i])
		}
		argsIndex += placeHolderCount
		realCmd := fmt.Sprintf(cmd, tmp...)

		err := executeCommand(realCmd, repeat)
		if err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func loadCurrentSet(shell *ishell.Shell) error {
	if currentAccount == nil {
		return nil
	}
	err := loadCustomCommands(shell, currentAccount.CurrentSet, currentAccount.CurrentSetType)
	if err != nil {
		return err
	}
	return nil
}

// loadCustomCommands loads a local command set
func loadCustomCommands(shell *ishell.Shell, cmdSetName string, typ int) error {
	if currentAccount == nil {
		return fmt.Errorf("current not login cli, please run loginCli first")
	}
	switch typ {
	case LOCAL:
		cmdPath := getCmdFile(cmdSetName)
		if !exists(cmdPath) {
			return fmt.Errorf("command set:%s does not exist in local", cmdSetName)
		}
		f, err := readFile(cmdPath)
		if err != nil {
			return err
		}
		defer f.Close()
		buf := bufio.NewReader(f)
		for {
			line, err := buf.ReadString('\n')
			line = strings.TrimSpace(line)
			lineSplits := strings.Split(line, "@")
			if len(lineSplits) == 2 {
				cmdName := strings.TrimSpace(lineSplits[0])
				cmdStr := strings.TrimSpace(lineSplits[1])
				cmds := strings.Split(cmdStr, ";")
				shell.DeleteCmd(cmdName)
				addCustomCommand(shell, cmds, cmdName, cmdStr)
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}
	case ACCOUNT:
		cmdSet, ok := currentAccount.CmdSets[cmdSetName]
		if ok {
			for cmdName, cmdStr := range cmdSet {
				shell.DeleteCmd(cmdName)
				cmds := strings.Split(cmdStr, ";")
				addCustomCommand(shell, cmds, cmdName, cmdStr)
			}
		} else {
			return fmt.Errorf("command set:%s does not exist in account", cmdSetName)
		}
	}

	return nil
}

// 卸载掉当前加载的命令集里的命令, 在切换命令集的时候用到
func unloadCurrentCommands(shell *ishell.Shell) error {
	currentSet := currentAccount.CurrentSet
	switch currentAccount.CurrentSetType {
	case LOCAL:
		cmdPath := filepath.Join(cmdDir, currentSet)
		if !exists(cmdPath) {
			logger.Printf("current command set:%s does not exist in local\n", currentSet)
			return nil
		}
		f, err := readFile(cmdPath)
		if err != nil {
			return err
		}
		defer f.Close()
		buf := bufio.NewReader(f)
		for {
			line, err := buf.ReadString('\n')
			line = strings.TrimSpace(line)
			lineSplits := strings.Split(line, "@")
			if len(lineSplits) == 2 {
				cmdName := strings.TrimSpace(lineSplits[0])
				shell.DeleteCmd(cmdName)
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}
	case ACCOUNT:
		cmdSets := currentAccount.CmdSets
		cmdSet, ok := cmdSets[currentSet]
		if ok {
			for cmdName := range cmdSet {
				shell.DeleteCmd(cmdName)
			}
		} else {
			logger.Printf("current command set:%s does not exist in account\n", currentSet)
			return nil
		}
	}

	return nil
}

func saveUsername(username string) error {
	f, err := writeFile(usernamePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(username)
	if err != nil {
		return err
	}
	return f.Sync()
}

func autoLoginCli(shell *ishell.Shell) error {
	if !exists(usernamePath) {
		shell.Println("current not login cli, please run loginCli first")
		return nil
	}
	f, err := readFile(usernamePath)
	if err != nil {
		return err
	}
	defer f.Close()
	tmpBuf := make([]byte, 128)
	length, err := f.Read(tmpBuf)
	if err != nil {
		return err
	}
	username := string(tmpBuf[:length])
	return loginCli(shell, username)
}

func checkLoginCli() error {
	if currentAccount == nil {
		return fmt.Errorf("current not login cli, please run loginCli first")
	}
	return nil
}

func clearCurrentSet(removeSet string) error {
	err := checkLoginCli()
	if err != nil {
		return err
	}
	if currentAccount.CurrentSet == removeSet {
		currentAccount.CurrentSet = ""
		err := currentAccount.Save()
		if err != nil {
			return err
		}
	}
	return nil
}

func isBuiltinCommand(cmd string) bool {
	for _, builtIn := range builtinCommands {
		if builtIn == cmd {
			return true
		}
	}
	return false
}

func joinShotHelpText(cmdStr string) string {
	helpStr := "[custom]: "
	if len(cmdStr) > 100 {
		helpStr += cmdStr[:100] + "..."
	} else {
		helpStr += cmdStr
	}
	return helpStr
}

func calCmdSetLength(cmd map[string]string) int64 {
	length := 0
	for k,v := range cmd {
		length += len(k) + len(v) + 2
	}
	return int64(length)
}

func addCustomCommand(shell *ishell.Shell, cmds []string, cmdName, cmdStr string) {
	newCmd := &ishell.Cmd{
		Name: cmdName,
		Func: func(c *ishell.Context) {
			err := runCommandSequence(c, cmds, c.Args, 1)
			if err != nil {
				c.Err(err)
			}
		},
		Help:     joinShotHelpText(cmdStr),
		LongHelp: strings.Join(cmds, "\n"),
	}

	newCmd.AddCmd(&ishell.Cmd{
		Name:      "repeat",
		Func: func(c *ishell.Context) {
			var repeat int64 = 1
			if len(c.Args) > 0 {
				var err error
				repeat, err = strconv.ParseInt(c.Args[len(c.Args) - 1], 10, 64)
				if err != nil {
					repeat = 1
				}
			}

			err := runCommandSequence(c, cmds, c.Args[:len(c.Args)-1], repeat)
			if err != nil {
				c.Err(err)
			}
		},
		Help:      joinShotHelpText(cmdStr) + "with repeat",
	})

	shell.AddCmd(newCmd)
}
