package core

// RedisCommand redis命令结构
type RedisCommand struct {
	Name            string                          // 命令名称
	RedisClientFunc func(client *RedisClient) error // 命令处理函数
}
