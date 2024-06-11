package core

import "redis-go/lib/redis/core/zip_list"

type List = ziplist.Ziplist

func NewList() *List {
	return ziplist.NewZiplist()
}
