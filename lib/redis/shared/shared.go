package shared

import (
	"redis-go/lib/redis/core"

	"github.com/cinea4678/resp3"
	lua "github.com/yuin/gopher-lua"
)

var (
	Shared *ValuesStruct
)

// ValuesStruct 共享对象
type ValuesStruct struct {
	Ok        *resp3.Value
	err       *resp3.Value
	syntaxErr *resp3.Value
	Nil       *resp3.Value
	cZero     *resp3.Value
	cOne      *resp3.Value
	oomErr    *resp3.Value
}

func CreateSharedValues() {
	Shared = &ValuesStruct{
		Ok:        &resp3.Value{Type: resp3.TypeSimpleString, Str: "OK"},
		err:       &resp3.Value{Type: resp3.TypeSimpleError, Str: "ERR"},
		syntaxErr: &resp3.Value{Type: resp3.TypeSimpleError, Str: "-ERR syntax error"},
		Nil:       resp3.NewNullValue(),
		cZero:     &resp3.Value{Type: resp3.TypeNumber, Integer: 0},
		cOne:      &resp3.Value{Type: resp3.TypeNumber, Integer: 1},
		oomErr:    &resp3.Value{Type: resp3.TypeSimpleError, Str: "-OOM command not allowed when used memory > 'maxmemory'"},
	}
}

var Server = &core.RedisServer{}

func RedisApiLuaLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), RedisApiLuaExports)
	L.Push(mod)
	return 1
}

var RedisApiLuaExports = map[string]lua.LGFunction{
	"setKey": func(L *lua.LState) int {
		db := L.ToInt(1)
		key := L.ToString(2)
		val := L.ToString(3)
		Server.Db[db].SetKey(key, core.CreateString(val))
		return 0
	},
	"setKeyInt": func(L *lua.LState) int {
		db := L.ToInt(1)
		key := L.ToString(2)
		val := L.ToInt64(3)
		Server.Db[db].SetKey(key, core.CreateInteger(val))
		return 0
	},
	"deleteKey": func(L *lua.LState) int {
		db := L.ToInt(1)
		key := L.ToString(2)
		Server.Db[db].DbDelete(key)
		return 0
	},
	"getKey": func(L *lua.LState) int {
		db := L.ToInt(1)
		key := L.ToString(2)
		if res, err := Server.Db[db].LookupKey(key).GetString(); err == nil {
			L.Push(lua.LString(res))
		} else {
			L.Push(lua.LNil)
		}
		return 1
	},
	"getKeyInt": func(L *lua.LState) int {
		db := L.ToInt(1)
		key := L.ToString(2)
		if res, err := Server.Db[db].LookupKey(key).GetInteger(); err == nil {
			L.Push(lua.LNumber(res))
		} else {
			L.Push(lua.LNil)
		}
		return 1
	},
	"setExpire": func(L *lua.LState) int {
		db := L.ToInt(1)
		key := L.ToString(2)
		expire := L.ToInt64(3)
		Server.Db[db].SetExpire(key, expire)
		return 0
	},
	"getExpire": func(L *lua.LState) int {
		db := L.ToInt(1)
		key := L.ToString(2)
		if expire, ok := Server.Db[db].GetExpire(key); ok {
			L.Push(lua.LNumber(expire))
		} else {
			L.Push(lua.LNil)
		}
		return 1
	},
}
