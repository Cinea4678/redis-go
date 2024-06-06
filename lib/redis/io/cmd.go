package io

import (
	"errors"
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/resistence"
	"redis-go/lib/redis/shared"
	"strings"

	"github.com/cinea4678/resp3"
	"github.com/rs/zerolog/log"
)

// RedisCommandTable 总命令表
var RedisCommandTable []*core.RedisCommand

// RedisCommandInfo 总命令信息表
var RedisCommandInfo []*core.RedisCommandInfo

// redisCommandInfoPrepared 已转换成resp3.Value的命令信息表
var redisCommandInfoPrepared *resp3.Value

var (
	errCommandUnknown = errors.New("command unknown")
)

// ProcessCommand 处理命令
func ProcessCommand(client *core.RedisClient) (err error) {
	cmd := client.ReqValue.Elems[0].Str
	cmd = strings.ToLower(cmd) // 转换为小写
	log.Info().Str("addr", client.Conn.RemoteAddr().String()).Str("command", cmd).Msg("command received")

	if cmd == "quit" {
		client.Flags |= RedisCloseAfterReply
		SendReplyToClient(client, shared.Shared.Ok)
		return nil
	} else if cmd == "command" {
		if redisCommandInfoPrepared == nil {
			redisCommandInfoPrepared = core.RedisCommandInfoToValue(RedisCommandInfo)
		}
		SendReplyToClient(client, redisCommandInfoPrepared)
		return nil
	}

	client.Cmd = lookupCommand(cmd)
	client.LastCmd = client.Cmd

	if client.Cmd == nil {
		log.Warn().Str("addr", client.Conn.RemoteAddr().String()).Str("command", cmd).Msg("not found")
		return errCommandUnknown
	}

	// 检查是否需要持久化该命令
	if resistence.NeedAOF(cmd) {
		resistence.AddToAOFBuffer(client.ReqValue)
	}

	return call(client, 0)
}

func lookupCommand(name string) *core.RedisCommand {
	cmd := shared.Server.Commands.DictFind(name)
	if cmd == nil {
		return nil
	}
	return cmd.(*core.RedisCommand)
}

func call(client *core.RedisClient, _ int) error {
	return client.Cmd.RedisClientFunc(client)
}
