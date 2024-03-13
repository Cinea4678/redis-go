package io

import (
	"errors"
	"github.com/rs/zerolog/log"
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/shared"
	"strings"
)

// RedisCommandTable 总命令表
var RedisCommandTable []*core.RedisCommand

var (
	errCommandUnknown = errors.New("command unknown")
)

// ProcessCommand 处理命令
func ProcessCommand(client *core.RedisClient) (err error) {
	cmd := client.ReqValue.Elems[0].Str
	log.Info().Str("addr", client.Conn.RemoteAddr().String()).Str("command", cmd)

	if cmd == "quit" {
		client.Flags |= RedisCloseAfterReply
		SendReplyToClient(client, shared.Shared.Ok)
		return nil
	}

	client.Cmd = lookupCommand(cmd)
	client.LastCmd = client.Cmd

	if client.Cmd == nil {
		log.Warn().Str("addr", client.Conn.RemoteAddr().String()).Str("command", cmd).Msg("not found")
		return errCommandUnknown
	}

	return call(client, 0)
}

func lookupCommand(name string) *core.RedisCommand {
	name = strings.ToLower(name) // 转换为小写
	cmd := shared.Server.Commands.DictFind(name)
	if cmd == nil {
		return nil
	}
	return cmd.(*core.RedisCommand)
}

func call(client *core.RedisClient, _ int) error {
	return client.Cmd.RedisClientFunc(client)
}
