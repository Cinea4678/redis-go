package io

import (
	"errors"
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/shared"
	"runtime"
	"strconv"
	"time"

	"github.com/cinea4678/resp3"
	"github.com/panjf2000/gnet/v2"
	"github.com/rs/zerolog/log"
)

var (
	errUnknownMessage = errors.New("unknown message")
)

const (
	RedisCloseAfterReply = 1 << 6 // 回复后关闭客户端
)

func generateClientId() int {
	shared.Server.ClientCounter++
	return shared.Server.ClientCounter
}

// 创建客户端对象，用来处理命令和回复命令
func createClient(c gnet.Conn) *core.RedisClient {
	c.SetContext(generateClientId())
	return &core.RedisClient{
		Id:   c.Context().(int),
		Conn: c,
		Db:   shared.Server.Db[0],
	}
}

// 释放某个客户端对象
func freeClient(client *core.RedisClient) {
	log.Info().Str("addr", client.Conn.RemoteAddr().String()).Msg("disconnected")
	id := strconv.FormatInt(int64(client.Id), 10)
	shared.Server.Clients.DictRemove(id)
	err := client.Conn.Close()
	if err != nil {
		log.Error().Msgf("close client err: %v", err)
	}
	client = nil
}

// 清理client数据，准备处理下一个命令
func resetClient(client *core.RedisClient) {
	client.ReqValue = nil
}

// AcceptHandler 接收到新的请求，创建客户端，用来处理命令和回复命令
func AcceptHandler(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Info().Str("addr", c.RemoteAddr().String()).Msg("connected")
	client := createClient(c)
	id := strconv.FormatInt(int64(client.Id), 10)
	shared.Server.Clients.DictAdd(id, client)
	return out, action
}

// DataHandler 接收到客户端的命令
func DataHandler(c gnet.Conn) (action gnet.Action) {
	id := strconv.FormatInt(int64(c.Context().(int)), 10)
	client := shared.Server.Clients.DictFind(id).(*core.RedisClient)

	// 将数据设置到client中
	frame, _ := c.Next(-1)
	client.RawReq = string(frame)
	value, err := resp3.FromString(client.RawReq)
	if err != nil {
		AddReplyError(client, err)
		return gnet.Close
	}
	client.ReqValue = value

	defer func() {
		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			log.Error().Msgf("%v => %s", err, string(buf[:n]))
		}
	}()

	// 处理数据
	err = processInputBuffer(client)
	if err != nil {
		AddReplyError(client, err)
	}

	return action
}

// 处理客户端收到的数据
func processInputBuffer(client *core.RedisClient) error {
	req := client.ReqValue

	if req.Type != resp3.TypeArray {
		// 未知消息类型
		return errUnknownMessage
	} else {
		// 计时 执行
		startTime := time.Now()
		err := ProcessCommand(client)
		duration := time.Since(startTime).Seconds()

		// 克隆Profile
		client.LastProfile = &core.Profile{
			MemStat:  &runtime.MemStats{},
			TimeCost: duration,
		}
		runtime.ReadMemStats(client.LastProfile.MemStat)
		return err
	}
}
