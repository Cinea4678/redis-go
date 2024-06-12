package core

import (
	"errors"
	"math"
	"strconv"
)

const (
	RedisString = iota
	RedisList
	RedisSet
	RedisZSet
)

const (
	ObjectEncodingStr = iota
	ObjectEncodingInt8
	ObjectEncodingInt16
	ObjectEncodingInt32
	ObjectEncodingInt64
	ObjectEncodingFloat32
	ObjectEncodingFloat64
	ObjectEncodingSet
	ObjectEncodingZSet
	ObjectEncodingList
)

var (
	errNotString = errors.New("object not a string")
)

type Object struct {
	Type     byte
	Encoding byte
	Lru      uint64
	Ptr      interface{}
}

func IsInteger(str string) bool {
	for i, ch := range str {
		if ch < '0' || ch > '9' {
			if i == 0 && ch == '-' {
				continue
			}
			return false
		}
	}
	return true
}

func (o *Object) IsInteger() bool {
	switch o.Encoding {
	case ObjectEncodingInt8:
		fallthrough
	case ObjectEncodingInt16:
		fallthrough
	case ObjectEncodingInt32:
		fallthrough
	case ObjectEncodingInt64:
		return true
	default:
	}
	return false
}

func CreateString(str string) *Object {
	if IsInteger(str) {
		integer, _ := strconv.ParseInt(str, 10, 64)
		return CreateInteger(integer)
	}
	return CreateObject(RedisString, ObjectEncodingStr, str)
}

func CreateInteger(integer int64) *Object {
	if integer < 0 {
		if integer > math.MinInt8 {
			return CreateObject(RedisString, ObjectEncodingInt8, int8(integer))
		} else if integer > math.MinInt16 {
			return CreateObject(RedisString, ObjectEncodingInt16, int16(integer))
		} else if integer > math.MinInt32 {
			return CreateObject(RedisString, ObjectEncodingInt32, int32(integer))
		} else {
			return CreateObject(RedisString, ObjectEncodingInt64, integer)
		}
	} else {
		if integer < math.MaxInt8 {
			return CreateObject(RedisString, ObjectEncodingInt8, int8(integer))
		} else if integer < math.MaxInt16 {
			return CreateObject(RedisString, ObjectEncodingInt16, int16(integer))
		} else if integer < math.MaxInt32 {
			return CreateObject(RedisString, ObjectEncodingInt32, int32(integer))
		} else {
			return CreateObject(RedisString, ObjectEncodingInt64, integer)
		}
	}
}

func CreateSet(set *Set) *Object {
	return CreateObject(RedisSet, ObjectEncodingSet, set)
}

func CreateZSet(zset *ZSet) *Object {
	return CreateObject(RedisZSet, ObjectEncodingZSet, zset)
}

func CreateList(list *List) *Object {
	return CreateObject(RedisList, ObjectEncodingList, list)
}

// GetString 获取以字符串形式表示的值
func (o *Object) GetString() (str string, err error) {
	if o == nil {
		return "", nil
	}
	if o.Type != RedisString {
		return "", errNotString
	}
	switch o.Encoding {
	case ObjectEncodingInt8:
		fallthrough
	case ObjectEncodingInt16:
		fallthrough
	case ObjectEncodingInt32:
		fallthrough
	case ObjectEncodingInt64:
		var integer int64
		integer, err = o.GetInteger()
		str = strconv.FormatInt(integer, 10)
	case ObjectEncodingStr:
		str = o.Ptr.(string)
	default:
	}
	return
}

// GetInteger 获取整数值（如果是的话）
func (o *Object) GetInteger() (integer int64, err error) {
	if o == nil {
		return 0, nil
	}
	if o.Type != RedisString {
		return 0, errNotString
	}
	switch o.Encoding {
	case ObjectEncodingInt8:
		integer = int64(o.Ptr.(int8))
	case ObjectEncodingInt16:
		integer = int64(o.Ptr.(int16))
	case ObjectEncodingInt32:
		integer = int64(o.Ptr.(int32))
	case ObjectEncodingInt64:
		integer = o.Ptr.(int64)
	default:
		err = errors.New("not an integer")
	}
	return
}

func CreateObject(t uint8, encoding byte, ptr interface{}) *Object {
	return &Object{
		Type:     t,
		Encoding: encoding,
		Ptr:      ptr,
		Lru:      0,
	}
}
