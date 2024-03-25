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
)

// StringsCommandTable 字符串相关命令
var StringsCommandTable = []*core.RedisCommand{
	{"set", Set},
	{"get", Get},
	{"incr", Increase},
	{"incrby", IncreaseBy},
	{"decr", Decrease},
	{"decrby", DecreaseBy},
	{"append", Append},
}
