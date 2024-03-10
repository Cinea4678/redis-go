package core

type RedisDb struct {
	Dict    *Dict //存储实际的kv对
	Expires *Dict //存某个key的过期时间
	Id      int
}

func (r *RedisDb) setKey(key *Robj, val *Robj) {
	if r.lookupKey(key) == nil {
		r.dbAdd(key, val)
	} else {
		r.dbOverwrite(key, val)
	}
	val.Refcount++
}

func (r *RedisDb) dbAdd(key *Robj, val *Robj) {
	r.Dict.DictAdd(key.Ptr, val)
}

func (r *RedisDb) dbOverwrite(key *Robj, val *Robj) {
	r.Dict.DictUpdate(key.Ptr, val)
}

// func (r *RedisDb) expireIfNeeded(key *Robj) int {
// 	when := r.getExpire(key)

// 	if when < 0 {
// 		return 0
// 	}
// 	now := msTime()

// 	if now <= when {
// 		return 0
// 	}

// 	return r.dbDelete(key)
// }

func (r *RedisDb) lookupKey(key *Robj) *Robj {
	//检查key是否过期，如果过期则删除
	// r.expireIfNeeded(key)

	return r.doLookupKey(key)
}

func (r *RedisDb) doLookupKey(key *Robj) *Robj {
	entry := r.Dict.DictFind(key.Ptr)
	if entry != nil {
		val := entry.(*Robj)
		//val.Lru = redis.lruClock()
		return val
	}
	return nil
}

func (r *RedisDb) setExpire(key *Robj, expire uint64) {
	kde := r.Dict.DictFind(key.Ptr)
	if kde != nil {
		r.Expires.DictAdd(key.Ptr, expire)
	}
}
