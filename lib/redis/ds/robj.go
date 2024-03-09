package ds

const (
	RedisString = iota
	RedisList
	RedisSet
)

type Robj struct {
	Rtype    uint8
	Encoding uint8
	Lru      uint64
	Refcount uint32
	Ptr      interface{}
}

func CreateObject(t uint8, ptr interface{}) *Robj {
	return &Robj{
		Rtype:    t,
		Encoding: 0,
		Ptr:      ptr,
		Lru:      0,
		Refcount: 0,
	}
}
