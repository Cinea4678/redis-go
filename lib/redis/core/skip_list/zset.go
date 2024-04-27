package zset

/*
#cgo CXXFLAGS: -std=c++17
#cgo LDFLAGS: -lstdc++
#include "zset.h"
*/
import "C"

import (
	"runtime"
	"unsafe"
)

var (
	// 不使用const是因为go需要在编译前确定const的值，而无论使用C中的常量还是宏定义都需要间接使用函数，那么就需要到编译时才能确定
	ZSetNotFound = float64(C.ZSetNotFoundSign()) // 2.225074e-308
	ZSetErr      = float64(C.ZSetSuccessSign())  // 1.797693e+308
)

// ZSet
// 参考HashDict，将对象本体放在objs中，在c中只保存对象在objs中的索引。
type ZSet struct {
	ptr           unsafe.Pointer // 有序列表对象
	objs          []interface{}  // Go对象
	availablePose []int          // objs数组中的可用索引，为了复用删除对象的位置
}

func NewZSet() *ZSet {
	ptr := C.NewZSet()
	zset := &ZSet{ptr: ptr}

	// 注册析构函数
	runtime.SetFinalizer(zset, func(zs *ZSet) {
		C.ReleaseZSet(zs.ptr)
	})

	return zset
}

func (zs *ZSet) ZSetAdd(score float64, val interface{}) float64 {
	pos := len(zs.objs)
	length := len(zs.availablePose)
	if length > 0 {
		pos = zs.availablePose[length-1]
		zs.objs[pos] = val
		zs.availablePose = zs.availablePose[:length-1]
	} else {
		zs.objs = append(zs.objs, val)
	}

	return float64(C.ZSetAdd(zs.ptr, C.double(score), C.int(pos)))
}
