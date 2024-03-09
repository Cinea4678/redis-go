package redis

import (
	"github.com/panjf2000/gnet/v2"
	"log"
	"strconv"
)

type eventloop struct {
	traffic func(c gnet.Conn) (action gnet.Action)
	open    func(c gnet.Conn) (out []byte, action gnet.Action)
	//tick    func() (delay time.Duration, action gnet.Action)

	*gnet.BuiltinEventEngine
}

// OnOpen 当新的客户端连接到服务器
func (e *eventloop) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	return e.open(c)
}

// OnTraffic 当有数据流量到达服务器
func (e *eventloop) OnTraffic(c gnet.Conn) gnet.Action {
	return e.traffic(c)
}

//// Tick 定时触发
//func (e *eventloop) Tick() (delay time.Duration, action gnet.Action) {
//	return e.tick()
//}

func elMain() {
	addr := "tcp://" + server.bindaddr + ":" + strconv.Itoa(server.port)
	log.Printf("Listening at: %s", addr)
	log.Fatal(gnet.Run(server.events, addr,
		gnet.WithMulticore(false), gnet.WithNumEventLoop(1)))
}
