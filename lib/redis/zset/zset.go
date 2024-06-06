package zset

import (
	"fmt"
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/core/zset"
	"redis-go/lib/redis/io"
	"redis-go/lib/redis/shared"
	"strconv"
	"strings"

	"github.com/cinea4678/resp3"
)

type ZSet = zset.ZSet

var (
	ZSetOk  = zset.ZSetOk
	ZSetErr = zset.ZSetErr
)

// ZAdd - Add one or more members to a sorted set, or update its score if it already exists
func ZAdd(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 2 {
		return errNotEnoughArgs
	}

	// 例如ZAdd myzset 0.1 value1 0.2这种缺失了value值的
	if len(req)%2 != 1 {
		return errInvalidArgs
	}

	zsetKey := req[0].Str

	db := client.Db

	var zs *zset.ZSet

	// 如果默认ZSet没有创建
	if zsetObj := db.LookupKey(zsetKey); zsetObj == nil {
		zs = zset.NewZSet()
		db.DbAdd(zsetKey, core.CreateZSet(zs))
	} else {
		zs = zsetObj.Ptr.(*zset.ZSet)
	}

	for i := 1; i < len(req); i += 2 {
		score, err := strconv.ParseFloat(req[i].Str, 64)
		if err != nil {
			return errInvalidArgs
		}

		value := req[i+1].Str

		fmt.Println(score, value)
		_, exist := zs.ZSetGetScore(value)
		if exist {
			// fmt.Println("ZSetUpdate", value, score, s)
			zs.ZSetUpdate(score, value)
		} else {
			// fmt.Println("ZSetAdd", value, score, s)
			zs.ZSetAdd(score, value)
		}

		io.AddReplyDouble(client, score)
	}

	return
}

// ZCard - Get the number of members in a sorted set
func ZCard(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 1 {
		return errNotEnoughArgs
	}

	db := client.Db
	zsetKey := req[0].Str

	zsetObj := db.LookupKey(zsetKey)
	if zsetObj == nil {
		return errZSetNotFound
	} else if zsetObj.Type != core.RedisZSet {
		return errNotAZSet
	}

	zs := zsetObj.Ptr.(*zset.ZSet)
	io.AddReplyNumber(client, int64(zs.Len()))

	return
}

func ZCount(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 3 {
		return errNotEnoughArgs
	}

	db := client.Db
	zsetKey := req[0].Str

	minScore, err := strconv.ParseFloat(req[1].Str, 64)
	if err != nil {
		return errInvalidArgs
	}
	maxScore, err := strconv.ParseFloat(req[2].Str, 64)
	if err != nil {
		return errInvalidArgs
	}

	zsetObj := db.LookupKey(zsetKey)
	if zsetObj == nil {
		io.AddReplyString(client, zsetKey+"not found")
		return errZSetNotFound
	}

	zs := zsetObj.Ptr.(*zset.ZSet)
	res := zs.ZSetSearchRange(minScore, maxScore)
	count := len(res)
	io.AddReplyNumber(client, int64(count))
	return
}

func ZIncrBy(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]
	if len(req) < 3 {
		return errNotEnoughArgs
	}

	key := req[0].Str
	increment := req[1].Double
	member := req[2].Str
	db := client.Db

	zsetObj := db.LookupKey(key)

	if zsetObj == nil {
		zset := zset.NewZSet()
		db.DbAdd(key, core.CreateZSet(zset))
		zset.ZSetAdd(increment, member)
		io.AddReplyDouble(client, increment)
	} else if zsetObj.Type != core.RedisZSet {
		return errNotAZSet
	}

	zs := zsetObj.Ptr.(*ZSet)
	oldScore, exists := zs.ZSetGetScore(member)
	if exists {
		newScore := oldScore + increment
		zs.ZSetUpdate(newScore, member) // 更新分数
		io.AddReplyDouble(client, newScore)
	} else {
		io.AddReplyDouble(client, increment)
	}

	return
}

