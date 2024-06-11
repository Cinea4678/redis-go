package str

import (
	"errors"
	"redis-go/lib/redis/core"
)

var (
	errNotEnoughArgs       = errors.New("not enough args")
	errNoKey               = errors.New("no key provided")
	errNotEnoughArgsExpire = errors.New("not enough args after EX, PX, EXAT or PXAT")
	errNXAndXXConflict     = errors.New("conflict arg: NX and XX")
	errExpiresConflict     = errors.New("conflict arg: Only one of EX, PX, EXAT or PXAT can exist at the same time")
	errInvalidInt          = errors.New("value is not an integer or out of range")
)

// StringsCommandTable 字符串相关命令
var StringsCommandTable = []*core.RedisCommand{
	{Name: "set", RedisClientFunc: Set},
	{Name: "get", RedisClientFunc: Get},
	{Name: "incr", RedisClientFunc: Increase},
	{Name: "incrby", RedisClientFunc: IncreaseBy},
	{Name: "decr", RedisClientFunc: Decrease},
	{Name: "decrby", RedisClientFunc: DecreaseBy},
	{Name: "append", RedisClientFunc: Append},
	{Name: "getrange", RedisClientFunc: GetRange},
	{Name: "getdel", RedisClientFunc: GetDel},
	{Name: "lcs", RedisClientFunc: LCS},
}

var StringsCommandInfoTable = []*core.RedisCommandInfo{
	core.NewRedisCommandInfo("set", -3, []string{"write", "denyoom"}, 1, 1, 1),
	core.NewRedisCommandInfo("get", -2, []string{"readonly"}, 1, 1, 1),
	core.NewRedisCommandInfo("incr", 2, []string{"write"}, 1, 1, 1),
	core.NewRedisCommandInfo("incrby", 3, []string{"write"}, 1, 1, 1),
	core.NewRedisCommandInfo("decr", 2, []string{"write"}, 1, 1, 1),
	core.NewRedisCommandInfo("decrby", 3, []string{"write"}, 1, 1, 1),
	core.NewRedisCommandInfo("append", -3, []string{"write", "denyoom"}, 1, 1, 1),
	core.NewRedisCommandInfo("getrange", 4, []string{"readonly"}, 1, 1, 1),
	core.NewRedisCommandInfo("getdel", 2, []string{"write"}, 1, 1, 1),
	core.NewRedisCommandInfo("lcs", -3, []string{"readonly"}, 1, 2, 1),
}
