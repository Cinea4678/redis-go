package list

import (
	"errors"
	"github.com/cinea4678/resp3"
	"redis-go/lib/redis/core"
	ziplist "redis-go/lib/redis/core/zip_list"
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
	var countNew int64 = 0
	for _, value := range values {
		str := value.Str
		if num, err := strconv.Atoi(str); err == nil {
			list.PushInteger(int64(num))
			countNew++
		} else {
			list.PushBytes([]byte(str))
			countNew++
		}
		if err != nil {
			return err
		}
	}
	io.AddReplyNumber(client, countNew)
	return
}

// lpop lpop命令 移除第一个元素
// https://redis.io/docs/latest/commands/lpop/
func LPop(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 1 {
		return errNotEnoughArgs
	}
	if len(req) > 2 {
		return errTooManyArgs
	}

	db := client.Db
	key := req[0].Str

	var list *core.List

	if listObj := db.LookupKey(key); listObj == nil {
		return errors.New("no such key")
	} else {
		if listObj.Type != core.RedisList {
			return errNotAList
		}
		list = listObj.Ptr.(*ziplist.Ziplist)
	}
	result := make([]*resp3.Value, 0)

	rmIndex := 1
	count := 1
	if len(req) == 2 {
		count, err = strconv.Atoi(req[1].Str)
		if err != nil {
			return errInvalidIndex
		}
		if count < 0 {
			return errIndexOutOfRange
		}
	}
	for i := 0; i < count; i++ {
		node := list.Index(rmIndex)
		if node == nil {
			return
		}
		if node.IsInteger() {
			result = append(result, resp3.NewSimpleStringValue(strconv.Itoa(int(node.GetInteger()))))
		} else {
			result = append(result, resp3.NewSimpleStringValue(string(node.GetByteArray())))
		}
		list.DeleteByPos(rmIndex)
	}
	io.AddReplyArray(client, result)
	return
}

// rpop rpop命令 移除最后一个元素
// https://redis.io/docs/latest/commands/rpop/
func RPop(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 1 {
		return errNotEnoughArgs
	}
	if len(req) > 2 {
		return errTooManyArgs
	}

	db := client.Db
	key := req[0].Str

	var list *core.List

	if listObj := db.LookupKey(key); listObj == nil {
		return errors.New("no such key")
	} else {
		if listObj.Type != core.RedisList {
			return errNotAList
		}
		list = listObj.Ptr.(*ziplist.Ziplist)
	}
	result := make([]*resp3.Value, 0)

	rmIndex := list.Len()
	count := 1
	if len(req) == 2 {
		count, err = strconv.Atoi(req[1].Str)
		if err != nil {
			return errInvalidIndex
		}
		if count < 0 {
			return errIndexOutOfRange
		}
	}
	for i := 0; i < count; i++ {
		rmIndex = list.Len()
		node := list.Index(rmIndex)
		if node == nil {
			return
		}
		if node.IsInteger() {
			result = append(result, resp3.NewSimpleStringValue(strconv.Itoa(int(node.GetInteger()))))
		} else {
			result = append(result, resp3.NewSimpleStringValue(string(node.GetByteArray())))
		}
		list.DeleteByPos(rmIndex)
	}
	io.AddReplyArray(client, result)
	return
}

// lindex 获取列表中指定位置的元素
// TODO 使用封装好的Index方法
// https://redis.io/commands/commands/lindex/
func LIndex(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]
	if len(req) < 2 {
		return errNotEnoughArgs
	}

	if len(req) > 2 {
		return errTooManyArgs
	}

	db := client.Db
	key := req[0].Str
	idx, err := strconv.Atoi(req[1].Str)
	if err != nil {
		return errInvalidIndex
	}

	var list *ziplist.Ziplist

	if listObj := db.LookupKey(key); listObj == nil {
		return errors.New("no such key")
	} else {
		if listObj.Type != core.RedisList {
			return errNotAList
		}
		list = listObj.Ptr.(*ziplist.Ziplist)
	}

	// Fetch the range of elements from the list
	result := ""

	if idx >= 0 {
		idx++
	} else {
		idx = list.Len() + 1 + idx
	}

	if idx > list.Len() || idx < 1 {
		return errIndexOutOfRange
	}

	node := list.Index(idx)
	if node == nil {
		return
	}
	if node.IsInteger() {
		result = strconv.Itoa(int(node.GetInteger()))
	} else {
		result = string(node.GetByteArray())
	}

	// Add the range result to the client's response
	io.AddReplyString(client, result)

	return
}

// lrange 获取列表中指定范围的元素序列
// TODO 用Go的Next方法
// https://redis.io/commands/commands/lrange/
func LRange(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]
	if len(req) < 3 {
		return errNotEnoughArgs
	}

	if len(req) > 3 {
		return errTooManyArgs
	}

	db := client.Db
	key := req[0].Str
	start, err := strconv.Atoi(req[1].Str)
	if err != nil {
		return errInvalidIndex
	}
	stop, err := strconv.Atoi(req[2].Str)
	if err != nil {
		return errInvalidIndex
	}

	var list *ziplist.Ziplist

	if listObj := db.LookupKey(key); listObj == nil {
		return errors.New("no such key")
	} else {
		if listObj.Type != core.RedisList {
			return errNotAList
		}
		list = listObj.Ptr.(*ziplist.Ziplist)
	}

	if start < 0 || start > list.Len() || (start > stop && stop > 0) {
		return errIndexOutOfRange
	}

	if stop > list.Len()-1 {
		stop = list.Len() - 1
	} else if stop < 0 {
		stop = stop + list.Len()
		if stop < 0 {
			return errIndexOutOfRange
		}
	}
	// Fetch the range of elements from the list
	result := make([]*resp3.Value, 0)

	for i := start + 1; i <= stop+1; i++ {
		node := list.Index(i)
		if node == nil {
			break
		}
		if node.IsInteger() {
			result = append(result, resp3.NewSimpleStringValue(strconv.Itoa(int(node.GetInteger()))))
		} else {
			result = append(result, resp3.NewSimpleStringValue(string(node.GetByteArray())))
		}
	}

	// Add the range result to the client's response
	io.AddReplyArray(client, result)

	return
}
