package zset

/*
#cgo CXXFLAGS: -std=c++17
#cgo LDFLAGS: -lstdc++
#include "zset.h"
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

const (
	ZSetOk = iota
	ZSetErr
)

var (
	// 不使用const是因为go需要在编译前确定const的值，而无论使用C中的常量还是宏定义都需要间接使用函数，那么就需要到编译时才能确定
	ZSetNotFound = float64(C.ZSetNotFoundSign()) // 2.225074e-308
	ZSetFound    = float64(C.ZSetSuccessSign())  // 1.797693e+308
)

type ZNode struct {
	score float64
	value string
}

func NewZNode(score float64, value string) ZNode {
	return ZNode{score, value}
}

// ZSet
// 参考HashDict，将对象本体放在objs中，在c中只保存对象在objs中的索引。
type ZSet struct {
	ptr           unsafe.Pointer // 有序列表对象
	objs          []ZNode        // Go对象
	availablePose []int          // objs数组中的可用索引，为了复用删除对象的位置
	v2i           map[string]int // value到index索引...
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

// 查找value对应score
func (zs *ZSet) ZSetGetScore(value string) (float64, error) {
	// score := float64(C.ZSetGetScore(zs.ptr, C.uint(zs.v2i[value])))
	// if score != ZSetNotFound {
	// 	return score, errors.New("Can't get score, value not found")
	// }
	// return score, nil
	index, ok := zs.v2i[value]
	if !ok {
		return ZSetNotFound, errors.New("Can't get score, value not found")
	}
	return zs.objs[index].score, nil
}

func (zs *ZSet) ZSetAdd(score float64, value string) float64 {
	pos := len(zs.objs)
	length := len(zs.availablePose)
	if length > 0 {
		pos = zs.availablePose[length-1]
		zs.objs[pos] = ZNode{score, value}
		zs.v2i[value] = pos
		zs.availablePose = zs.availablePose[:length-1]
	} else {
		zs.objs = append(zs.objs, ZNode{score, value})
	}

	// XXX: int转uint应该没事？
	return float64(C.ZSetAdd(zs.ptr, C.double(score), C.uint(pos)))
}

func (zs *ZSet) ZSetRemoveValue(value string) int {
	pos := int(C.ZSetRemoveValue(zs.ptr, C.uint(zs.v2i[value])))
	if pos >= 0 {
		zs.objs[pos].score = ZSetNotFound
		zs.objs[pos].value = ""
		zs.availablePose = append(zs.availablePose, pos)
		zs.v2i[value] = -1
		return ZSetOk
	} else {
		return ZSetErr
	}
}

func (zs *ZSet) ZSetRemoveScore(score float64) {
	arrPtr := C.ZSetRemoveScore(zs.ptr, C.double(score))
	// TODO: 这里的长度如何获取？
	slice := (*[1 << 30]uint32)(unsafe.Pointer(arrPtr))[:length:length]

	for pos := range slice {
		if pos >= 0 {
			zs.objs[pos].score = ZSetNotFound
			zs.objs[pos].value = ""
			zs.availablePose = append(zs.availablePose, pos)
			zs.v2i[value] = -1
			return ZSetOk
		} else {
			return ZSetErr
		}
	}
}

func (zs *ZSet) ZSetSearch(value string) {

}
