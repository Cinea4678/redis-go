package system

import (
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/io"
)

// Info 获取服务器的数据
// https://redis.io/commands/info/ 这里没有完整实现所有条目
func Info(client *core.RedisClient) (err error) {
	io.AddReplyString(client, `# Server
redis_version:6.2.14
redis_mode:standalone
`)
	return nil
}
