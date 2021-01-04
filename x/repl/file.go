package repl

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lonng/nano/log"
	"github.com/mitchellh/go-homedir"
)

// ExecuteFromFile execute from file which contains a sequence of command
func ExecuteFromFile(fileName string) {
	var err error
	defer func() {
		if err != nil {
			log.Printf("error: %s", err.Error())
		}
	}()

	var file *os.File
	file, err = readFile(fileName)
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		command := scanner.Text()
		err = executeCommand(command, 1)
		if err != nil {
			return
		}
	}

	err = scanner.Err()
	if err != nil {
		return
	}
}

// executeCommand 按顺序执行命令；两个地方用到：1.读文件然后执行命令 2.执行setCommand自定义的命令
func executeCommand(command string, repeat int64) error {
	parts := strings.Split(command, " ")
	switch parts[0] {
	case "connect":
		return connect(parts[1], func(route string, data []byte) {
			log.Printf("server-> %s:%s\n", route, string(data))
		})

	case "request":
		for i := int64(0); i < repeat; i++ {
			if err := request(parts[1:]); err != nil {
				return err
			}
			time.Sleep(200 * time.Millisecond)
		}
		return nil

	case "notify":
		return notify(parts[1:])

	case "disconnect":
		if pClient == nil {
			log.Println("already disconnected")
			return nil
		}
		disconnect()
		log.Println("disconnected")
		return nil

	case "history":
		return history(parts[1:])

	case "clearHistory":
		return clearHistory()

	case "setSerializer":
		return setSerializer(parts[1])

	default:
		return errors.New("command not found")
	}
}

const (
	dataDirName      string = ".nanocli"
	cmdDirName       string = "cmd"
	historyFileName  string = "history"
	currentFileName  string = "current"
	usernameFileName string = "username"
)

var (
	dataDir      string
	cmdDir       string
	historyPath  string
	currentPath  string
	usernamePath string
)

func init() {
	logger = NewCliLogger(nil)
	log.SetLogger(logger)
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	dataDir = filepath.Join(home, dataDirName)
	if !exists(dataDir) {
		if err := os.Mkdir(dataDir, 0644); err != nil {
			panic(err)
		}
	}

	cmdDir = filepath.Join(dataDir, cmdDirName)
	if !exists(cmdDir) {
		if err := os.Mkdir(cmdDir, 0644); err != nil {
			panic(err)
		}
	}

	historyPath = filepath.Join(dataDir, historyFileName)
	currentPath = filepath.Join(dataDir, currentFileName)
	usernamePath = filepath.Join(dataDir, usernameFileName)
}

func readFile(f string) (*os.File, error) {
	return os.OpenFile(f, os.O_CREATE|os.O_RDONLY, 0644)
}

func writeFile(f string) (*os.File, error) {
	return os.OpenFile(f, os.O_WRONLY|os.O_CREATE, 0644)
}

func newFile(f string) (*os.File, error) {
	return os.OpenFile(f, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
}

func appendFile(f string) (*os.File, error) {
	return os.OpenFile(f, os.O_APPEND|os.O_CREATE, 0644)
}

func removeFile(f string) error {
	if exists(f) {
		return os.Remove(f)
	}
	return nil
}

func getCurrentCmdFile() string {
	if currentAccount != nil {
		currentSet := currentAccount.CurrentSet
		return filepath.Join(cmdDir, currentSet)
	} else {
		currentSet, err := getCurrentCommandSet()
		if err != nil {
			panic(err)
		}
		return filepath.Join(cmdDir, currentSet)
	}
}

func getCmdFile(name string) string {
	return filepath.Join(cmdDir, name)
}

func exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		return os.IsExist(err)
	}
	return true
}
