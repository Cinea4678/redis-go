package core

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"github.com/bytecodealliance/wasmtime-go/v20"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

// #include <string.h>
import "C"

var engine *wasmtime.Engine

type WasmAddr = int32

// WasmPlugin wasm插件
type WasmPlugin struct {
	Name     string                     // 扩展名称
	Module   *wasmtime.Module           // Module，在启动时创建
	Commands []string                   // 声明处理的命令
	Store    map[int]*wasmtime.Store    // Wasm Store, 在运行时创建 (key是Db的Id)
	Linkers  map[int]*wasmtime.Linker   // Key是Db的Id
	Instance map[int]*wasmtime.Instance // Key是Db的Id, 直接复用, 不隔离了
}

var Plugins []WasmPlugin

// WasmProvideInfo 由wasm模块提供的元数据
type WasmProvideInfo struct {
	Name     string   `json:"name"`     // 名称
	Commands []string `json:"commands"` // 可以处理的命令
}

// 初始化Wasm运行时
func InitWasmRuntime() {
	config := wasmtime.NewConfig()
	config.SetConsumeFuel(true)
	engine = wasmtime.NewEngine()
}

func ExtensionHandleRequest(client *RedisClient, plugin *WasmPlugin) (string, error) {
	db := client.Db
	dbi := db.Id

	var err error

	module := plugin.Module

	var store *wasmtime.Store
	var linker *wasmtime.Linker
	if plugin.Store[dbi] == nil {
		store := wasmtime.NewStore(engine)
		linker = getLinker(db, store)

		plugin.Store[dbi] = store
		plugin.Linkers[dbi] = linker
	} else {
		store = plugin.Store[dbi]
		linker = plugin.Linkers[dbi]
	}
	instance, err := linker.Instantiate(store, module)
	if err != nil {
		return "", err
	}

	rawReq := client.RawReq
	handleFunc := instance.GetFunc(store, "handle")

	resp, err := handleFunc.Call(store, rawReq)
	if err != nil {
		return "", err
	}

	return resp.(string), nil
}

// checkWasmModuleValid 检查Wasm模块是否导出了必须的函数
func checkWasmModuleValid(module *wasmtime.Module) (err error) {
	realExports := make(map[string]bool)

	for _, e := range module.Exports() {
		if e.Type().FuncType() != nil {
			realExports[e.Name()] = true
		}
	}

	funcList := []string{"handle", "info", "alloc", "free"}
	for _, f := range funcList {
		if realExports[f] == false {
			return errors.New("function " + f + " is not exported by wasm.")
		}
	}
	return
}

func getLinker(db *RedisDb, store *wasmtime.Store) *wasmtime.Linker {
	memory, _ := wasmtime.NewMemory(store, wasmtime.NewMemoryType(12*1024, true, 12*1024))

	getStr := func(addr WasmAddr) string {
		unsafePtr := unsafe.Pointer(uintptr(memory.Data(store)) + uintptr(addr))
		cStr := (*C.char)(unsafePtr)
		l := C.strlen(cStr)
		return unsafe.String((*byte)(unsafePtr), l)
	}

	setKey := func(key string, val string) {
		db.SetKey(key, CreateString(val))
	}

	setKeyInt := func(keyAddr WasmAddr, val int64) {
		key := getStr(keyAddr)
		db.SetKey(key, CreateInteger(val))
	}

	deleteKey := func(key string) {
		db.DbDelete(key)
	}

	getKey := func(key string) string {
		if res, err := db.LookupKey(key).GetString(); err != nil {
			return res
		} else {
			return ""
		}
	}

	getKeyInt := func(keyAddr WasmAddr) int64 {
		key := getStr(keyAddr)

		if res, err := db.LookupKey(key).GetInteger(); err != nil {
			return res
		} else {
			return 0
		}
	}

	setExpire := func(key string, expire int64) {
		db.SetExpire(key, expire)
	}

	getExpire := func(key string) int64 {
		if res, ok := db.GetExpire(key); ok {
			return res
		} else {
			return 0
		}
	}

	linker := wasmtime.NewLinker(store.Engine)
	linker.DefineFunc(store, "env", "set_key", setKey)
	linker.DefineFunc(store, "env", "set_key_int", setKeyInt)
	linker.DefineFunc(store, "env", "delete_key", deleteKey)
	linker.DefineFunc(store, "env", "get_key", getKey)
	linker.DefineFunc(store, "env", "get_key_int", getKeyInt)
	linker.DefineFunc(store, "env", "get_expire", getExpire)
	linker.DefineFunc(store, "env", "set_expire", setExpire)
	linker.Define(store, "env", "memory", memory)

	return linker
}

// 初始化Wasm插件列表
func InitWasmPlugins() {
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
	err = filepath.Walk(pluginDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".wasm" {
			filename := filepath.Base(path)
			if strings.HasSuffix(filename, "-ext.wasm") {
				data, err := os.ReadFile(path)
				if err != nil {
					log.Warn().Str("path", path).Msg("reading plugin failed")
					return nil
				}
				wg.Add(1)
				go loadWasmInfo(data, filename, &wg)
			}
		}
		return nil
	})
	wg.Wait()

	log.Info().Int("pluginNum", len(Plugins)).Msg("load plugin finished")
}

func loadWasmInfo(wasmBin []byte, filename string, wg *sync.WaitGroup) {
	defer wg.Done()

	module, err := wasmtime.NewModule(engine, wasmBin)
	if err != nil {
		log.Warn().Str("filename", filename).Err(err).Msg("load wasm plugin failed")
		return
	}

	err = checkWasmModuleValid(module)
	if err != nil {
		log.Warn().Str("filename", filename).Err(err).Msg("load wasm plugin failed")
		return
	}

	// 获取需要的信息
	store := wasmtime.NewStore(engine)
	linker := getLinker(nil, store)
	instance, err := linker.Instantiate(store, module)
	if err != nil {
		log.Warn().Str("filename", filename).Err(err).Msg("load wasm plugin failed")
		return
	}

	infoFunc := instance.GetFunc(store, "info")
	ret, err := infoFunc.Call(store)
	if err != nil {
		log.Warn().Str("filename", filename).Err(err).Msg("getting information of wasm plugin failed")
		return
	}

	var provideInfo WasmProvideInfo
	err = jsoniter.UnmarshalFromString(ret.(string), &provideInfo)
	if err != nil {
		log.Warn().Str("filename", filename).Err(err).Str("metadata", ret.(string)).Msg("wasm plugin is not supported")
	}

	plugin := WasmPlugin{
		Name:     provideInfo.Name,
		Commands: provideInfo.Commands,
		Module:   module,
	}

	Plugins = append(Plugins, plugin)
}
