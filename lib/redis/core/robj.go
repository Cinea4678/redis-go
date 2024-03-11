package core

const (
	RedisString = iota
	//redisList
	//redisSet
)

type Robj struct {
	Rtype uint8
	Lru   uint64
	Ptr   interface{}
}

func CreateString(str *string) *Robj {
	return CreateObject(RedisString, str)
}

func CreateObject(t uint8, ptr interface{}) *Robj {
	return &Robj{
		Rtype: t,
		Ptr:   ptr,
		Lru:   0,
	}
}
