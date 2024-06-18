package json

import "redis-go/lib/redis/core"

var JsonCommandTable = []*core.RedisCommand{
	{Name: "jget", RedisClientFunc: JGet},
}

var JsonCommandInfoTable = []*core.RedisCommandInfo{
	core.NewRedisCommandInfo("jget", 3, []string{"readonly"}, 1, 1, 1),
}
