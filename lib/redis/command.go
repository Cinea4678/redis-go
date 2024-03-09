package redis

import (
	"errors"
	"log"
)

var redisCommandTable = []*redisCommand{
	{name: "ping", redisClientFunc: cmdPing},
}

var (
	errCommandUnknown = errors.New("command unknown")
)

// redis命令结构
type redisCommand struct {
	name            string                          // 命令名称
	redisClientFunc func(client *redisClient) error // 命令处理函数
}

// 处理命令
func processCommand(client *redisClient) (err error) {
	cmd := client.reqValue.Elems[0].Str
	if cmd == "quit" {
		client.flags |= redisCloseAfterReply
		sendReplyToClient(client, shared.ok)
		return nil
	}

	client.cmd = lookupCommand(cmd)
	client.lastCmd = client.cmd

	if client.cmd == nil {
		return errCommandUnknown
	}

	return call(client, 0)
}

func lookupCommand(name string) *redisCommand {
	cmd := server.commands.DictFind(name)
	log.Printf("lookup command: %v", cmd)
	if cmd == nil {
		return nil
	}
	return cmd.(*redisCommand)
}

func call(client *redisClient, _ int) error {
	log.Printf("call command: %v", client.argv)
	return client.cmd.redisClientFunc(client)
}
