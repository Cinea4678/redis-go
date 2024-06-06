package core

import (
	"github.com/cinea4678/resp3"
	"github.com/panjf2000/gnet/v2"
)

type RedisClient struct {
	Id   int
	Conn gnet.Conn // 客户端连接
	Db   *RedisDb  // Db
	Name *Object

	Cmd     *RedisCommand // 当前执行的命令
	LastCmd *RedisCommand // 最后执行的命令

	RawReq   string       // 从客户端收到的原始请求
	ReqValue *resp3.Value // 从客户端收到的请求

	Flags int  //处理标记
	IsAOF bool //是否为AOF虚拟客户端
}
