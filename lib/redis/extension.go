package redis

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/shared"
	"sync"

	jsoniter "github.com/json-iterator/go"
	cp "github.com/otiai10/copy"
	"github.com/rs/zerolog/log"
	lua "github.com/yuin/gopher-lua"
)

// InitPlugins 初始化插件列表
func InitPlugins() {
	cur, err := os.Getwd()
	if err != nil {
		log.Warn().Msg("Get current dir failed")
		return
	}

	pluginDir := filepath.Join(cur, "plugin")
	info, err := os.Stat(pluginDir)
	if os.IsNotExist(err) {
		log.Info().Msg("plugin directory not exists, skip loading plugin")
		return
	}
	if !info.IsDir() {
		log.Warn().Msg("./plugin is not a directory, skip loading plugin")
		return
	}

	var wg sync.WaitGroup
	err = filepath.WalkDir(pluginDir, func(path string, d fs.DirEntry, err error) error {
		if d.Type().IsDir() {
			pluginDirName := d.Name()
			pluginFile := filepath.Join(path, "plugin.lua")
			pluginInfo, err := os.Stat(pluginFile)
			if err == nil && !pluginInfo.IsDir() {
				log.Info().Str("name", pluginDirName).Msg("find plugin")

				rtPath := prepareExecDir(path, pluginDirName)

				wg.Add(1)
				go loadPlugin(rtPath, &wg)
			}
		}
		return nil
	})
	wg.Wait()

	log.Info().Int("pluginNum", len(core.Plugins)).Msg("load plugin finished")
}

// prepareExecDir 准备运行时目录，返回运行时中的插件path目录
func prepareExecDir(pluginDir string, dirName string) string {
	cur, _ := os.Getwd()
	defDep := filepath.Join(cur, "plugin-dep")

	pluginRtDir := filepath.Join(os.TempDir(), "redis-go", "plugin-rt", dirName)
	os.RemoveAll(pluginRtDir)
	os.MkdirAll(pluginRtDir, os.ModePerm)

	cp.Copy(pluginDir, pluginRtDir)
	cp.Copy(defDep, pluginRtDir)

	// 如果有存根，就删掉，以免和运行时注入的冲突
	apiPath := filepath.Join(pluginRtDir, "redisApi.lua")
	os.RemoveAll(apiPath)

	return pluginRtDir
}

// loadPlugin 读取插件信息、初始化
func loadPlugin(pluginRtDir string, wg *sync.WaitGroup) {
	defer wg.Done()

	L := lua.NewState()

	L.DoString(fmt.Sprintf("package.path = package.path .. ';%s/?.lua'", pluginRtDir))
	pluginPath := filepath.Join(pluginRtDir, "plugin.lua")

	L.PreloadModule("redisApi", shared.RedisApiLuaLoader)
	if err := L.DoFile(pluginPath); err != nil {
		log.Error().Err(err).Msg("init plugin failed")
		return
	}

	if err := checkPluginValid(L); err != nil {
		log.Error().Err(err).Msg("init plugin failed")
		return
	}

	// 获取信息
	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("Info"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		log.Error().Err(err).Msg("init plugin failed")
		return
	}
	infoRaw := L.ToString(-1)
	L.Pop(1)

	var provideInfo core.LuaPluginProvideInfo
	if err := jsoniter.UnmarshalFromString(infoRaw, &provideInfo); err != nil {
		log.Error().Err(err).Str("metadata", infoRaw).Msg("init plugin failed")
		return
	}

	p := core.LuaPlugin{
		Name:     provideInfo.Name,
		Commands: provideInfo.Commands,
		RtDir:    pluginRtDir,
		LPool: &sync.Pool{
			New: func() interface{} {
				L := lua.NewState()
				L.DoString(fmt.Sprintf("package.path = package.path .. ';%s/?.lua'", pluginRtDir))
				pluginPath := filepath.Join(pluginRtDir, "plugin.lua")
				L.PreloadModule("redisApi", shared.RedisApiLuaLoader)
				L.DoFile(pluginPath)
				return L
			},
		},
	}
	log.Info().Str("name", p.Name).Msg("load plugin success")

	core.Plugins = append(core.Plugins, &p)
}

func checkPluginValid(L *lua.LState) error {
	flist := []string{"Handle", "Info"}
	for _, f := range flist {
		if !luaGlobalFuncExist(L, f) {
			return errors.New("function " + f + " is not found")
		}
	}
	return nil
}

// luaGlobalFuncExist 工具函数，检查lua全局函数存在
func luaGlobalFuncExist(L *lua.LState, fname string) bool {
	v := L.GetGlobal(fname)
	if v.Type() == lua.LTNil {
		return false
	}
	_, ok := v.(*lua.LFunction)
	return ok
}
