package io

import (
	"redis-go/lib/redis/core"

	lua "github.com/yuin/gopher-lua"
)

func PluginHandle(p *core.LuaPlugin, client *core.RedisClient) (err error) {
	L := p.LPool.Get().(*lua.LState)
	defer p.LPool.Put(L)

	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("Handle"),
		NRet:    1,
		Protect: true,
	}, lua.LNumber(client.Db.Id), lua.LString(client.RawReq))
	if err != nil {
		return
	}

	ret := L.ToString(-1)
	L.Pop(1)

	SendRawReplyToClient(client, []byte(ret))

	return
}
