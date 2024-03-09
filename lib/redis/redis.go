package redis

import (
	"os"
	"redis-go/lib/redis/ds"
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

var (
	shared *sharedObjectsStruct
)

type redisServer struct {
	pid int

	hz       int // hz
	db       *redisDb
	commands *ds.Dict //redis命令字典 string(如get/set) : *redisCommand

	clientCounter int      //存储client的id计数器
	clients       *ds.Dict //客户端字典 id : *redisClient

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
	name *ds.Robj
	argc int        // 命令数量
	argv []*ds.Robj // 命令值

	cmd     *redisCommand // 当前执行的命令
	lastCmd *redisCommand // 最后执行的命令

	reqtype  int    // 请求类型
	queryBuf string // 从客户端读到的数据

	buf     []byte // 准备发回客户端的数据
	bufpos  int    // 发回数据的pos
	sentlen int    // 已发送的字节数

	flags int //处理标记
}

// redis命令结构
type redisCommand struct {
	name            string                    // 命令名称
	redisClientFunc func(client *redisClient) // 命令处理函数
}

// 共享对象
type sharedObjectsStruct struct {
	crlf      *ds.Robj
	ok        *ds.Robj
	err       *ds.Robj
	syntaxerr *ds.Robj
	nullbulk  *ds.Robj
	czero     *ds.Robj
	cone      *ds.Robj
	oomerr    *ds.Robj
}

// 初始化server配置
func initServerConfig() {
	server.port = redisServerPort
	server.tcpBacklog = redisTcpBacklog
	// TODO server.events = &eventloop{}

}

// 初始化server
func initServer() {
	server.pid = os.Getpid()

	server.clients = &ds.Dict{}

	// 初始化事件处理器
	server.events.traffic = dataHandler
	server.events.open = acceptHandler
	server.events.tick = func() (delay time.Duration, action gnet.Action) {
		return serverCron(), action
	}

	server.db = &redisDb{
		dict:    &ds.Dict{},
		expires: &ds.Dict{},
		id:      1,
	}

	createSharedObjects()
}

func generateClientId() int {
	server.clientCounter++
	return server.clientCounter
}

func createSharedObjects() {
	shared = &sharedObjectsStruct{
		crlf:      ds.CreateObject(ds.RedisString, string("\r\n")),
		ok:        ds.CreateObject(ds.RedisString, string("+OK\r\n")),
		err:       ds.CreateObject(ds.RedisString, string("-ERR\r\n")),
		syntaxerr: ds.CreateObject(ds.RedisString, string("-ERR syntax error\r\n")),
		nullbulk:  ds.CreateObject(ds.RedisString, string("$-1\r\n")),
		czero:     ds.CreateObject(ds.RedisString, string(":0\r\n")),
		cone:      ds.CreateObject(ds.RedisString, string(":1\r\n")),
		oomerr:    ds.CreateObject(ds.RedisString, string("-OOM command not allowed when used memory > 'maxmemory'.\r\n")),
	}
}

// 处理命令
func processCommand(client *redisClient) int { return 0 }

func lruClock() uint64 {
	if 1000/server.hz <= redisLruClockResolution {
		return server.lruclock
	}
	return getLruClock()
}

func getLruClock() uint64 {
	return uint64(mstime() / redisLruClockResolution & redisLruClockMax)
}

func mstime() int64 {
	return int64(time.Now().UnixNano() / 1000 / 1000)
}

func ustime() int64 {
	return int64(time.Now().UnixNano() / 1000)
}

func serverCron() time.Duration {
	server.lruclock = getLruClock()
	databasesCron()
	return time.Millisecond * time.Duration(1000/server.hz)
}

var server = &redisServer{}
