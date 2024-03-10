package system

import (
	"errors"
	"github.com/cinea4678/resp3"
	"github.com/emirpasic/gods/maps/linkedhashmap"
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/io"
	"strconv"
	"time"
)

// UtilsCommand 系统工具类的命令
var UtilsCommand = []*core.RedisCommand{
	{"ping", Ping},
	{"time", Time},
	{"echo", Echo},
	{"hello", Hello},
}

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

var errEchoNoMessage = errors.New("no message given")

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
	kv.Put(resp3.NewSimpleStringValue("version"), resp3.NewSimpleStringValue("1.0.0"))
	kv.Put(resp3.NewSimpleStringValue("proto"), resp3.NewNumberValue(int64(3)))
	kv.Put(resp3.NewSimpleStringValue("id"), resp3.NewNumberValue(int64(client.Id)))
	kv.Put(resp3.NewSimpleStringValue("mode"), resp3.NewSimpleStringValue("standalone"))
	kv.Put(resp3.NewSimpleStringValue("role"), resp3.NewSimpleStringValue("master"))
	kv.Put(resp3.NewSimpleStringValue("modules"), resp3.NewArrayValue([]*resp3.Value{}))

	io.AddReplyMap(client, kv)
	return nil
}
