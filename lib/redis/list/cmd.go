package list

import (
	"errors"
	"redis-go/lib/redis/core"
)

var (
	errNotEnoughArgs   = errors.New("not enough args")
	errNotAList        = errors.New("key not a list")
	errInvalidIndex    = errors.New("invalid index")
	errIndexOutOfRange = errors.New("index out of range")
	errKeyNotFound     = errors.New("key not found")
	errUnknown         = errors.New("unknown error")
)

var ListCommandTable = []*core.RedisCommand{
	{"lpush", LPush},
	{"rpush", RPush},
	{"lpop", LPop},
	{"rpop", RPop},
	{"lindex", LIndex},
	{"lrange", LRange},
}
