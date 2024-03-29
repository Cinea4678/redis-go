package redis

import (
	"github.com/panjf2000/gnet/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/io"
	"redis-go/lib/redis/set"
	"redis-go/lib/redis/shared"
	"redis-go/lib/redis/str"
	"redis-go/lib/redis/system"
	"strconv"
)

const (
// unitSeconds = iota
// unitMilliseconds
)

// 初始化server配置
func initServerConfig() {
	shared.Server.Port = shared.RedisServerPort
	shared.Server.TcpBacklog = shared.RedisTcpBacklog
	shared.Server.Events = &core.EventLoop{}

	io.RedisCommandTable = append(io.RedisCommandTable, system.CommandTable...)
	io.RedisCommandTable = append(io.RedisCommandTable, str.StringsCommandTable...)
	io.RedisCommandTable = append(io.RedisCommandTable, set.SetCommandTable...)
}

// 初始化server
func initServer() {
	shared.Server.Pid = os.Getpid()

	shared.Server.Clients = core.NewDict()

	// 初始化事件处理器
	shared.Server.Events.Traffic = io.DataHandler
	shared.Server.Events.Open = io.AcceptHandler
	//server.events.tick = func() (delay time.Duration, action gnet.Action) {
	//	return serverCron(), action
	//}

	shared.Server.Commands = initCommandDict()

	shared.Server.Db = &core.RedisDb{
		Dict:    core.NewDict(),
		Expires: core.NewDict(),
		Id:      1,
	}

	shared.CreateSharedValues()
}

func initCommandDict() *core.Dict {
	d := core.NewDict()
	for _, cmd := range io.RedisCommandTable {
		d.DictAdd(cmd.Name, cmd)
	}
	return d
}

//func lruClock() uint64 {
//	if 1000/Server.hz <= redisLruClockResolution {
//		return Server.lruclock
//	}
//	return getLruClock()
//}

//func getLruClock() uint64 {
//	return uint64(msTime() / redisLruClockResolution & redisLruClockMax)
//}

//func msTime() int64 {
//	return time.Now().UnixNano() / 1000 / 1000
//}

//func ustime() int64 {
//	return time.Now().UnixNano() / 1000
//}
//
//func serverCron() time.Duration {
//	server.lruclock = getLruClock()
//	databasesCron()
//	return time.Millisecond * time.Duration(1000/server.hz)
//}

// Start 启动服务器
func Start() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	initServerConfig()
	initServer()
	elMain()
}

func elMain() {
	addr := "tcp://" + shared.Server.BindAddr + ":" + strconv.Itoa(shared.Server.Port)
	log.Info().Str("addr", addr).Msg("server is now listening")
	log.Fatal().Err(gnet.Run(shared.Server.Events, addr,
		gnet.WithMulticore(false), gnet.WithNumEventLoop(1)))
}
