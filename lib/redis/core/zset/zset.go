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

	zset.v2i = make(map[string]int)

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
		return ZSetNotFound, errors.New("[zs.ZSetGetScore] can't get score, value not found")
	}
	return zs.objs[index].score, nil
}

// bool返回true表示已存在
func (zs *ZSet) ZSetAdd(score float64, value string) (float64, bool) {
	s, _ := zs.ZSetGetScore(value)
	// 已存在
	if s != ZSetNotFound {
		return s, true
	}

	// 不存在
	length := len(zs.availablePose)
	if length > 0 {
		pos := zs.availablePose[length-1]
		zs.objs[pos] = ZNode{score, value}
		zs.v2i[value] = pos
		zs.availablePose = zs.availablePose[:length-1]
		// fmt.Println("zs.objs[", pos, "]", "={", score, value, "}")
		// fmt.Println("zs.v2i[", value, "]", "={", pos, "}")
		// fmt.Println("zs.availablePose", zs.availablePose)
	} else {
		// fmt.Println("add")
		pos := len(zs.objs)
		zs.v2i[value] = pos
		// fmt.Println("zs.v2i[", value, "]", "={", pos, "}")
		zs.objs = append(zs.objs, ZNode{score, value})
	}

	return score, false
}

func (zs *ZSet) ZSetRemoveValue(value string) int {
	pos := int(C.ZSetRemoveValue(zs.ptr, C.uint(zs.v2i[value])))
	if pos >= 0 {
		zs.objs[pos].score = ZSetNotFound
		zs.objs[pos].value = ""
		zs.availablePose = append(zs.availablePose, pos)
		delete(zs.v2i, value)
		return ZSetOk
	} else {
		return ZSetErr
	}
}

func (zs *ZSet) ZSetRemoveScore(score float64) int {
	var cLen C.int
	// 指针传长度，函数返回值数组
	arrPtr := C.ZSetRemoveScore(zs.ptr, C.double(score), &cLen)
	length := int(cLen)
	slice := (*[1 << 30]int)(unsafe.Pointer(arrPtr))[:length:length]

	for pos := range slice {
		if pos >= 0 {
			zs.objs[pos].score = ZSetNotFound
			zs.objs[pos].value = ""
			zs.availablePose = append(zs.availablePose, pos)
			delete(zs.v2i, zs.objs[pos].value)
		} else {
			// 虽然这样写不能保证原子性
			return ZSetErr
		}
	}
	return ZSetOk
}

// 只返回value索引不返回score
func (zs *ZSet) ZSetSearch(score float64) []int {
	var cLen C.int
	// 指针传长度，函数返回值数组
	arrPtr := C.ZSetSearch(zs.ptr, C.double(score), &cLen)
	length := int(cLen)
	slice := (*[1 << 30]int)(unsafe.Pointer(arrPtr))[:length:length]
	return slice
}

func (zs *ZSet) ZSetSearchRange(lscore, rscore float64) []ZNode {
	var cLen C.int
	// 指针传长度，函数返回值数组
	arrPtr := C.ZSetSearchRange(zs.ptr, C.double(lscore), C.double(rscore), &cLen)
	length := int(cLen)
	slice := (*[1 << 30]ZNode)(unsafe.Pointer(arrPtr))[:length:length]
	return slice
}

func (zs *ZSet) ZSetUpdate(newscore float64, value string) (float64, error) {
	score, err := zs.ZSetGetScore(value)
	if err != nil {
		return ZSetNotFound, errors.New("[zs.ZSetUpdate] can't get score, value not found")
	}

	// 更新实际就是移除再插入
	status := zs.ZSetRemoveValue(value)
	if status == ZSetErr {
		return score, errors.New("[zs.ZSetUpdate] remove err")
	}

	zs.ZSetAdd(newscore, value)

	return score, nil
}

func (zs *ZSet) ZSetSearchRank(rank int) (value string, score float64, err error) {
	index := int(C.ZSetSearchRank(zs.ptr, C.int(rank)))
	if index == -1 {
		return "", ZSetNotFound, errors.New("[zs.ZSetSearchRank] not found")
	}
	return value, zs.objs[index].score, nil
}

func (zs *ZSet) ZSetSearchRankRange(lrank, rrank int) []ZNode {
	var cLen C.int
	// 指针传长度，函数返回值数组
	arrPtr := C.ZSetSearchRankRange(zs.ptr, C.int(lrank), C.int(rrank), &cLen)
	length := int(cLen)
	slice := (*[1 << 30]ZNode)(unsafe.Pointer(arrPtr))[:length:length]
	return slice
}
