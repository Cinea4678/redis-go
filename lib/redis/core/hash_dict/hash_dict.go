package hash_dict

/*
#cgo CXXFLAGS: -std=c++17
#cgo LDFLAGS: -lstdc++
#include "hash_dict.h"
*/
import "C"
import "unsafe"

const (
	DictOk = iota
	DictErr
)

// HashDict
/**
Go的对象不能完全脱离Go（必须被Go管理），所以我们将对象本体放在objs中，
在哈希表中只保存对象在objs中的索引。
*/
type HashDict struct {
	ptr           unsafe.Pointer // 哈希表对象
	objs          []interface{}  // Go对象
	availablePose []int          // objs数组中的可用索引
}

func NewDict() *HashDict {
	ptr := C.NewHashDict()
	return &HashDict{ptr: ptr}
}

func (d *HashDict) DictAdd(key string, val interface{}) int {
	pos := len(d.objs)
	if len(d.availablePose) > 0 {
		// 存在空余的空间
		pos = d.availablePose[len(d.availablePose)-1]
		d.objs[pos] = val
		d.availablePose = d.availablePose[:len(d.availablePose)-1]
	} else {
		d.objs = append(d.objs, val)
	}

	return int(C.DictAdd(d.ptr, C.CString(key), C.int(pos)))
}

func (d *HashDict) DictRemove(key string) int {
	pos := int(C.DictRemove(d.ptr, C.CString(key)))
	if pos >= 0 {
		d.objs[pos] = nil
		d.availablePose = append(d.availablePose, pos)
		return DictOk
	} else {
		return DictErr
	}
}

func (d *HashDict) DictUpdate(key string, val interface{}) int {
	if pos := d.dictFind(key); pos < 0 {
		return DictErr
	} else {
		d.objs[pos] = val
		return DictOk
	}
}

func (d *HashDict) DictInsertOrUpdate(key string, val interface{}) int {
	if pos := d.dictFind(key); pos < 0 {
		return d.DictAdd(key, val)
	} else {
		d.objs[pos] = val
	}
	return DictOk
}

func (d *HashDict) dictFind(key string) int {
	return int(C.DictFind(d.ptr, C.CString(key)))
}

func (d *HashDict) DictFind(key string) interface{} {
	if pos := d.dictFind(key); pos < 0 {
		return nil
	} else {
		return d.objs[pos]
	}
}
