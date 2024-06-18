package system

import (
	"errors"
	"redis-go/lib/redis/core"
)

var errEchoNoMessage = errors.New("no message given")

// CommandTable 系统工具类的命令
var CommandTable = []*core.RedisCommand{
	{Name: "ping", RedisClientFunc: Ping},
	{Name: "time", RedisClientFunc: Time},
	{Name: "echo", RedisClientFunc: Echo},
	{Name: "hello", RedisClientFunc: Hello},
	{Name: "info", RedisClientFunc: Info},
	{Name: "prof", RedisClientFunc: Prof},
}

var CommandInfoTable = []*core.RedisCommandInfo{
	core.NewRedisCommandInfo("ping", 1, []string{"fast"}, 0, 0, 0),
	core.NewRedisCommandInfo("time", 1, []string{"loading", "fast"}, 0, 0, 0),
	core.NewRedisCommandInfo("echo", 2, []string{"loading", "fast"}, 0, 0, 0),
	core.NewRedisCommandInfo("hello", 1, []string{"fast"}, 0, 0, 0),
	core.NewRedisCommandInfo("info", 1, []string{"fast"}, 0, 0, 0),
	core.NewRedisCommandInfo("prof", 1, []string{"fast"}, 0, 0, 0),
}
