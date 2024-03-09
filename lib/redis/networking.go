package redis

import (
	"errors"
	"github.com/panjf2000/gnet/v2"
	"github.com/smallnest/resp3"
	"log"
	"runtime"
)

var (
	errUnknownMessage = errors.New("unknown message")
)

// 创建客户端对象，用来处理命令和回复命令
func createClient(c gnet.Conn) *redisClient {
	c.SetContext(generateClientId())
	return &redisClient{
		id:   c.Context().(int),
		conn: c,
		db:   server.db,
		argc: 0,
	}
}

// 释放某个客户端对象
func freeClient(client *redisClient) {
	server.clients.DictRemove(client.id)
	err := client.conn.Close()
	if err != nil {
		log.Printf("close client err: %v", err)
	}
	client = nil
}

// 清理client数据，准备处理下一个命令
func resetClient(client *redisClient) {
	client.argc = 0
	client.argv = nil
}

// 接收到新的请求，创建客户端，用来处理命令和回复命令
func acceptHandler(c gnet.Conn) (out []byte, action gnet.Action) {
	client := createClient(c)
	server.clients.DictAdd(client.id, client)
	return out, action
}

// 接收到客户端的命令
func dataHandler(c gnet.Conn) (action gnet.Action) {
	client := server.clients.DictFind(c.Context()).(*redisClient)

	// 将数据设置到client中
	frame, _ := c.Next(-1)
	value, err := resp3.FromString(string(frame))
	if err != nil {
		addReplyError(client, err)
		return gnet.Close
	}
	client.reqValue = value

	defer func() {
		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			log.Print(err)
			log.Printf("==> %s \n", string(buf[:n]))
		}
	}()

	// 处理数据
	err = processInputBuffer(client)
	if err != nil {
		addReplyError(client, err)
	}

	return action
}

// 处理客户端收到的数据
func processInputBuffer(client *redisClient) error {
	req := client.reqValue

	if req.Type != resp3.TypeArray {
		// 未知消息类型
		return errUnknownMessage
	} else {
		return processCommand(client)
	}
}
