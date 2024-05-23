package ziplist

/*
#cgo CXXFLAGS: -std=c++17
#cgo LDFLAGS: -lstdc++
#include "ziplist.h"
*/
import "C"
import (
	"runtime"
	"unsafe"
)

// Ziplist 是压缩列表的 Go 结构体
type Ziplist struct {
	ptr unsafe.Pointer
}

type ZiplistNode struct {
	ptr unsafe.Pointer
}

// NewZiplist 创建一个新的压缩列表
func NewZiplist() *Ziplist {
	ptr := C.NewZiplist()
	l := &Ziplist{ptr: ptr}
	runtime.SetFinalizer(l, func(l *Ziplist) {
		C.ReleaseZiplist(l.ptr)
	})
	return l
}

// PushBytes 向压缩列表中推送字节数组
func (zl *Ziplist) PushBytes(bytes []byte) int {
	return int(C.ZiplistPushBytes(zl.ptr, (*C.char)(unsafe.Pointer(&bytes[0])), C.int(len(bytes))))
}

// PushInteger 向压缩列表中推送整数
func (zl *Ziplist) PushInteger(value int64) int {
	return int(C.ZiplistPushInteger(zl.ptr, C.int64_t(value)))
}

// InsertInteger inserts an integer at a specified position in the ziplist
func (zl *Ziplist) InsertInteger(pos int, value int64) int {
	return int(C.ZiplistInsertInteger(zl.ptr, C.int(pos), C.longlong(value)))
}

// InsertBytes inserts a byte array at a specified position in the ziplist
func (zl *Ziplist) InsertBytes(pos int, bytes []byte) int {
	return int(C.ZiplistInsertBytes(zl.ptr, C.int(pos), (*C.char)(unsafe.Pointer(&bytes[0])), C.int(len(bytes))))
}

func (zl *Ziplist) Index(index int) *ZiplistNode {
	zlLen := zl.Len()
	if index > zlLen {
		return nil
	}
	nodePtr := C.ZiplistIndex(zl.ptr, C.int(index))
	if nodePtr == nil {
		return nil
	}
	return &ZiplistNode{ptr: nodePtr}
}

func (zl *Ziplist) FindBytes(bytes []byte) *ZiplistNode {
	nodePtr := C.ZiplistFindBytes(zl.ptr, (*C.char)(unsafe.Pointer(&bytes[0])), C.int(len(bytes)))
	if nodePtr == nil {
		return nil
	}
	return &ZiplistNode{ptr: nodePtr}
}

func (zl *Ziplist) FindInteger(integer int64) *ZiplistNode {
	nodePtr := C.ZiplistFindInteger(zl.ptr, C.int64_t(integer))
	if nodePtr == nil {
		return nil
	}
	return &ZiplistNode{ptr: nodePtr}
}

func (zl *Ziplist) Next(zn *ZiplistNode) *ZiplistNode {
	nextPtr := C.ZiplistNext(zl.ptr, zn.ptr)
	if nextPtr == nil {
		return nil
	}
	return &ZiplistNode{ptr: nextPtr}
}

func (zl *Ziplist) Prev(zn *ZiplistNode) *ZiplistNode {
	prevPtr := C.ZiplistPrev(zl.ptr, zn.ptr)
	if prevPtr == nil {
		return nil
	}
	return &ZiplistNode{ptr: prevPtr}
}

func (zl *ZiplistNode) GetInteger() int64 {
	return int64(C.ZiplistGetInteger(zl.ptr))
}

// GetByteArray retrieves a slice of bytes from the ziplist node
// Note: This function assumes that the C++ side uses a consistent memory allocation strategy
// that Go can free. You might need to adjust memory allocation and deallocation accordingly.
func (zl *ZiplistNode) GetByteArray() []byte {
	var cArray *C.uint8_t
	var cLen C.int
	C.ZiplistGetByteArray(zl.ptr, &cArray, &cLen)
	length := int(cLen)
	// Make a Go slice backed by the C array
	slice := (*[1 << 28]byte)(unsafe.Pointer(cArray))[:length:length]
	return slice
}

func (zl *Ziplist) Delete(node *ZiplistNode) int {
	return int(C.ZiplistDelete(node.ptr))
}

func (zl *Ziplist) DeleteRange(startNode *ZiplistNode, length int) int {
	return int(C.ZiplistDeleteRange(zl.ptr, startNode.ptr, C.int(length)))
}

func (zl *Ziplist) DeleteByPos(pos int) int {
	return int(C.ZiplistDeleteByPos(zl.ptr, C.size_t(pos)))
}

func (zl *Ziplist) BlobLen() int {
	return int(C.ZiplistBlobLen(zl.ptr))
}

func (zl *Ziplist) Len() int {
	return int(C.ZiplistLen(zl.ptr))
}

func (zn *ZiplistNode) IsInteger() bool {
	if val := zn.GetByteArray(); val == nil || len(val) == 0 {
		return true
	} else {
		return false
	}
}

func (zn *ZiplistNode) IsBytes() bool {
	if val := zn.GetByteArray(); val == nil || len(val) == 0 {
		return false
	} else {
		return true
	}
}
