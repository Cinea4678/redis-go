package redis

import (
	"github.com/smallnest/resp3"
	"os"
	"time"

	"github.com/panjf2000/gnet/v2"
)

// TODO! 可修改为配置文件设定

// client flags
const (
	redisCloseAfterReply = 1 << 6
)

const (
	redisLruClockResolution = 1000
	redisLruBits            = 24                  //LRU时钟位数
	redisLruClockMax        = 1<<redisLruBits - 1 //LRU时钟最大值

	redisServerPort = 6389
	redisTcpBacklog = 511

	redisOk  = 0
	redisErr = -1
)

const (
	unitSeconds = iota
	unitMilliseconds
)

var (
	shared *sharedValuesStruct
)

type redisServer struct {
	pid int

	hz       int // hz
	db       *redisDb
	commands *dict //redis命令字典 string(如get/set) : *redisCommand

	clientCounter int   //存储client的id计数器
	clients       *dict //客户端字典 id : *redisClient

	port       int
	tcpBacklog int
	bindaddr   string
	ipfdCount  int

	events *eventloop

	lruclock uint64
}

type redisClient struct {
	id   int
	conn gnet.Conn // 客户端连接
	db   *redisDb  // Db
	name *robj
	argc int     // 命令数量
	argv []*robj // 命令值

	cmd     *redisCommand // 当前执行的命令
	lastCmd *redisCommand // 最后执行的命令

	reqValue *resp3.Value // 从客户端收到的请求

	flags int //处理标记
}

// 共享对象
type sharedValuesStruct struct {
	ok        *resp3.Value
	err       *resp3.Value
	syntaxErr *resp3.Value
	nullBulk  *resp3.Value
	cZero     *resp3.Value
	cOne      *resp3.Value
	oomErr    *resp3.Value
}

// 初始化server配置
func initServerConfig() {
	server.port = redisServerPort
	server.tcpBacklog = redisTcpBacklog
	server.events = &eventloop{}

}

// 初始化server
func initServer() {
	server.pid = os.Getpid()

	server.clients = &dict{}

	// 初始化事件处理器
	server.events.traffic = dataHandler
	server.events.open = acceptHandler
	//server.events.tick = func() (delay time.Duration, action gnet.Action) {
	//	return serverCron(), action
	//}

	server.commands = initCommandDict()

	server.db = &redisDb{
		dict:    &dict{},
		expires: &dict{},
		id:      1,
	}

	createSharedValues()
}

func initCommandDict() *dict {
	d := dict{}
	for _, cmd := range redisCommandTable {
		d.DictAdd(cmd.name, cmd)
	}
	return &d
}

func generateClientId() int {
	server.clientCounter++
	return server.clientCounter
}

func createSharedValues() {
	shared = &sharedValuesStruct{
		ok:        &resp3.Value{Type: resp3.TypeSimpleString, Str: "OK"},
		err:       &resp3.Value{Type: resp3.TypeSimpleError, Str: "ERR"},
		syntaxErr: &resp3.Value{Type: resp3.TypeSimpleError, Str: "-ERR syntax error"},
		cZero:     &resp3.Value{Type: resp3.TypeNumber, Integer: 0},
		cOne:      &resp3.Value{Type: resp3.TypeNumber, Integer: 1},
		oomErr:    &resp3.Value{Type: resp3.TypeSimpleError, Str: "-OOM command not allowed when used memory > 'maxmemory'"},
	}
}

func lruClock() uint64 {
	if 1000/server.hz <= redisLruClockResolution {
		return server.lruclock
	}
	return getLruClock()
}

func getLruClock() uint64 {
	return uint64(msTime() / redisLruClockResolution & redisLruClockMax)
}

func msTime() int64 {
	return time.Now().UnixNano() / 1000 / 1000
}

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
	initServerConfig()
	initServer()
	elMain()
}

var server = &redisServer{}
