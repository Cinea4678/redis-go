package list

import (
	"errors"
	"redis-go/lib/redis/core"
)

var (
	errNotEnoughArgs   = errors.New("not enough args")
	errTooManyArgs     = errors.New("too many arguments")
	errNotAList        = errors.New("key not a list")
	errInvalidIndex    = errors.New("invalid index")
	errIndexOutOfRange = errors.New("index out of range")
)

var ListCommandTable = []*core.RedisCommand{
	{"lpush", LPush},
	{"rpush", RPush},
	{"lpop", LPop},
	{"rpop", RPop},
	{"lindex", LIndex},
	{"lrange", LRange},
}
