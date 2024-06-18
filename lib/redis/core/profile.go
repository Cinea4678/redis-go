package core

import "runtime"

type Profile struct {
	MemStat  *runtime.MemStats // 系统内存情况，按需调用
	TimeCost float64           // 耗时
}
