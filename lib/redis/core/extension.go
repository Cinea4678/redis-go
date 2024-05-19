package core

import (
	"sync"
)

type LuaPlugin struct {
	Name     string
	RtDir    string
	Commands []string
	LPool    *sync.Pool
}

var Plugins []*LuaPlugin

// LuaPluginProvideInfo 由插件提供的元数据
type LuaPluginProvideInfo struct {
	Name     string   `json:"name"`     // 名称
	Commands []string `json:"commands"` // 可以处理的命令
}
