package zset

import (
	"redis-go/lib/redis/core"
)

func ZAdd(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 2 {
		return errNotEnoughArgs
	}

	// key := req[0].Str
	// appendValue := req[1].Str

	// db := client.Db

	// // 查找键对应的值
	// obj := db.LookupKey(key)

	// // 如果键不存在，直接将值设置为给定的字符串，并返回新字符串的长度
	// if obj == nil {
	// 	newObj := core.CreateString(appendValue)
	// 	db.SetKey(key, newObj)
	// 	io.SendReplyToClient(client, resp3.NewNumberValue(int64(len(appendValue))))
	// 	return
	// }

	// // 追加字符串
	// str, err := obj.GetString()
	// newValue := str + appendValue
	// newObj := core.CreateString(newValue)
	// db.SetKey(key, newObj)

	// // 返回追加后的字符串长度
	// io.SendReplyToClient(client, resp3.NewNumberValue(int64(len(newValue))))
	return
}

// value应该是什么类型？
func doAdd(score float64, value string) {

}
