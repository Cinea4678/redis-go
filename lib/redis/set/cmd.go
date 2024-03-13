package set

import (
	"errors"
	"redis-go/lib/redis/core"
)

var (
	errNotEnoughArgs = errors.New("not enough args")
	errNotASet       = errors.New("key not a set")
)

var SetCommandTable = []*core.RedisCommand{
	{"sadd", SAdd},
	{"sismember", SIsMember},
	{"smembers", SMembers},
}
