package core

import (
	"github.com/panjf2000/gnet/v2"
)

type EventLoop struct {
	Traffic func(c gnet.Conn) (action gnet.Action)
	Open    func(c gnet.Conn) (out []byte, action gnet.Action)
	//tick    func() (delay time.Duration, action gnet.Action)

	*gnet.BuiltinEventEngine
}

// OnOpen 当新的客户端连接到服务器
func (e *EventLoop) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	return e.Open(c)
}

// OnTraffic 当有数据流量到达服务器
func (e *EventLoop) OnTraffic(c gnet.Conn) gnet.Action {
	return e.Traffic(c)
}

//// Tick 定时触发
//func (e *EventLoop) Tick() (delay time.Duration, action gnet.Action) {
//	return e.tick()
//}
