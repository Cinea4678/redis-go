package core

import "redis-go/lib/redis/core/hash_dict"

type Dict = hash_dict.HashDict

func NewDict() *Dict {
	return hash_dict.NewDict()
}

//const (
//	dictOk = iota
//	dictErr
//)
//
//func (d Dict) DictAdd(key interface{}, val interface{}) int {
//	if d[key] == nil {
//		d[key] = val
//		return dictOk
//	} else {
//		return dictErr
//	}
//}
//
//func (d Dict) DictRemove(key interface{}) int {
//	if d[key] == nil {
//		return dictErr
//	} else {
//		d[key] = nil
//		return dictOk
//	}
//}
//
//func (d Dict) DictUpdate(key interface{}, val interface{}) int {
//	if d[key] != nil {
//		d[key] = val
//		return dictOk
//	} else {
//		return dictErr
//	}
//}
//
//func (d Dict) DictInsertOrUpdate(key interface{}, val interface{}) int {
//	d[key] = val
//	return dictOk
//}
//
//func (d Dict) DictFind(key interface{}) interface{} {
//	return d[key]
//}
