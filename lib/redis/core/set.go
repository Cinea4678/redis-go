package core

import (
	"redis-go/lib/redis/core/hash_dict"
	"redis-go/lib/redis/core/intset"
)

const (
	encNone   = iota // 底层未初始化
	encIntset        // 底层类型：整数集合
	encDict          // 底层类型：Dict
)

// Set
/**
根据存储内容自动选择底层的数据结构；
在存储少量整数（32个及以下）时，采用整数集合；否则，采取哈希表

Set只可以存储字符串类型。
*/
type Set struct {
	enc byte
	ptr interface{}
}

// 转换底层格式为哈希表
func (s *Set) intsetToDict() {
	if s.enc != encIntset {
		return
	}
	is := s.ptr.(*intset.Intset)

	intsetLen := is.IntsetLen()
	nums := make([]*Object, 0, intsetLen)
	for i := range intsetLen {
		integer := is.IntsetGet(i)
		nums = append(nums, CreateInteger(integer))
	}

	dict := NewDict()
	for _, v := range nums {
		str, _ := v.GetString()
		dict.DictAdd(str, true)
	}
	s.enc = encDict
	s.ptr = dict
}

// Add 添加元素,在发生重复时不报错.
func (s *Set) Add(obj *Object) (repeat bool, err error) {
	if obj.Type != RedisString {
		return false, errNotString
	}

	// 检查是否需要提升为Dict
	if s.enc == encIntset && (!obj.IsInteger() || s.Size() >= 32) {
		s.intsetToDict()
	}

	if s.enc == encNone {
		if obj.IsInteger() {
			s.enc = encIntset
			s.ptr = intset.NewIntset()
		} else {
			s.enc = encDict
			s.ptr = hash_dict.NewDict()
		}
	}

	if s.enc == encIntset {
		is := s.ptr.(*intset.Intset)
		integer, _ := obj.GetInteger()
		println("SET ADD (integer)", integer)
		repeat = is.IntsetFind(integer) == intset.Ok

		is.IntsetAdd(integer)
	} else if s.enc == encDict {
		dict := s.ptr.(*Dict)
		str, _ := obj.GetString()
		println("SET ADD (string)", str)
		repeat = dict.DictFind(str) != nil

		dict.DictAdd(str, true)
	}

	return
}

// Remove 删除元素，元素不存在时不报错
func (s *Set) Remove(obj *Object) (ok bool, err error) {
	if obj.Type != RedisString {
		return false, errNotString
	}

	if s.enc == encIntset {
		is := s.ptr.(*intset.Intset)
		integer, _ := obj.GetInteger()

		ok = is.IntsetRemove(integer) == intset.Ok
	} else if s.enc == encDict {
		dict := s.ptr.(*Dict)
		str, _ := obj.GetString()

		ok = dict.DictRemove(str) == hash_dict.DictOk
	}
	return
}

// Find 查询元素是否存在
func (s *Set) Find(obj *Object) bool {
	if obj.Type != RedisString {
		return false
	}

	if s.enc == encIntset {
		is := s.ptr.(*intset.Intset)
		integer, _ := obj.GetInteger()

		return is.IntsetFind(integer) == intset.Ok
	} else if s.enc == encDict {
		dict := s.ptr.(*Dict)
		str, _ := obj.GetString()

		return dict.DictFind(str) != nil
	}
	return false
}

// Size 查询元素数量
func (s *Set) Size() int {
	if s.enc == encIntset {
		is := s.ptr.(*intset.Intset)
		return is.IntsetLen()
	} else if s.enc == encDict {
		dict := s.ptr.(*Dict)
		return dict.DictLen()
	}
	return 0
}

func (s *Set) ForEach(callback func(object *Object)) {
	if s.enc == encIntset {
		is := s.ptr.(*intset.Intset)
		isl := is.IntsetLen()
		for i := range isl {
			integer := is.IntsetGet(i)
			obj := CreateInteger(integer)
			callback(obj)
		}
	} else if s.enc == encDict {
		dict := s.ptr.(*Dict)
		dict.ForEach(func(key string, _ interface{}) {
			obj := CreateString(key)
			callback(obj)
		})
	}
}
