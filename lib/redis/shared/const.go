package shared

import "time"

// TODO! 可修改为配置文件设定

const (
	//redisLruClockResolution = 1000
	//redisLruBits            = 24                  //LRU时钟位数
	//redisLruClockMax        = 1<<redisLruBits - 1 //LRU时钟最大值

	RedisServerPort = 6389
	RedisTcpBacklog = 511

	AOFInterval = 1 * time.Second // aof间隔时间
	AOFBuffer   = 1000            //aof缓冲区刷新大小
)
