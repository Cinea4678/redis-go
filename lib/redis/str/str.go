package str

import (
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/io"
	"redis-go/lib/redis/shared"
	"strconv"
	"strings"

	"github.com/cinea4678/resp3"
)

/**
String 基本操作实现
*/

const (
	objNoFlags = 0
	objSetNX   = 1 << (iota - 1) // Set if key not exists
	objSetXX                     // Set if key exists
	objEX                        // Set if time in seconds is given
	objPX                        // Set if time in ms in given
	objKeepTtl                   // Set and keep the ttl
	objSetGet                    // Set if want to get key before set
	objEXAT                      // Set if timestamp in second is given
	objPXAT                      // Set if timestamp in ms is given
	objPersist                   // Set if we need to remove the ttl
)

// Set 通用的Set命令处理函数
func Set(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]
	flags, expire, ts, err := parseSetArgs(req)
	if err != nil {
		return err
	}

	db := client.Db
	key := req[0].Str

	// 如果指定需要GET，那就先执行GET返回旧值
	if flags&objSetGet > 0 {
		err = doGet(client, key)
		if err != nil {
			return
		}
	}

	found := db.LookupKey(key) != nil

	// 不满足NX或者XX的条件
	if (flags&objSetNX > 0 && found) || (flags&objSetXX > 0 && !found) {
		if !(flags&objSetGet > 0) {
			io.SendReplyToClient(client, shared.Shared.Nil)
		}
		return
	}

	db.SetKey(key, core.CreateString(req[1].Str))

	if expire {
		expireTime := parseExpireTime(flags, ts)
		db.SetExpire(key, expireTime)
	}

	if !(flags&objSetGet > 0) {
		io.SendReplyToClient(client, shared.Shared.Ok)
	}

	return
}

func parseSetArgs(req []*resp3.Value) (flags int, expire bool, timestamp int64, err error) {
	if len(req) < 2 {
		err = errNotEnoughArgs
		return
	}
	req = req[2:]

	countNxXx := 0
	countExpire := 0

	for len(req) > 0 {
		arg := strings.ToUpper(req[0].Str)

		switch arg {
		case "NX":
			flags |= objSetNX
			countNxXx++
		case "XX":
			flags |= objSetXX
			countNxXx++
		case "GET":
			flags |= objSetGet
		case "EX":
			flags |= objEX
			countExpire++
		case "PX":
			flags |= objPX
			countExpire++
		case "EXAT":
			flags |= objEXAT
			countExpire++
		case "PXAT":
			flags |= objPXAT
			countExpire++
		case "KEEPTTL":
			flags |= objKeepTtl
		}

		req = req[1:]

		if countNxXx > 1 {
			err = errNXAndXXConflict
			return
		}

		if countExpire > 1 {
			err = errExpiresConflict
			return
		}

		if arg == "EX" || arg == "PX" || arg == "EXAT" || arg == "PXAT" {
			if len(req) == 0 {
				err = errNotEnoughArgsExpire
				return
			}
			ts := req[0].Str
			timestamp, err = strconv.ParseInt(ts, 10, 64)
			if err != nil {
				return
			}
			req = req[1:]
		}
	}
	expire = countExpire > 0

	return
}

func parseExpireTime(flags int, expire int64) int64 {
	if flags&objEX > 0 {
		return core.GetTimeUnixMilli() + expire*1000
	} else if flags&objPX > 0 {
		return core.GetTimeUnixMilli() + expire
	} else if flags&objEXAT > 0 {
		return expire * 1000
	} else if flags&objPXAT > 0 {
		return expire
	} else {
		return 0
	}
}

// Get GET命令
func Get(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 1 {
		return errNoKey
	}

	key := req[0].Str
	return doGet(client, key)
}

func doGet(client *core.RedisClient, key string) (err error) {
	value := client.Db.LookupKey(key)

	if value == nil {
		io.SendReplyToClient(client, shared.Shared.Nil)
	} else {
		io.AddReplyObject(client, value)
	}

	return
}

/*
Append：如果key已存在且为字符串，则追加到末尾；如果不存在，则设置为空字符串
Syntax: APPEND key value
Reply: append完成后value的长度
*/
func Append(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 2 {
		return errNotEnoughArgs
	}

	key := req[0].Str
	appendValue := req[1].Str

	db := client.Db

	// 查找键对应的值
	obj := db.LookupKey(key)

	// 如果键不存在，直接将值设置为给定的字符串，并返回新字符串的长度
	if obj == nil {
		newObj := core.CreateString(appendValue)
		db.SetKey(key, newObj)
		io.SendReplyToClient(client, resp3.NewNumberValue(int64(len(appendValue))))
		return
	}

	// 追加字符串
	str, err := obj.GetString()
	newValue := str + appendValue
	newObj := core.CreateString(newValue)
	db.SetKey(key, newObj)

	// 返回追加后的字符串长度
	io.SendReplyToClient(client, resp3.NewNumberValue(int64(len(newValue))))
	return
}

//func setGenericCommand(client *RedisClient, flags int, key *Object, val *Object, expire *Object, unit int, ok_reply *Object, abort_reply *Object) {
//	milliseconds := 0
//	found := 0
//	setkey_falgs := 0
//
//	//if expire != nil &&
//}
//
//func getExpireMillisecondsOrReply(client *RedisClient, expire *Object, flags int, unit int) (milliseconds uint64, err error) {
//	milliseconds, err = expire.getUint64()
//	if err != nil {
//		return
//	}
//
//	if unit == unitSeconds && milliseconds > math.MaxUint64/1000 {
//		addReplyErrorExpireTime(client)
//		return 0,
//	}
//
//}
