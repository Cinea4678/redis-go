package redis

import (
	"bytes"
	"log"
	"redis-go/lib/redis/ds"
	"runtime"
	"strconv"
	"strings"

	"github.com/panjf2000/gnet/v2"
)

const (
	redisReqInline    = 1	//内联命令请求：SET key value
	redisReqMultibulk = 2	//多条批量命令请求：*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
)

// 创建客户端对象，用来处理命令和回复命令
func createClient(c gnet.Conn) *redisClient {
	c.SetContext(generateClientId())
	return &redisClient{
		id:     c.Context().(int),
		conn:   c,
		db:     server.db,
		argc:   0,
		buf:    make([]byte, 1024*12),
		bufpos: 0,
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
	client.bufpos = 0
	client.reqtype = 0
	client.sentlen = 0
}

// 接收到新的请求，创建客户端，用来处理命令和回复命令
func acceptHandler(c gnet.Conn) (out byte[], action gnet.Action) {
	client := createClient(c)
	server.clients.DictAdd(client.id, client)
	return action
}

// 接收到客户端的命令
func dataHandler(c gnet.Conn) (action gnet.Action) {
	client := server.clients.DictFind(c.Context()).(*redisClient)

	// 将数据设置到client中
	frame, _ := c.Next(-1)
	client.queryBuf = string(frame)

	defer func() {
		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			log.Print(err)
			log.Printf("==> %s \n", string(buf[:n]))
		}
	}()

	// 处理数据
	processInputBuffer(client)

	return action
}

// 处理客户端收到的数据
func processInputBuffer(client *redisClient) {
	//判断命令类型
	if client.queryBuf[0] == '*' {
		client.reqtype = redisReqMultibulk
	} else {
		client.reqtype = redisReqInline
	}
	log.Printf("client reqtype is : %v", client.reqtype)

	//协议解析
	if client.reqtype == redisReqInline {
		if processInlineBuffer(client) == redisErr {
			//error
			log.Printf("analysis inline protocol error")
		}
	}
	if client.reqtype == redisReqMultibulk {
		if processMultibulkBuffer(client) == redisErr {
			//error
			log.Printf("analysis protocol error")
		}
	} else {
		panic("Unknown request type")
	}

	if client.argc == 0 {
		resetClient(client)
	} else {
		if processCommand(client) == redisErr {
			//error
			log.Printf("process command error")
		}
		resetClient(client)
		//server.currentClient = nil
	}
}

func addReplyLongLong(client *redisClient, ll int64) {
	if ll == 0 {
		addReply(client, shared.czero)
	} else if ll == 1 {
		addReply(client, shared.cone)
	} else {
		addReplyLongLongWithPrefix(client, ll, ":")
	}

}

func addReplyLongLongWithPrefix(client *redisClient, ll int64, prefix string) {
	output := prefix + strconv.FormatInt(ll, 10) + "\r\n"
	addReplyString(client, output)
}

func addReplyString(client *redisClient, str string) {
	addReplyToBuffer(client, string(str))
	sendReplyToClient(client)
}

func addReplyBulkLen(client *redisClient, obj *ds.Robj) {
	bulkLen := "$" + strconv.Itoa(len(obj.Ptr.(string))) + "\r\n"
	addReply(client, ds.CreateObject(ds.RedisString, string(bulkLen)))
}

func addReplyBulk(client *redisClient, obj *ds.Robj) {
	addReplyBulkLen(client, obj)
	addReply(client, obj)
	addReply(client, shared.crlf)
}

func addReply(client *redisClient, robj *ds.Robj) {
	log.Printf("add reply: %v", robj)
	//redis中使用reactor，所以这里理论上不是马上执行的, redis是先将 sendReplyToClient事件注册上去，然后再执行addReplyToBuffer
	addReplyToBuffer(client, robj.Ptr.(string))
	sendReplyToClient(client)
}

func addReplyToBuffer(client *redisClient, data string) {
	copy(client.buf[client.bufpos:], data)
	client.bufpos = client.bufpos + len([]byte(data))
}

func sendReplyToClient(client *redisClient) int {
	log.Printf("send reply to client: %v", client.buf[client.sentlen:client.bufpos])
	err := client.conn.AsyncWrite(client.buf[client.sentlen:client.bufpos], func(c gnet.Conn, err error) error { return err })
	if err != nil {
		log.Printf("err: %v", err)
	}
	client.sentlen = client.bufpos
	if client.flags&redisCloseAfterReply == 1 {
		freeClient(client)
	}
	return redisOk
}

func processInlineBuffer(client *redisClient) int {
	return redisErr
}

func processMultibulkBuffer(client *redisClient) int {
	newLines := strings.Split(string(client.queryBuf), "\n")

	argIdx := 0
	for i, line := range newLines {
		line = strings.Replace(line, "\r", "", 1)
		if i == 0 {
			//arg count
			var err error
			client.argc, err = strconv.Atoi(line[1:])
			if err != nil {
				return redisErr
			}
			client.argv = make([]*ds.Robj, client.argc)
			continue
		}
		if client.argc <= argIdx {
			break
		}
		if line[0] != '$' {
			client.argv[argIdx] = ds.CreateObject(ds.RedisString, string(line))
			argIdx++
		}
	}
	log.Printf("analysis command, command count: %v, value: %v", client.argc, client.argv)
	return redisOk
}