// ZRange - 获取指定区间的成员
func ZRange(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 3 {
		return errNotEnoughArgs
	}

	db := client.Db
	zsetKey := req[0].Str

	start, err := strconv.Atoi(req[1].Str)
	if err != nil {
		return err
	}
	stop, err := strconv.Atoi(req[2].Str)
	if err != nil {
		return err
	}

	var withscores bool
	if len(req) > 3 {
		w := req[3].Str
		withscores = (strings.ToUpper(w) == "WITHSCORES")
	}

	zsetObj := db.LookupKey(zsetKey)
	if zsetObj == nil {
		io.SendReplyToClient(client, shared.Shared.Nil)
		return
	}

	zs := zsetObj.Ptr.(*ZSet)
	members := zs.ZSetSearchRankRange(start, stop)

	if withscores {
		results := make([]*resp3.Value, 2*len(members))
		for i, member := range members {
			results[2*i] = resp3.NewSimpleStringValue(member.Value)
			score, _ := zs.ZSetGetScore(member.Value)
			results[2*i+1] = resp3.NewDoubleValue(score)
		}
		io.AddReplyArray(client, results)
	} else {
		results := make([]*resp3.Value, len(members))
		for i, member := range members {
			results[i] = resp3.NewSimpleStringValue(member.Value)
		}
		io.AddReplyArray(client, results)
	}
	return
}

// ZRangeByScore - 通过分数区间获取成员
func ZRangeByScore(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 3 {
		return errNotEnoughArgs
	}

	db := client.Db
	zsetKey := req[0].Str

	minScore, err := strconv.ParseFloat(req[1].Str, 64)
	if err != nil {
		return err
	}
	maxScore, err := strconv.ParseFloat(req[2].Str, 64)
	if err != nil {
		return err
	}

	var withscores bool
	if len(req) > 3 {
		w := req[3].Str
		withscores = (strings.ToUpper(w) == "WITHSCORES")
	}

	zsetObj := db.LookupKey(zsetKey)
	if zsetObj == nil {
		io.SendReplyToClient(client, shared.Shared.Nil)
		return
	}

	zs := zsetObj.Ptr.(*ZSet)
	members := zs.ZSetSearchRange(minScore, maxScore)
	fmt.Println(members)
	// for _, m := range members {
	// 	io.AddReplyString(client, m.Value)
	// 	io.AddReplyDouble(client, m.Score)
	// }

	if withscores {
		results := make([]*resp3.Value, 2*len(members))
		for i, member := range members {
			results[2*i] = resp3.NewSimpleStringValue(member.Value)
			score, _ := zs.ZSetGetScore(member.Value)
			results[2*i+1] = resp3.NewDoubleValue(score)
		}
		io.AddReplyArray(client, results)
	} else {
		results := make([]*resp3.Value, len(members))
		for i, member := range members {
			results[i] = resp3.NewSimpleStringValue(member.Value)
		}
		io.AddReplyArray(client, results)
	}
	return
}

// ZRem - Remove one or more members from a sorted set
func ZRem(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 2 {
		return errNotEnoughArgs
	}

	db := client.Db
	zsetKey := req[0].Str
	members := req[1:]

	zsetObj := db.LookupKey(zsetKey)
	if zsetObj == nil {
		io.AddReplyString(client, zsetKey+"not found")
		return errZSetNotFound
	} else if zsetObj.Type != core.RedisZSet {
		return errNotAZSet
	}

	zs := zsetObj.Ptr.(*zset.ZSet)
	removedCount := 0

	for _, member := range members {
		status := zs.ZSetRemoveValue(member.Str)
		if status == zset.ZSetOk {
			removedCount++
		} else {
			io.AddReplyString(client, member.Str+"not found")
		}
	}

	io.AddReplyNumber(client, int64(removedCount))
	return
}

// ZScore - 获取一个成员的分数
func ZScore(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 2 {
		return errNotEnoughArgs
	}

	db := client.Db
	zsetKey := req[0].Str
	member := req[1].Str

	zsetObj := db.LookupKey(zsetKey)
	if zsetObj == nil {
		io.AddReplyString(client, zsetKey+"not found")
		return errZSetNotFound
	} else if zsetObj.Type != core.RedisZSet {
		return errNotAZSet
	}

	zs := zsetObj.Ptr.(*zset.ZSet)
	score, exist := zs.ZSetGetScore(member)
	if exist {
		io.AddReplyDouble(client, score)
	} else {
		io.SendReplyToClient(client, shared.Shared.Nil)
	}

	return
}
