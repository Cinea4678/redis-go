package list

import (
	"github.com/cinea4678/resp3"
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/io"
	"strconv"
)

/**
list基本操作命令实现
*/

// lpush lpush命令 向list头部添加成员。
// https://redis.io/docs/latest/commands/lpush/
func LPush(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]
	if len(req) < 2 {
		return errNotEnoughArgs
	}

	db := client.Db
	key := req[0].Str
	values := req[1:]

	var list *core.List

	if listObj := db.LookupKey(key); listObj == nil {
		list = core.NewList()
		db.DbAdd(key, core.CreateList(list))
	} else {
		if listObj.Type != core.RedisList {
			return errNotAList
		}
		list = listObj.Ptr.(*core.List)
	}

	var pos = 0
	var countNew int64 = 0
	for _, value := range values {
		str := value.Str
		if num, err := strconv.Atoi(str); err == nil {
			//fmt.Println(num)
			list.InsertInteger(pos, int64(num))
			countNew++
		} else {
			//fmt.Println(str)
			list.InsertBytes(pos, []byte(str))
			countNew++
		}
		if err != nil {
			return err
		}
	}
	io.AddReplyNumber(client, countNew)
	return
}

// rpush rpush命令 向list尾部添加成员。
// https://redis.io/docs/latest/commands/lpush/
func RPush(client *core.RedisClient) (err error) {
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
			return errNotAList
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

// lpop lpop命令 移除第一个元素
// https://redis.io/docs/latest/commands/lpop/
func LPop(client *core.RedisClient) (err error) {
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
			return errNotAList
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

// rpop rpop命令 移除最后一个元素
// https://redis.io/docs/latest/commands/rpop/
func RPop(client *core.RedisClient) (err error) {
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
			return errNotAList
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

// lindex 获取列表中指定位置的元素
// TODO 使用封装好的Index方法
// https://redis.io/commands/commands/lindex/
func LIndex(client *core.RedisClient) (err error) {
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
			return errNotAList
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

// lrange 获取列表中指定范围的元素序列
// TODO 用Go的Next方法
// https://redis.io/commands/commands/lrange/
func LRange(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]
	if len(req) < 2 {
		return errNotEnoughArgs
	}

	db := client.Db
	key := req[0].Str
	if setObj := db.LookupKey(key); setObj == nil {
		io.AddReplyNumber(client, 0)
	} else {
		if setObj.Type != core.RedisSet {
			return errNotAList
		}
		set := setObj.Ptr.(*core.Set)

		size := set.Size()
		io.AddReplyNumber(client, int64(size))
	}
	return
}
