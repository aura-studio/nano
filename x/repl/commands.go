// Copyright (c) TFG Co. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package repl

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"

	"github.com/lonng/nano/serialize/json"
	"github.com/lonng/nano/serialize/protobuf"
	"gopkg.in/abiosoft/ishell.v2"
)

func connectStatus() (connected bool) {
	if pClient != nil && pClient.ConnectedStatus() {
		return true
	}
	return false
}

// info 查看当前集群内各个节点运行状态
func info() error {
	if pClient == nil {
		return errors.New("not connected")
	}

	if !pClient.ConnectedStatus() {
		return errors.New("not connected")
	}
	route := "Backend.QueryRuntimeStats"
	_, err := pClient.SendRequest(route, []byte("{}"))
	if err != nil {
		return err
	}
	return nil
}

// connect 连接server
func connect(addr string, onMessageCallback func(route string, data []byte)) (err error) {
	if pClient != nil && pClient.ConnectedStatus() {
		return errors.New("already connected")
	}

	err = client()

	if err != nil {
		return err
	}

	pClient.OnConnected(func(data interface{}) {
		logger.Println("Successfully connected to ", addr)
	})

	if err = tryConnect(addr); err != nil {
		logger.Println("Failed to connect!")
		return err
	}

	disconnectedCh = make(chan bool, 1)
	go readServerMessages(onMessageCallback)

	return nil
}

