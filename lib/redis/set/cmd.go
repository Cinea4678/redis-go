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
	{Name: "sadd", RedisClientFunc: SAdd},
	{Name: "sismember", RedisClientFunc: SIsMember},
	{Name: "smembers", RedisClientFunc: SMembers},
	{Name: "scard", RedisClientFunc: SCard},
	{Name: "srem", RedisClientFunc: SRem},
}

var SetCommandInfoTable = []*core.RedisCommandInfo{
	core.NewRedisCommandInfo("sadd", -3, []string{"write", "denyoom"}, 1, 1, 1),
	core.NewRedisCommandInfo("sismember", 3, []string{"readonly"}, 1, 2, 1),
	core.NewRedisCommandInfo("smembers", 2, []string{"readonly"}, 1, 1, 1),
	core.NewRedisCommandInfo("scard", 2, []string{"readonly"}, 1, 1, 1),
	core.NewRedisCommandInfo("srem", 3, []string{"write"}, 1, 2, 1),
}
