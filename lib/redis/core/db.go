package core

type RedisDb struct {
	Dict    *Dict //存储实际的kv对
	Expires *Dict //存某个key的过期时间
	Id      int
}

func (r *RedisDb) SetKey(key string, val *Object) {
	if r.LookupKey(key) == nil {
		r.DbAdd(key, val)
	} else {
		r.DbOverwrite(key, val)
	}
}

func (r *RedisDb) DbAdd(key string, val *Object) {
	r.Dict.DictAdd(key, val)
}

func (r *RedisDb) DbDelete(key string) {
	r.Dict.DictRemove(key)
}

func (r *RedisDb) DbOverwrite(key string, val *Object) {
	r.Dict.DictUpdate(key, val)
}

func (r *RedisDb) expireIfNeeded(key string) {
	when, ok := r.GetExpire(key)

	if !ok || when < 0 {
		return
	}
	now := GetTimeUnixMilli()

	if now <= when {
		return
	}

	r.DbDelete(key)
}

func (r *RedisDb) LookupKey(key string) *Object {
	// 检查key是否过期，如果过期则删除
	r.expireIfNeeded(key)

	return r.doLookupKey(key)
}

func (r *RedisDb) doLookupKey(key string) *Object {
	entry := r.Dict.DictFind(key)
	if entry != nil {
		val := entry.(*Object)
		//val.Lru = redis.lruClock()
		return val
	}
	return nil
}

// 查找并删除
// XXX: 其他操作比如expire是否也可以优化？而不是查询多次
func (r *RedisDb) LookupKeyDel(key string) *Object {
	entry := r.Dict.DictFindDel(key)
	if entry != nil {
		return entry.(*Object)
	}
	return nil
}

func (r *RedisDb) SetExpire(key string, expire int64) {
	kde := r.Expires.DictFind(key)
	if kde == nil {
		r.Expires.DictAdd(key, expire)
	} else {
		r.Expires.DictUpdate(key, expire)
	}
}

func (r *RedisDb) GetExpire(key string) (time int64, ok bool) {
	res := r.Expires.DictFind(key)
	if res == nil {
		return 0, false
	} else {
		return res.(int64), true
	}
}
