package io

import (
	"github.com/cinea4678/resp3"
	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/panjf2000/gnet/v2"
	"log"
	"math/big"
	"redis-go/lib/redis/core"
)

// AddReplyArray 向客户端发回一组值
func AddReplyArray(client *core.RedisClient, elements []*resp3.Value) {
	v := resp3.Value{Type: resp3.TypeArray, Elems: elements}
	SendReplyToClient(client, &v)
}

func AddReplyMap(client *core.RedisClient, map_ *linkedhashmap.Map) {
	v := resp3.Value{Type: resp3.TypeMap, KV: map_}
	SendReplyToClient(client, &v)
}

// AddReplyNumber 向客户端发回一个数字
func AddReplyNumber(client *core.RedisClient, number int64) {
	v := resp3.Value{Type: resp3.TypeNumber, Integer: number}
	SendReplyToClient(client, &v)
}

// AddReplyBigNumber 向客户端发回一个大数
func AddReplyBigNumber(client *core.RedisClient, number *big.Int) {
	v := resp3.Value{Type: resp3.TypeNumber, BigInt: number}
	SendReplyToClient(client, &v)
}

// AddReplyError 向客户端发回一个错误
func AddReplyError(client *core.RedisClient, err error) {
	s := "ERR " + err.Error()
	v := resp3.Value{Type: resp3.TypeSimpleError, Err: s}
	SendReplyToClient(client, &v)
}

// AddReplyString 向客户端发回一个字符串
func AddReplyString(client *core.RedisClient, str string) {
	var resType byte
	if len(str) > 32 {
		resType = resp3.TypeBlobString
	} else {
		resType = resp3.TypeSimpleString
	}
	v := resp3.Value{Type: resType, Str: str}
	SendReplyToClient(client, &v)
}

// AddReplyObject 向客户端发回一个Redis对象
func AddReplyObject(client *core.RedisClient, obj *core.Robj) {
	switch obj.Rtype {
	case core.RedisString:
		AddReplyString(client, *obj.Ptr.(*string))
	}
}

// SendReplyToClient 向客户端发送一个已经构造为resp3.Value的值
func SendReplyToClient(client *core.RedisClient, value *resp3.Value) {
	s := value.ToRESP3String()
	sendRawReplyToClient(client, []byte(s))
}

// 向客户端发送原始的字节
func sendRawReplyToClient(client *core.RedisClient, bytes []byte) {
	err := client.Conn.AsyncWrite(bytes, func(c gnet.Conn, err error) error { return err })
	if err != nil {
		log.Printf("err: %v", err)
	}
	if client.Flags&RedisCloseAfterReply == 1 {
		freeClient(client)
	}
}
