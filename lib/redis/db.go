package redis

type redisDb struct {
	dict    *dict //存储实际的kv对
	expires *dict //存某个key的过期时间
	id      int
}

func (r *redisDb) setKey(key *robj, val *robj) {
	if r.lookupKey(key) == nil {
		r.dbAdd(key, val)
	} else {
		r.dbOverwrite(key, val)
	}
	val.Refcount++
}

func (r *redisDb) dbAdd(key *robj, val *robj) {
	r.dict.DictAdd(key.Ptr, val)
}

func (r *redisDb) dbOverwrite(key *robj, val *robj) {
	r.dict.DictUpdate(key.Ptr, val)
}

// func (r *redisDb) expireIfNeeded(key *robj) int {
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

func (r *redisDb) lookupKey(key *robj) *robj {
	//检查key是否过期，如果过期则删除
	// r.expireIfNeeded(key)

	return r.doLookupKey(key)
}

func (r *redisDb) doLookupKey(key *robj) *robj {
	entry := r.dict.DictFind(key.Ptr)
	if entry != nil {
		val := entry.(*robj)
		val.Lru = lruClock()
		return val
	}
	return nil
}

func (r *redisDb) setExpire(key *robj, expire uint64) {
	kde := r.dict.DictFind(key.Ptr)
	if kde != nil {
		r.expires.DictAdd(key.Ptr, expire)
	}
}
