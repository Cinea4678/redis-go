package core

import "redis-go/lib/redis/core/hash_dict"

type Dict = hash_dict.HashDict

func NewDict() *Dict {
	return hash_dict.NewDict()
}
