package str

import (
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/io"
	"redis-go/lib/redis/shared"
	"strconv"
	"strings"

	"github.com/cinea4678/resp3"
	"go.uber.org/multierr"
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
GetRange: 获取字符串值的子字符串(负数表示倒数第x个)
Syntax: GETRANGE key start end
Reply: 子字符串

https://www.redisdocs.com/commands/getrange/
*/
func GetRange(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 3 {
		return errNotEnoughArgs
	}

	key := req[0].Str
	start, err := strconv.ParseInt(req[1].Str, 10, 64)
	if err != nil {
		return errInvalidInt
	}

	end, err := strconv.ParseInt(req[2].Str, 10, 64)
	if err != nil {
		return errInvalidInt
	}

	db := client.Db

	// 查找键对应的值（其实官方实现是返回空字符串而非nil）
	obj := db.LookupKey(key)
	if obj == nil {
		io.SendReplyToClient(client, shared.Shared.Nil)
		return
	}

	str, err := obj.GetString()
	strLen := int64(len(str))

	// 索引超限
	if start >= strLen || end < -strLen {
		// log.Println("out of range")
		io.AddReplyString(client, "")
		return
	}
	// 特殊情况的索引处理，部分细节处理与官方不同

	if start < -strLen {
		// 超出倒数范围，则规定为头索引
		start = 0
	} else if start < 0 {
		// 没超限的负数，转换为倒数
		start = strLen + start
	}

	if end >= strLen {
		// 超出数组长度，则规定为尾索引
		end = strLen - 1
	} else if end < 0 {
		// 没超限的负数，转换为倒数
		end = strLen + end
	}

	// 转换后start大于end，返回空字符串
	if start > end {
		io.AddReplyString(client, "")
		return
	}

	// 获取子字符串并返回
	subrange := str[start : end+1]
	io.AddReplyString(client, subrange)
	return
}

/*
GetDel 获取指定键的值并删除键（但目前的实现要先查找再查找后删除）
Syntax: GETDEL key
Reply: 键对应值
*/
func GetDel(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 1 {
		return errNotEnoughArgs
	}

	key := req[0].Str

	db := client.Db

	// 查找并删除键对应的值
	obj := db.LookupKeyDel(key)
	if obj == nil {
		io.SendReplyToClient(client, shared.Shared.Nil)
		return
	}

	str, err := obj.GetString()

	// 返回键值
	io.AddReplyString(client, str)

	return
}

/*
Append：如果key已存在且为字符串，则追加到末尾；如果不存在，则设置为空字符串
Syntax: APPEND key value
Reply: append完成后value的长度

https://www.redisdocs.com/commands/append/
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
		io.AddReplyNumber(client, int64(len(appendValue)))
		return
	}

	// 追加字符串
	str, err := obj.GetString()
	newValue := str + appendValue
	newObj := core.CreateString(newValue)
	db.SetKey(key, newObj)

	// 返回追加后的字符串长度
	io.AddReplyNumber(client, int64(len(newValue)))
	return
}

func LCS(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 2 {
		return errNotEnoughArgs
	}

	key1 := req[0].Str
	key2 := req[1].Str

	db := client.Db

	obj1 := db.LookupKey(key1)
	obj2 := db.LookupKey(key2)

	if obj1 == nil || obj2 == nil {
		io.SendReplyToClient(client, shared.Shared.Nil)
		return
	}

	str1, err1 := obj1.GetString()
	str2, err2 := obj2.GetString()

	if err1 != nil && err2 != nil {
		err = multierr.Append(err1, err2)
		return
	}

	lcs := FindLCS(str1, str2)
	io.AddReplyString(client, lcs)
	return
}

func FindLCS(str1, str2 string) string {
	n, m := len(str1), len(str2)

	// 最长公共子字符串在f1中的结束位置
	end := 0
	// 最长公共子字符串的长度
	maxLen := 0

	// 二维数组，dp[i][j]是s1中以i结尾和s2中以j结尾的最长公共子字符串的长度
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}

	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if str1[i-1] == str2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
				if dp[i][j] > maxLen {
					maxLen = dp[i][j]
					end = i - 1 // 更新最长公共子字符串的结束位置
				}
			}
		}
	}

	// 根据maxLen和end提取最长公共子字符串
	if maxLen > 0 {
		return str1[end-maxLen+1 : end+1]
	}

	return "" // 如果最长公共子字符串长度为0，则返回空字符串
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
