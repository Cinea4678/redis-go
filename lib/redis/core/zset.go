package core

import (
	"redis-go/lib/redis/core/zset"
)

type ZSet = zset.ZSet

func NewZSet() *zset.ZSet {
	return zset.NewZSet()
}
