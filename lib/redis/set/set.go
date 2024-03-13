package set

import (
	"github.com/cinea4678/resp3"
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/io"
)

/**
集合基本操作命令の实现
*/

// SAdd SADD命令 向集合添加一个或多个成员，如果成员已存在则忽略。
// https://redis.io/commands/sadd/
func SAdd(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 2 {
		return errNotEnoughArgs
	}

	db := client.Db
	key := req[0].Str
	values := req[1:]

	var set *core.Set

	if setObj := db.LookupKey(key); setObj == nil {
		set = &core.Set{}
		db.DbAdd(key, core.CreateSet(set))
	} else {
		if setObj.Type != core.RedisSet {
			return errNotASet
		}
		set = setObj.Ptr.(*core.Set)
	}

	var countNew int64 = 0
	for _, value := range values {
		str := value.Str
		obj := core.CreateString(str)
		repeat, err := set.Add(obj)
		if err != nil {
			return err
		}
		if !repeat {
			countNew++
		}
	}

	io.AddReplyNumber(client, countNew)
	return
}

// SIsMember SISMEMBER命令 判断成员元素是否是集合的成员
// https://redis.io/commands/sismember/
func SIsMember(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 2 {
		return errNotEnoughArgs
	}

	db := client.Db
	key := req[0].Str
	value := req[1].Str

	if setObj := db.LookupKey(key); setObj == nil {
		io.AddReplyNumber(client, 0)
	} else {
		if setObj.Type != core.RedisSet {
			return errNotASet
		}
		set := setObj.Ptr.(*core.Set)

		if set.Find(core.CreateString(value)) {
			io.AddReplyNumber(client, 1)
		} else {
			io.AddReplyNumber(client, 0)
		}
	}
	return
}

// SMembers 返回集合中的所有成员
// https://redis.io/commands/smembers/
func SMembers(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 1 {
		return errNotEnoughArgs
	}

	db := client.Db
	key := req[0].Str
	if setObj := db.LookupKey(key); setObj == nil {
		io.AddReplyArray(client, []*resp3.Value{})
	} else {
		if setObj.Type != core.RedisSet {
			return errNotASet
		}
		set := setObj.Ptr.(*core.Set)

		size := set.Size()
		res := make([]*resp3.Value, 0, size)
		set.ForEach(func(obj *core.Object) {
			str, _ := obj.GetString()
			res = append(res, resp3.NewSimpleStringValue(str))
		})
		io.AddReplyArray(client, res)
	}
	return
}

// SCard 获取集合中元素的数量
// https://redis.io/commands/scard/
func SCard(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 1 {
		return errNotEnoughArgs
	}

	db := client.Db
	key := req[0].Str
	if setObj := db.LookupKey(key); setObj == nil {
		io.AddReplyNumber(client, 0)
	} else {
		if setObj.Type != core.RedisSet {
			return errNotASet
		}
		set := setObj.Ptr.(*core.Set)

		size := set.Size()
		io.AddReplyNumber(client, int64(size))
	}
	return
}

// SRem 移除集合中一个或多个成员
// https://redis.io/commands/srem/
func SRem(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 2 {
		return errNotEnoughArgs
	}

	db := client.Db
	key := req[0].Str
	values := req[1:]

	if setObj := db.LookupKey(key); setObj == nil {
		io.AddReplyNumber(client, 0)
	} else {
		if setObj.Type != core.RedisSet {
			return errNotASet
		}
		set := setObj.Ptr.(*core.Set)

		var count int64 = 0
		for _, value := range values {
			str := value.Str
			obj := core.CreateString(str)
			ok, err := set.Remove(obj)
			if err != nil {
				return err
			}
			if ok {
				count++
			}
		}

		io.AddReplyNumber(client, count)
		return
	}
	return
}
