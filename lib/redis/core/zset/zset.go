package zset

/*
#cgo CXXFLAGS: -std=c++17
#cgo LDFLAGS: -lstdc++
#include "zset.h"
*/
import "C"

import (
	"errors"
	"fmt"
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

// 大写，因为要是导出字段

type CZNode struct {
	Score C.double
	Value C.uint
}

type ZNode struct {
	Score float64
	Value string
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
	zset := &ZSet{
		ptr:           ptr,
		v2i:           make(map[string]int), // 初始化 map
		availablePose: make([]int, 0),
	}

	// 注册析构函数
	runtime.SetFinalizer(zset, func(zs *ZSet) {
		C.ReleaseZSet(zs.ptr)
	})

	return zset
}

func (zs *ZSet) Len() int {
	return int(C.ZSetLen(zs.ptr))
}

// 查找value对应score
func (zs *ZSet) ZSetGetScore(value string) (float64, bool) {
	index, ok := zs.v2i[value]
	if !ok {
		return ZSetNotFound, false
	}
	return zs.objs[index].Score, true
}

// bool返回true表示已存在
func (zs *ZSet) ZSetAdd(score float64, value string) (float64, bool) {
	s, exist := zs.ZSetGetScore(value)
	fmt.Println(s)
	// 已存在
	if exist {
		fmt.Println("exist")
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
	C.ZSetAdd(zs.ptr, C.double(score), C.uint(zs.v2i[value]))

	return score, false
}

func (zs *ZSet) ZSetRemoveValue(value string) int {

	// value不存在
	if _, ok := zs.v2i[value]; !ok {
		return ZSetErr
	}

	pos := int(C.ZSetRemoveValue(zs.ptr, C.uint(zs.v2i[value])))
	if pos >= 0 {
		zs.objs[pos].Score = ZSetNotFound
		zs.objs[pos].Value = ""
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
	// C.free(arrPtr)

	for pos := range slice {
		if pos >= 0 {
			zs.objs[pos].Score = ZSetNotFound
			zs.objs[pos].Value = ""
			zs.availablePose = append(zs.availablePose, pos)
			delete(zs.v2i, zs.objs[pos].Value)
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
	// C.free(arrPtr)
	return slice
}

func (zs *ZSet) ZSetSearchRange(lscore, rscore float64) []ZNode {
	var cLen C.int
	// 指针传长度，函数返回值数组
	arrPtr := C.ZSetSearchRange(zs.ptr, C.double(lscore), C.double(rscore), &cLen)
	defer C.free(arrPtr)

	length := int(cLen)
	Cslice := (*[1 << 30]C.uint)(unsafe.Pointer(arrPtr))[:length:length]
	var slice []ZNode
	for _, cindex := range Cslice {
		index := uint(cindex)
		slice = append(slice, zs.objs[index])
	}
	return slice
}

func (zs *ZSet) ZSetUpdate(newscore float64, value string) (float64, error) {
	score, exist := zs.ZSetGetScore(value)
	if !exist {
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
	return value, zs.objs[index].Score, nil
}

func (zs *ZSet) ZSetSearchRankRange(lrank, rrank int) []ZNode {
	var cLen C.int
	// 指针传长度，函数返回值数组
	arrPtr := C.ZSetSearchRankRange(zs.ptr, C.int(lrank), C.int(rrank), &cLen)
	length := int(cLen)
	slice := (*[1 << 30]ZNode)(unsafe.Pointer(arrPtr))[:length:length]
	// C.free(arrPtr)
	return slice
}
