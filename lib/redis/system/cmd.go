package system

import (
	"errors"
	"redis-go/lib/redis/core"
)

var errEchoNoMessage = errors.New("no message given")

// CommandTable 系统工具类的命令
var CommandTable = []*core.RedisCommand{
	{"ping", Ping},
	{"time", Time},
	{"echo", Echo},
	{"hello", Hello},
	{"info", Info},
}
