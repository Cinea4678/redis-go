package zset

import (
	"errors"
	"redis-go/lib/redis/core"
)

var (
	errNotEnoughArgs = errors.New("not enough args")
	errInvalidArgs   = errors.New("invalid args")
	errZSetNotFound  = errors.New("zset not found")
	errNotAZSet      = errors.New("the key is not a zset")
	errNoScore       = errors.New("no score provided")
	errScoreNotFound = errors.New("score not found")
	errValueNotFound = errors.New("value not found")
	errValueExists   = errors.New("value already exists")
	errInvalidRange  = errors.New("invalid range")
)

// ZSetCommandTable 有序集合相关命令
var ZSetCommandTable = []*core.RedisCommand{
	{"zadd", ZAdd},
	{"zcard", ZCard},
	{"zcount", ZCount},
	{"zincrby", ZIncrBy},
	// {"zinterstore", ZInterStore},
	// {"zlexcount", ZLexCount},
	{"zrange", ZRange},
	// {"zrangebylex", ZRangeByLex},
	{"zrangebyscore", ZRangeByScore},
	// {"zrank", ZRank},
	{"zrem", ZRem},
	// {"zremrangebylex", ZRemRangeByLex},
	// {"zremrangebyrank", ZRemRangeByRank},
	// {"zremrangebyscore", ZRemRangeByScore},
	// {"zrevrange", ZRevRange},
	// {"zrevrangebyscore", ZRevRangeByScore},
	// {"zrevrank", ZRevRank},
	{"zscore", ZScore},
	// {"zunionstore", ZUnionStore},
	// {"zscan", ZScan},
}

var ZSetCommandInfoTable = []*core.RedisCommandInfo{
	core.NewRedisCommandInfo("zadd", -4, []string{"write", "denyoom"}, 1, 1, 1),
	core.NewRedisCommandInfo("zcard", 2, []string{"readonly"}, 1, 1, 1),
	core.NewRedisCommandInfo("zcount", 4, []string{"readonly"}, 1, 1, 1),
	core.NewRedisCommandInfo("zincrby", 4, []string{"write"}, 1, 1, 1),
	core.NewRedisCommandInfo("zrangebyscore", -4, []string{"readonly"}, 1, 1, 1),
	core.NewRedisCommandInfo("zrem", -3, []string{"write"}, 1, 1, 1),
	core.NewRedisCommandInfo("zscore", 3, []string{"readonly"}, 1, 1, 1),
	core.NewRedisCommandInfo("zrank", 3, []string{"readonly"}, 1, 1, 1),
	core.NewRedisCommandInfo("zrange", -4, []string{"readonly"}, 1, 1, 1),
}
