package system

import (
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/io"
	"strconv"
	"time"

	"github.com/cinea4678/resp3"
	"github.com/emirpasic/gods/maps/linkedhashmap"
	jsoniter "github.com/json-iterator/go"
)

// Ping Ping命令
// https://redis.io/commands/ping/
func Ping(client *core.RedisClient) error {
	req := client.ReqValue
	if len(req.Elems) < 2 {
		io.AddReplyString(client, "PONG")
	} else {
		io.AddReplyArray(client, req.Elems[1:])
	}
	return nil
}

// Time Time命令
// https://redis.io/commands/time/
// 遵照标准，返回字符串而不是数字
func Time(client *core.RedisClient) error {
	now := time.Now()

	unixTimestamp := now.Unix()
	microseconds := now.Nanosecond() / 1000

	req := []*resp3.Value{
		{Type: resp3.TypeSimpleString, Str: strconv.FormatInt(unixTimestamp, 10)},
		{Type: resp3.TypeSimpleString, Str: strconv.FormatInt(int64(microseconds), 10)},
	}

	io.AddReplyArray(client, req)
	return nil
}

// Echo Echo命令
// https://redis.io/commands/echo/
func Echo(client *core.RedisClient) error {
	req := client.ReqValue
	if len(req.Elems) < 2 {
		return errEchoNoMessage
	} else {
		io.AddReplyString(client, req.Elems[1].Str)
	}
	return nil
}

// Hello HELLO命令
// https://redis.io/commands/hello/
func Hello(client *core.RedisClient) error {
	kv := linkedhashmap.New()
	kv.Put(resp3.NewSimpleStringValue("server"), resp3.NewSimpleStringValue("redis-go-tj-sse"))
	kv.Put(resp3.NewSimpleStringValue("version"), resp3.NewSimpleStringValue("6.2.14"))
	kv.Put(resp3.NewSimpleStringValue("proto"), resp3.NewNumberValue(int64(3)))
	kv.Put(resp3.NewSimpleStringValue("id"), resp3.NewNumberValue(int64(client.Id)))
	kv.Put(resp3.NewSimpleStringValue("mode"), resp3.NewSimpleStringValue("standalone"))
	kv.Put(resp3.NewSimpleStringValue("role"), resp3.NewSimpleStringValue("master"))
	kv.Put(resp3.NewSimpleStringValue("modules"), resp3.NewArrayValue([]*resp3.Value{}))

	io.AddReplyMap(client, kv)
	return nil
}

// Prof PROF命令
// 我自己编的
func Prof(client *core.RedisClient) error {
	if client.LastProfile != nil {
		j, err := jsoniter.MarshalToString(client.LastProfile)
		if err != nil {
			return err
		}
		io.AddReplyString(client, j)
	} else {
		io.AddReplyNull(client)
	}
	return nil
}
