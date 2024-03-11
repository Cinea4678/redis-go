package str

import (
	"errors"
	"math"
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/io"
	"strconv"
)

var (
	errNotString = errors.New("target not a string")
	errOverflow  = errors.New("increment or decrement would overflow")
)

func Increase(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 1 {
		return errNoKey
	}

	return doIncrease(client, req[0].Str, 1)
}

func IncreaseBy(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 2 {
		return errNotEnoughArgs
	}

	increment, err := strconv.ParseInt(req[1].Str, 10, 64)

	return doIncrease(client, req[0].Str, increment)
}

func Decrease(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 1 {
		return errNoKey
	}

	return doIncrease(client, req[0].Str, -1)
}

func DecreaseBy(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]

	if len(req) < 2 {
		return errNotEnoughArgs
	}

	increment, err := strconv.ParseInt(req[1].Str, 10, 64)

	return doIncrease(client, req[0].Str, -increment)
}

func doIncrease(client *core.RedisClient, key string, increment int64) (err error) {
	var oldValue int64
	o := client.Db.LookupKey(key)
	if o != nil {
		if o.Rtype != core.RedisString {
			return errNotString
		}
		oldValue, _ = strconv.ParseInt(*o.Ptr.(*string), 10, 64)
	}

	// 检查是否会溢出
	if (increment < 0 && oldValue < 0 && increment < (math.MinInt64-oldValue)) || (increment > 0 && oldValue > 0 && increment > (math.MaxInt64-oldValue)) {
		return errOverflow
	}

	newValue := oldValue + increment
	newValueStr := strconv.FormatInt(newValue, 10)
	client.Db.SetKey(key, core.CreateString(&newValueStr))

	io.AddReplyNumber(client, newValue)
	return
}
