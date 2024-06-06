package resistence

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/shared"
	"strings"
	"sync"
	"time"

	"github.com/cinea4678/resp3"
)

var (
	aofFile   *os.File
	aofMutex  sync.Mutex
	aofBuffer []string
)

var (
	errCommandUnknown = errors.New("command unknown")
)

// 空白占位符，使用map实现集合
var ept = struct{}{}

// 记录需要持久化的写指令
var aofCommands = map[string]struct{}{
	// str
	"set":    ept,
	"incr":   ept,
	"incrby": ept,
	"decr":   ept,
	"decrby": ept,
	"append": ept,

	// list
	"lpush": ept,
	"lpop":  ept,
	"rpush": ept,
	"rpop":  ept,

	//set
	"sadd": ept,
	"srem": ept,

	//zset
	"zadd":             ept,
	"zincrby":          ept,
	"zrem":             ept,
	"zremrangebyscore": ept,

	// Add other write-related commands here
}

func NeedAOF(command string) bool {
	if _, ok := aofCommands[command]; ok {
		return true
	}
	return false
}

func InitAOF(filePath string) error {
	var err error
	aofFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	aofBuffer = make([]string, 0)
	go flushAOFBufferPeriodically()

	return nil
}

// parseCommandString 解析命令字符串
func parseCommandString(commandStr string) (*core.RedisCommand, []string, error) {
	parts := strings.Fields(commandStr)
	if len(parts) == 0 {
		return nil, nil, errCommandUnknown
	}

	cmdName := strings.ToLower(parts[0])
	cmd := lookupCommand(cmdName)
	if cmd == nil {
		return nil, nil, errCommandUnknown
	}

	return cmd, parts[1:], nil
}

func lookupCommand(name string) *core.RedisCommand {
	name = strings.ToLower(name) // 转换为小写
	cmd := shared.Server.Commands.DictFind(name)
	if cmd == nil {
		return nil
	}
	return cmd.(*core.RedisCommand)
}

// executeCommand 从AOF文件执行命令
func executeCommand(commandStr string) error {
	cmd, args, err := parseCommandString(commandStr)
	if err != nil {
		return err
	}

	client := &core.RedisClient{
		ReqValue: &resp3.Value{
			Elems: make([]*resp3.Value, len(args)+1),
		},
		Db:    shared.Server.Db[0],
		IsAOF: true,
	}

	client.ReqValue.Elems[0] = &resp3.Value{Str: cmd.Name}
	for i, arg := range args {
		client.ReqValue.Elems[i+1] = &resp3.Value{Str: arg}
	}

	fmt.Println(cmd)
	client.Cmd = cmd
	client.LastCmd = cmd

	return call(client, 0)
}

func call(client *core.RedisClient, _ int) error {
	return client.Cmd.RedisClientFunc(client)
}

func LoadAOF(filePath string) (err error) {
	file, err := os.Open(filePath)
	fmt.Println("open AOF file successfully: ", file.Name())
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		command := scanner.Text()
		fmt.Println("command: ", command)
		// 执行指令
		if err := executeCommand(command); err != nil {
			return err
		}
	}
	return nil
	// return scanner.Err()
}

func AddToAOFBuffer(commandValue *resp3.Value) {

	strbuilder := strings.Builder{}

	for _, elem := range commandValue.Elems {
		cmd := elem.Str + " "
		strbuilder.WriteString(cmd)
	}

	command := strings.TrimSpace(strbuilder.String())

	aofMutex.Lock()
	defer aofMutex.Unlock()

	aofBuffer = append(aofBuffer, command)
	if len(aofBuffer) >= shared.AOFBuffer { // 缓冲区达到1000条命令时刷新
		flushAOFBuffer()
	}
}

func flushAOFBuffer() (err error) {
	if len(aofBuffer) == 0 {
		return
	}

	for _, command := range aofBuffer {
		_, err := aofFile.WriteString(command + "\n")
		if err != nil {
			return err
		}
	}

	aofBuffer = aofBuffer[:0] // Clear the buffer
	return
}

func flushAOFBufferPeriodically() (err error) {
	for {
		time.Sleep(shared.AOFInterval)
		aofMutex.Lock()
		// fmt.Println(aofBuffer)
		err := flushAOFBuffer()
		if err != nil {
			fmt.Println(err)
		}
		aofMutex.Unlock()
	}
}

func CloseAOF() error {
	aofMutex.Lock()
	defer aofMutex.Unlock()

	flushAOFBuffer()
	return aofFile.Close()
}

// package resistence

// import (
// 	"os"
// 	"sync"
// )

// var aofFile *os.File
// var aofMutex sync.Mutex

// func InitAOF(filePath string) error {
// 	var err error
// 	aofFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	return err
// }

// func WriteToAOF(command string) error {
// 	aofMutex.Lock()
// 	defer aofMutex.Unlock()

// 	_, err := aofFile.WriteString(command + "\n")
// 	return err
// }

// func CloseAOF() error {
// 	return aofFile.Close()
// }
