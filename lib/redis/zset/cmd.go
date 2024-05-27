package zset

import (
	"errors"
	"redis-go/lib/redis/core"
)

var (
	errNotEnoughArgs = errors.New("not enough args")
	errNoScore       = errors.New("no score provided")
	errScoreNotFound = errors.New("score not found")
	errValueNotFound = errors.New("value not found")
	errValueExists   = errors.New("value already exists")
	errInvalidRange  = errors.New("invalid range")
)

// ZSetCommandTable 有序集合相关命令
var ZSetCommandTable = []*core.RedisCommand{
	// {"zadd", ZAdd},
	// {"zcard", ZCard},
	// {"zcount", ZCount},
	// {"zincrby", ZIncrBy},
	// {"zinterstore", ZInterStore},
	// {"zlexcount", ZLexCount},
	// {"zrange", ZRange},
	// {"zrangebylex", ZRangeByLex},
	// {"zrangebyscore", ZRangeByScore},
	// {"zrank", ZRank},
	// {"zrem", ZRem},
	// {"zremrangebylex", ZRemRangeByLex},
	// {"zremrangebyrank", ZRemRangeByRank},
	// {"zremrangebyscore", ZRemRangeByScore},
	// {"zrevrange", ZRevRange},
	// {"zrevrangebyscore", ZRevRangeByScore},
	// {"zrevrank", ZRevRank},
	// {"zscore", ZScore},
	// {"zunionstore", ZUnionStore},
	// {"zscan", ZScan},
}
