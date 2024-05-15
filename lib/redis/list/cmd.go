package list

import (
	"errors"
	"redis-go/lib/redis/core"
)

var (
	errNotEnoughArgs = errors.New("not enough args")
	errNotAList      = errors.New("key not a list")
)

var ListCommandTable = []*core.RedisCommand{
	{"lpush", LPush},
	{"rpush", RPush},
	{"lpop", LPop},
	{"rpop", RPop},
	{"lindex", LIndex},
	{"lrange", LRange},
}
