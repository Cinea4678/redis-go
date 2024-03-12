package intset

/*
#cgo CXXFLAGS: -std=c++17
#cgo LDFLAGS: -lstdc++
#include "intset.h"
*/
import "C"

import (
	"runtime"
	"unsafe"
)

const (
	Ok = iota
	Err
)

// Intset 整数集合
/**
用于存储少量的整数
*/
type Intset struct {
	ptr unsafe.Pointer
}

func NewIntset() *Intset {
	ptr := C.NewIntset()

	s := &Intset{ptr: ptr}

	runtime.SetFinalizer(s, func(s *Intset) {
		C.ReleaseIntset(s.ptr)
	})

	return s
}

// IntsetAdd 创建一个新的整数集合
func (s *Intset) IntsetAdd(val int64) int {
	return int(C.IntsetAdd(s.ptr, C.longlong(val)))
}

// IntsetRemove 删除指定元素
func (s *Intset) IntsetRemove(val int64) int {
	return int(C.IntsetRemove(s.ptr, C.longlong(val)))
}

// IntsetFind 查找指定元素是否存在（返回IntsetOk或IntsetErr）
func (s *Intset) IntsetFind(val int64) int {
	return int(C.IntsetFind(s.ptr, C.longlong(val)))
}

// IntsetRandom 返回随机元素
func (s *Intset) IntsetRandom() int64 {
	return int64(C.IntsetRandom(s.ptr))
}

// IntsetGet 获取指定位置的元素
func (s *Intset) IntsetGet(index int) int64 {
	return int64(C.IntsetGet(s.ptr, C.int(index)))
}

// IntsetLen 获取元素数量
func (s *Intset) IntsetLen() int {
	return int(C.IntsetLen(s.ptr))
}

// IntsetBlobLen 获取元素占据空间大小（字节）
func (s *Intset) IntsetBlobLen() int {
	return int(C.IntsetBlobLen(s.ptr))
}
