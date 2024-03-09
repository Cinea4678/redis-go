package redis

type dict map[interface{}]interface{}

const (
	dictOk = iota
	dictErr
)

func (d dict) DictAdd(key interface{}, val interface{}) int {
	if d[key] == nil {
		d[key] = val
		return dictOk
	} else {
		return dictErr
	}
}

func (d dict) DictRemove(key interface{}) int {
	if d[key] == nil {
		return dictErr
	} else {
		d[key] = nil
		return dictOk
	}
}

func (d dict) DictUpdate(key interface{}, val interface{}) int {
	if d[key] != nil {
		d[key] = val
		return dictOk
	} else {
		return dictErr
	}
}

func (d dict) DictInsertOrUpdate(key interface{}, val interface{}) int {
	d[key] = val
	return dictOk
}

func (d dict) DictFind(key interface{}) interface{} {
	return d[key]
}
