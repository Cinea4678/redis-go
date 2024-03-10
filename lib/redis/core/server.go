package core

type RedisServer struct {
	Pid int

	Hz       int // Hz
	Db       *RedisDb
	Commands *Dict //redis命令字典 string(如get/set) : *RedisCommand

	ClientCounter int   //存储client的id计数器
	Clients       *Dict //客户端字典 Id : *RedisClient

	Port       int
	TcpBacklog int
	BindAddr   string
	IpfdCount  int

	Events *EventLoop

	LruClock uint64
}
