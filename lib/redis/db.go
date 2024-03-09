package redis

import (
	"redis-go/lib/redis/ds"
)

type redisDb struct {
	dict    *ds.Dict //存储实际的kv对
	expires *ds.Dict //存某个key的过期时间
	id      int
}

func (r *redisDb) setKey(key *ds.Robj, val *ds.Robj) {
	if r.lookupKey(key) == nil {
		r.dbAdd(key, val)
	} else {
		r.dbOverwrite(key, val)
	}
	val.Refcount++
}

func (r *redisDb) dbAdd(key *ds.Robj, val *ds.Robj) {
	r.dict.DictAdd(key.Ptr, val)
}

func (r *redisDb) dbOverwrite(key *ds.Robj, val *ds.Robj) {
	r.dict.DictUpdate(key.Ptr, val)
}

// func (r *redisDb) expireIfNeeded(key *robj) int {
// 	when := r.getExpire(key)

// 	if when < 0 {
// 		return 0
// 	}
// 	now := mstime()

// 	if now <= when {
// 		return 0
// 	}

// 	return r.dbDelete(key)
// }

func (r *redisDb) lookupKey(key *ds.Robj) *ds.Robj {
	//检查key是否过期，如果过期则删除
	// r.expireIfNeeded(key)

	return r.doLookupKey(key)
}

func (r *redisDb) doLookupKey(key *ds.Robj) *ds.Robj {
	entry := r.dict.DictFind(key.Ptr)
	if entry != nil {
		val := entry.(*ds.Robj)
		val.Lru = lruClock()
		return val
	}
	return nil
}

func (r *redisDb) setExpire(key *ds.Robj, expire uint64) {
	kde := r.dict.DictFind(key.Ptr)
	if kde != nil {
		r.expires.DictAdd(key.Ptr, expire)
	}
}
