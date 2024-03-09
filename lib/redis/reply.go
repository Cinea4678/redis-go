package redis

import (
	"github.com/panjf2000/gnet/v2"
	"github.com/smallnest/resp3"
	"log"
	"math/big"
)

func addReplyNumber(client *redisClient, number int64) {
	v := resp3.Value{Type: resp3.TypeNumber, Integer: number}
	sendReplyToClient(client, &v)
}

func addReplyBigNumber(client *redisClient, number *big.Int) {
	v := resp3.Value{Type: resp3.TypeNumber, BigInt: number}
	sendReplyToClient(client, &v)
}

func addReplyError(client *redisClient, err error) {
	s := "ERR " + err.Error()

	var resType byte
	if len(s) > 32 {
		resType = resp3.TypeBlobError
	} else {
		resType = resp3.TypeSimpleError
	}
	v := resp3.Value{Type: resType, Err: s}
	sendReplyToClient(client, &v)
}

func addReplyString(client *redisClient, str string) {
	var resType byte
	if len(str) > 32 {
		resType = resp3.TypeBlobString
	} else {
		resType = resp3.TypeSimpleString
	}
	v := resp3.Value{Type: resType, Str: str}
	sendReplyToClient(client, &v)
}

func sendReplyToClient(client *redisClient, value *resp3.Value) {
	s := value.ToRESP3String()
	sendRawReplyToClient(client, []byte(s))
}

func sendRawReplyToClient(client *redisClient, bytes []byte) {
	err := client.conn.AsyncWrite(bytes, func(c gnet.Conn, err error) error { return err })
	if err != nil {
		log.Printf("err: %v", err)
	}
	if client.flags&redisCloseAfterReply == 1 {
		freeClient(client)
	}
}
