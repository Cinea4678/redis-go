package core

import "strconv"

const (
	redisString = iota
	//redisList
	//redisSet
)

type Robj struct {
	Rtype    uint8
	Encoding uint8
	Lru      uint64
	Refcount uint32
	Ptr      interface{}
}

func createObject(t uint8, ptr interface{}) *Robj {
	return &Robj{
		Rtype:    t,
		Encoding: 0,
		Ptr:      ptr,
		Lru:      0,
		Refcount: 0,
	}
}

func (o *Robj) getUint64() (res uint64, err error) {
	if o != nil {
		if o.Rtype == redisString {
			res, err = strconv.ParseUint(o.Ptr.(string), 10, 64)
		}
	}
	return
}