// request 发请求
func request(args []string) error {
	if pClient == nil {
		return errors.New("not connected")
	}

	if !pClient.ConnectedStatus() {
		return errors.New("not connected")
	}

	if len(args) < 1 {
		return errors.New(`request should be in the format: request {route} [data]`)
	}

	args0 := args[0]
	times, err := strconv.ParseInt(args0, 10, 64)
	// 如果第一个参数 是数字 这个参数就用来当做 请求的次数
	// 如果不是数字 那就是默认的1次 第一个参数就当做route  次数这个参数不能放在尾部，是因为data这个参数里空格很多，无法区分
	if err != nil {
		route := args[0]
		var data []byte
		if len(args) > 1 {
			data = []byte(strings.Join(args[1:], ""))
		}

		_, err := pClient.SendRequest(route, data)
		if err != nil {
			return err
		}
	} else {
		route := args[1]
		var data []byte
		if len(args) > 2 {
			data = []byte(strings.Join(args[2:], ""))
		}
		for i := int64(0); i < times; i++ {
			_, err := pClient.SendRequest(route, data)
			time.Sleep(200 * time.Millisecond)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// notify 发notice
func notify(args []string) error {
	if pClient == nil {
		return errors.New("not connected")
	}

	if !pClient.ConnectedStatus() {
		return errors.New("not connected")
	}

	if len(args) < 1 {
		return errors.New(`notify should be in the format: notify {route} [data]`)
	}

	route := args[0]
	var data []byte
	if len(args) > 1 {
		data = []byte(strings.Join(args[1:], ""))
	}

	return pClient.SendNotify(route, data)
}

// disconnect 断开连接
func disconnect() {
	if pClient == nil {
		logger.Println("already disconnected")
		return
	}
	if pClient.ConnectedStatus() {
		disconnectedCh <- true
		pClient.Disconnect()
	}
}

// history 查询命令历史(默认100条，带参数n的话，就显示n条)
func history(args []string) error {
	lines := 100
	if len(args) > 0 {
		newlines, err := strconv.ParseInt(args[0], 10, 0)
		if err == nil {
			lines = int(newlines)
		}
	}
	f, err := readFile(historyPath)
	if err != nil {
		logger.Println(err)
		return err
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	allLines := make([]string, 0)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		allLines = append(allLines, line)
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Println(err)
			return err
		}
	}
	lineLen := len(allLines)
	start := lineLen - lines
	if start < 0 {
		start = 0
	}
	latestLines := allLines[start:]
	for _, hisCmd := range latestLines {
		logger.Println(hisCmd)
	}
	return nil
}

// clearHistory 清除历史记录
func clearHistory() error {
	f, err := newFile(historyPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.WriteString(f, "")
	if err != nil {
		return err
	}
	return nil
}

// setSerializer 设置序列化方法 JSON或是Protobuf
func setSerializer(serializer string) error {
	if strings.ToUpper(serializer) == "JSON" {
		opt.Serializer = json.NewSerializer()
		opt.SerializerName = "JSON"

	} else {
		opt.Serializer = protobuf.NewSerializer()
		opt.SerializerName = "Protobuf"
	}
	return nil
}

// setCommand 自定义命令
func setCommand(shell *ishell.Shell, args []string) error {
	cmdName := strings.TrimSpace(args[0])
	switch currentAccount.CurrentSetType {
	case LOCAL:
		shell.DeleteCmd(cmdName)
		// 如果第一个是文件名 那么去读文件  否则读参数里的命令
		if exists(args[1]) {
			f, err := readFile(args[1])
			if err != nil {
				logger.Println("open file:%s error:%v", args[1], err)
				return err
			}
			defer f.Close()
			allSubCmds := make([]string, 0)
			buf := bufio.NewReader(f)
			for {
				line, err := buf.ReadString('\n')
				line = strings.TrimSpace(line)
				allSubCmds = append(allSubCmds, line)
				if err != nil {
					if err == io.EOF {
						break
					}
					logger.Println(err)
					return err
				}
			}
			cmdStr := strings.Join(allSubCmds, ";")
			err = saveCmdsToFile(cmdName, cmdStr)
			if err != nil {
				return err
			}
			addCustomCommand(shell, allSubCmds, cmdName, cmdStr)
		} else {
			cmds := args[1:]
			cmdStr := strings.Join(cmds, ";")
			err := saveCmdsToFile(cmdName, cmdStr)
			if err != nil {
				return err
			}

			addCustomCommand(shell, cmds, cmdName, cmdStr)
		}
	case ACCOUNT:
		cmds := args[1:]
		cmdStr := strings.Join(cmds, ";")
		currentSet := currentAccount.CurrentSet
		_, ok := currentAccount.CmdSets[currentSet]
		if !ok {
			currentAccount.CmdSets[currentSet] = make(map[string]string)
		}
		currentAccount.CmdSets[currentSet][cmdName] = cmdStr
		err := currentAccount.Save()
		if err != nil {
			return err
		}
		addCustomCommand(shell, cmds, cmdName, cmdStr)
	}
	return nil
}

// delCommand 删除当前使用的命令集里的一条命令
func delCommand(cmdName string) error {
	err := checkLoginCli()
	if err != nil {
		return err
	}
	currentSet := currentAccount.CurrentSet
	switch currentAccount.CurrentSetType {
	case LOCAL:
		currentPath = filepath.Join(cmdDir, currentSet)
		f, err := writeFile(currentPath)
		if err != nil {
			return err
		}
		defer f.Close()
		buf := bufio.NewReader(f)
		// 在delCommand的时候 顺便做一个整理 setCommand了多次的命令只保留最后一个
		allCmds := make(map[string]string)
		for {
			line, err := buf.ReadString('\n')
			line = strings.TrimSpace(line)
			lineSplits := strings.Split(line, "@")
			if len(lineSplits) == 2 && lineSplits[0] != cmdName {
				allCmds[lineSplits[0]] = lineSplits[1]
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}
		var strBuf bytes.Buffer
		for name, cmd := range allCmds {
			strBuf.WriteString(name)
			strBuf.WriteString("@")
			strBuf.WriteString(cmd)
			strBuf.WriteString("\n")
		}
		if _, err := f.WriteString(strBuf.String()); err != nil {
			panic(err)
		}
		return f.Sync()
	case ACCOUNT:
		_, ok := currentAccount.CmdSets[currentSet]
		if !ok {
			return fmt.Errorf("command set:%s not exist in account command sets", currentSet)
		}
		delete(currentAccount.CmdSets[currentSet], cmdName)
		err := currentAccount.Save()
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

// upload uploads local command config to remote
func upload(localName, remoteName string) error {
	getRedisClient, err := getRedisClient()
	if err != nil {
		return fmt.Errorf("connect to redis error: %v", err)
	}
	cmdPath := getCmdFile(localName)
	if !exists(cmdPath) {
		return fmt.Errorf("local command set %v does not exist", localName)
	}
	srcFile, err := readFile(cmdPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	srcData := make([]byte, 1024*1024) //最大1M
	length, err := srcFile.Read(srcData)
	if err == io.EOF {
		return fmt.Errorf("empty local command set %v", localName)
	} else if err != nil {
		return err
	}
	_, err = getRedisClient.HSet(cliHashKey, remoteName, string(srcData[:length])).Result()
	if err != nil {
		return err
	}
	logger.Printf("upload success, %d bytes transferred\n", length)
	return nil
}

// listAll show all command sets
func listAll() error {
	logger.Println("local:")
	err := listLocal()
	if err != nil {
		return err
	}
	logger.Println("account:")
	err = listAccount()
	if err != nil {
		return err
	}

	logger.Println("remote:")
	err = listRemote()
	if err != nil {
		return err
	}
	return nil
}

// listLocal show all local command sets
func listLocal() error {
	if currentAccount == nil {
		return fmt.Errorf("current not login cli, please run loginCli first")
	}
	currentSet := currentAccount.CurrentSet
	err := filepath.Walk(cmdDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			var prefix string
			if info.Name() == currentSet && currentAccount.CurrentSetType == LOCAL {
				prefix = "*   "
			} else {
				prefix = "    "
			}
			logger.Printf("%s%s\t%s\t%d bytes\n", prefix, info.Name(), info.ModTime().Format(time.UnixDate), info.Size())
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// listRemote shows all command sets in remote
func listRemote() error {
	getRedisClient, err := getRedisClient()
	if err != nil {
		return fmt.Errorf("connect to redis error: %v", err)
	}
	result, err := getRedisClient.HGetAll(cliHashKey).Result()
	if err != nil {
		return err
	}
	for k, v := range result {
		logger.Printf("    %s\t%d bytes\n", k, len(v))
	}
	return nil
}

// listAccount list command sets bound to current account
func listAccount() error {
	err := checkLoginCli()
	if err != nil {
		return err
	}
	currentSet := currentAccount.CurrentSet
	if currentAccount.CmdSets != nil {
		for cmdName, cmd := range currentAccount.CmdSets {
			var prefix string
			if cmdName == currentSet && currentAccount.CurrentSetType == ACCOUNT {
				prefix = "*   "
			} else {
				prefix = "    "
			}
			logger.Printf("%s%s\t%d bytes\n", prefix, cmdName, calCmdSetLength(cmd))
		}
	}
	return nil
}

// download downloads command set from remote to local
func download(remoteName, localName string) error {
	getRedisClient, err := getRedisClient()
	if err != nil {
		return fmt.Errorf("connect to redis error: %v", err)
	}

	exists, err := getRedisClient.HExists(cliHashKey, remoteName).Result()
	if err != nil {
		return err
	}
	if !exists {
		return errors.New(remoteName + " does not exist in remote")
	}

	result, err := getRedisClient.HGet(cliHashKey, remoteName).Result()
	if err != nil {
		return err
	}

	cmdPath := getCmdFile(localName)
	f, err := newFile(cmdPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(result); err != nil {
		panic(err)
	}
	if err := f.Sync(); err != nil {
		panic(err)
	}

	logger.Printf("download success, %d bytes transferred\n", len(result))
	return nil
}

// removeLocal 删除本地命令集的文件
func removeLocal(name string) error {
	cmdPath := getCmdFile(name)
	if err := removeFile(cmdPath); err != nil {
		return err
	}
	return nil
}

// removeRemote 删除远程命令集
func removeRemote(name string) error {
	getRedisClient, err := getRedisClient()
	if err != nil {
		return fmt.Errorf("connect to redis error: %v", err)
	}
	exists, err := getRedisClient.HExists(cliHashKey, name).Result()
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%s does not exist in remote", name)
	}
	_, err = getRedisClient.HDel(cliHashKey, name).Result()
	if err != nil {
		return err
	}
	return nil
}

// removeAccount remove command set bound to account
func removeAccount(name string) error {
	err := checkLoginCli()
	if err != nil {
		return err
	}
	_, ok := currentAccount.CmdSets[name]
	if ok {
		delete(currentAccount.CmdSets, name)
		err := currentAccount.Save()
		if err != nil {
			return err
		}
		return nil
	} else {
		return fmt.Errorf("command set:%s not exist in account command sets", name)
	}
}

// use select a command set as current use command set
// if command set not exist in local, it will automatic download from remote
func use(shell *ishell.Shell, name string, typ int) error {
	err := checkLoginCli()
	if err != nil {
		return err
	}
	// 这种情况在当前使用的命令集被remove的时候 会出现
	if currentAccount.CurrentSet != "" {
		err := unloadCurrentCommands(shell)
		if err != nil {
			return err
		}
	}
	err = loadCustomCommands(shell, name, typ)

	if err != nil {
		return err
	}

	if currentAccount != nil {
		currentAccount.CurrentSet = name
		currentAccount.CurrentSetType = typ
		err := currentAccount.Save()
		if err != nil {
			return err
		}
	}

	return nil
}

func useNewAccount(shell *ishell.Shell, name string) error {
	cmdSets := currentAccount.CmdSets
	_, ok := cmdSets[name]
	if ok {
		return fmt.Errorf("command set:%s already exists in account command sets", name)
	} else {
		cmdSets[name] = make(map[string]string)
		return currentAccount.Save()
	}
}

// useNewLocal new一个本地命令集
func useNewLocal(shell *ishell.Shell, name string) error {
	cmdPath := getCmdFile(name)
	if exists(cmdPath) {
		return fmt.Errorf("%s already exists, can not create new", name)
	}
	f, err := newFile(cmdPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Sync(); err != nil {
		panic(err)
	}
	err = unloadCurrentCommands(shell)
	if err != nil {
		return err
	}

	err = setCurrentCommandSet(name)
	if err != nil {
		return err
	}
	return nil
}

// alias 取别名
func alias(shell *ishell.Shell, cmdName, alias string) (bool, error) {
	if currentAccount == nil {
		return false, fmt.Errorf("current not login cli, please run loginCli first")
	}
	allCmds := shell.Cmds()
	for _, cmd := range allCmds {
		if cmd.Name == cmdName {
			exist := false
			for _, als := range cmd.Aliases {
				if als == alias {
					exist = true
					break
				}
			}
			if !exist {
				cmd.Aliases = append(cmd.Aliases, alias)
				if currentAccount.Aliases[cmdName] == nil {
					currentAccount.Aliases[cmdName] = make([]string, 0)
				}
				currentAccount.Aliases[cmdName] = append(currentAccount.Aliases[cmdName], alias)
				err := currentAccount.Save()
				if err != nil {
					return false, err
				}
			}
			return true, nil
		}
	}
	return false, nil
}

// loginCli cli登录，启动时登录，且必须登录，才有account类型的命令集
func loginCli(shell *ishell.Shell, username string) error {
	currentAccount = &Account{
		Username:       username,
		Aliases:        make(map[string][]string),
		CmdSets:        make(map[string]map[string]string),
		CurrentSet:     "default",
		CurrentSetType: LOCAL,
	}
	isNew, err := currentAccount.Load()
	if err != nil {
		return err
	}
	if err = saveUsername(username); err != nil {
		return err
	}
	if !isNew {
		loadAlias(shell, currentAccount.Aliases)
	}
	return nil
}

// sync 同步命令集(在local,account,remote三种类型中进行命令集的同步)
func sync(shell *ishell.Shell, src, srcCmd, dst, dstCmd string) error {
	err := checkLoginCli()
	if err != nil {
		return err
	}
	switch src {
	case "local":
		cmdPath := getCmdFile(srcCmd)
		if !exists(cmdPath) {
			return fmt.Errorf("cmd set named:%s not exist in local", srcCmd)
		}
		srcFile, err := readFile(cmdPath)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		var cmdStrBuf bytes.Buffer
		var cmds = make(map[string]string)
		buf := bufio.NewReader(srcFile)
		for {
			line, err := buf.ReadString('\n')
			line = strings.TrimSpace(line)
			lineSplits := strings.Split(line, "@")
			if len(lineSplits) == 2 {
				cmdName := strings.TrimSpace(lineSplits[0])
				cmdStr := strings.TrimSpace(lineSplits[1])
				cmds[cmdName] = cmdStr
			}
			cmdStrBuf.WriteString(line)
			cmdStrBuf.WriteString("\n")
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}
		switch dst {
		case "account":
			currentAccount.CmdSets[dstCmd] = cmds
			err := currentAccount.Save()
			if err != nil {
				return err
			}
		case "remote":
			getRedisClient, err := getRedisClient()
			if err != nil {
				return fmt.Errorf("connect to redis error: %v", err)
			}
			_, err = getRedisClient.HSet(cliHashKey, dstCmd, cmdStrBuf.String()).Result()
			if err != nil {
				return err
			}
		}
	case "account":
		cmdSet, ok := currentAccount.CmdSets[srcCmd]
		if !ok {
			return fmt.Errorf("cmd set named:%s not exist in account", srcCmd)
		}
		var cmdStrBuf bytes.Buffer
		for k, v := range cmdSet {
			cmdStrBuf.WriteString(k)
			cmdStrBuf.WriteString("@")
			cmdStrBuf.WriteString(v)
			cmdStrBuf.WriteString("\n")
		}
		switch dst {
		case "local":
			cmdPath := getCmdFile(dstCmd)
			dstFile, err := newFile(cmdPath)
			if err != nil {
				return err
			}
			defer dstFile.Close()
			_, err = dstFile.WriteString(cmdStrBuf.String())
			if err != nil {
				return err
			}
			return dstFile.Sync()
		case "remote":
			getRedisClient, err := getRedisClient()
			if err != nil {
				return fmt.Errorf("connect to redis error: %v", err)
			}
			_, err = getRedisClient.HSet(cliHashKey, dstCmd, cmdStrBuf.String()).Result()
			if err != nil {
				return err
			}
		}
	case "remote":
		getRedisClient, err := getRedisClient()
		if err != nil {
			return fmt.Errorf("connect to redis error: %v", err)
		}
		cmdStr, err := getRedisClient.HGet(cliHashKey, srcCmd).Result()
		if err == redis.Nil {
			return fmt.Errorf("cmd set named:%s not exist in remote", srcCmd)
		} else if err != nil {
			return err
		}
		switch dst {
		case "local":
			cmdPath := getCmdFile(dstCmd)
			dstFile, err := newFile(cmdPath)
			if err != nil {
				return err
			}
			defer dstFile.Close()
			_, err = dstFile.WriteString(cmdStr)
			if err != nil {
				return err
			}
			return dstFile.Sync()
		case "account":
			cmdPairs := strings.Split(cmdStr, "\n")
			cmds := make(map[string]string)
			for _, cmdPair := range cmdPairs {
				cmdSplits := strings.Split(cmdPair, "@")
				if len(cmdSplits) == 2 {
					cmds[cmdSplits[0]] = cmdSplits[1]
				}
			}
			currentAccount.CmdSets[dstCmd] = cmds
			err := currentAccount.Save()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
