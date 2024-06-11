package main

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func setupRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6389",
		Password: "",
		DB:       0,
	})
}

var ctx = context.Background()

func teardown(rdb *redis.Client) {
	rdb.Close()
}

func TestKeyCommands(t *testing.T) {
	rdb := setupRedisClient()
	defer teardown(rdb)

	// Test DEL
	err := rdb.Set(ctx, "key1", "value1", 0).Err()
	if err != nil {
		t.Errorf("Setup error: %v", err)
	}
	delRes, err := rdb.Del(ctx, "key1").Result()
	if err != nil || delRes != 1 {
		t.Errorf("DEL failed, err: %v", err)
	}

	// Test EXISTS
	existsRes, err := rdb.Exists(ctx, "key1").Result()
	if err != nil || existsRes != 0 {
		t.Errorf("EXISTS failed, err: %v", err)
	}

	// Test EXPIRE
	err = rdb.Set(ctx, "key2", "value2", 0).Err()
	if err != nil {
		t.Errorf("Setup error: %v", err)
	}
	expireRes, err := rdb.Expire(ctx, "key2", 10*time.Second).Result()
	if err != nil || !expireRes {
		t.Errorf("EXPIRE failed, err: %v", err)
	}

	// Test EXPIREAT
	expireAtRes, err := rdb.ExpireAt(ctx, "key2", time.Now().Add(10*time.Second)).Result()
	if err != nil || !expireAtRes {
		t.Errorf("EXPIREAT failed, err: %v", err)
	}

	// Test KEYS
	err = rdb.Set(ctx, "key3", "value3", 0).Err()
	keysRes, err := rdb.Keys(ctx, "key*").Result()
	if err != nil || len(keysRes) < 1 {
		t.Errorf("KEYS failed, err: %v", err)
	}

	// Test MOVE
	moveRes, err := rdb.Move(ctx, "key3", 1).Result()
	if err != nil || !moveRes {
		t.Errorf("MOVE failed, err: %v", err)
	}

	// Test PERSIST
	err = rdb.Set(ctx, "key4", "value4", 0).Err()
	rdb.Expire(ctx, "key4", 10*time.Second)
	persistRes, err := rdb.Persist(ctx, "key4").Result()
	if err != nil || !persistRes {
		t.Errorf("PERSIST failed, err: %v", err)
	}

	// Test PTTL
	pttlRes, err := rdb.PTTL(ctx, "key4").Result()
	if err != nil || pttlRes <= 0 {
		t.Errorf("PTTL failed, err: %v", err)
	}

	// Test RENAME
	err = rdb.Set(ctx, "key5", "value5", 0).Err()
	renameRes, err := rdb.Rename(ctx, "key5", "newkey5").Result()
	if err != nil || renameRes != "OK" {
		t.Errorf("RENAME failed, err: %v", err)
	}

	// Test RENAMENX
	err = rdb.Set(ctx, "key6", "value6", 0).Err()
	rdb.Del(ctx, "newkey6")
	renamenxRes, err := rdb.RenameNX(ctx, "key6", "newkey6").Result()
	if err != nil || !renamenxRes {
		t.Errorf("RENAMENX failed, err: %v", err)
	}

	// Test TYPE
	err = rdb.Set(ctx, "key7", "value7", 0).Err()
	typeRes, err := rdb.Type(ctx, "key7").Result()
	if err != nil || typeRes != "string" {
		t.Errorf("TYPE failed, err: %v", err)
	}
}

func TestStringCommands(t *testing.T) {
	rdb := setupRedisClient()
	defer teardown(rdb)

	// Test SET
	err := rdb.Set(ctx, "key1", "hello", 0).Err()
	if err != nil {
		t.Errorf("SET error: %v", err)
	}

	// Test GET
	value, err := rdb.Get(ctx, "key1").Result()
	if err != nil || value != "hello" {
		t.Errorf("GET error: %v, value: %s", err, value)
	}

	// Test GETRANGE
	subStr, err := rdb.GetRange(ctx, "key1", 1, 3).Result()
	if err != nil || subStr != "ell" {
		t.Errorf("GETRANGE error: %v, subStr: %s", err, subStr)
	}

	// Test GETSET
	oldValue, err := rdb.GetSet(ctx, "key1", "world").Result()
	if err != nil || oldValue != "hello" {
		t.Errorf("GETSET error: %v, oldValue: %s", err, oldValue)
	}

	// Test GETBIT
	bit, err := rdb.GetBit(ctx, "key1", 0).Result()
	if err != nil || bit != 1 {
		t.Errorf("GETBIT error: %v, bit: %d", err, bit)
	}

	// Test MGET
	rdb.Set(ctx, "key2", "world", 0)
	values, err := rdb.MGet(ctx, "key1", "key2").Result()
	if err != nil || len(values) != 2 {
		t.Errorf("MGET error: %v", err)
	}

	// Test SETBIT
	setBitRes, err := rdb.SetBit(ctx, "key1", 7, 1).Result()
	if err != nil || setBitRes != 1 {
		t.Errorf("SETBIT error: %v, res: %d", err, setBitRes)
	}

	// Test SETEX
	err = rdb.SetEX(ctx, "key3", "timeout", 10).Err()
	if err != nil {
		t.Errorf("SETEX error: %v", err)
	}

	// Test SETNX
	setNxRes, err := rdb.SetNX(ctx, "key4", "new", 0).Result()
	if err != nil || !setNxRes {
		t.Errorf("SETNX error: %v, res: %t", err, setNxRes)
	}

	// Test SETRANGE
	setRangeRes, err := rdb.SetRange(ctx, "key1", 0, "Hi").Result()
	if err != nil || setRangeRes != 5 {
		t.Errorf("SETRANGE error: %v, newLen: %d", err, setRangeRes)
	}

	// Test STRLEN
	strLen, err := rdb.StrLen(ctx, "key1").Result()
	if err != nil || strLen != 5 {
		t.Errorf("STRLEN error: %v, len: %d", err, strLen)
	}

	// Test MSET
	msetRes, err := rdb.MSet(ctx, "key1", "val1", "key2", "val2").Result()
	if err != nil || msetRes != "OK" {
		t.Errorf("MSET error: %v, res: %s", err, msetRes)
	}

	// Test MSETNX
	msetnxRes, err := rdb.MSetNX(ctx, "key3", "val3", "key4", "val4").Result()
	if err != nil || !msetnxRes {
		t.Errorf("MSETNX error: %v, res: %t", err, msetnxRes)
	}

	// // Test PSETEX
	// err = rdb.PSetEX(ctx, "key5", "temp", 5000).Err()
	// if err != nil {
	// 	t.Errorf("PSETEX error: %v", err)
	// }

	// Test INCR
	incrRes, err := rdb.Incr(ctx, "num1").Result()
	if err != nil || incrRes != 1 {
		t.Errorf("INCR error: %v, res: %d", err, incrRes)
	}

	// Test INCRBY
	incrByRes, err := rdb.IncrBy(ctx, "num1", 10).Result()
	if err != nil || incrByRes != 11 {
		t.Errorf("INCRBY error: %v, res: %d", err, incrByRes)
	}

	// Test INCRBYFLOAT
	incrByFloatRes, err := rdb.IncrByFloat(ctx, "num2", 1.5).Result()
	if err != nil || incrByFloatRes != 1.5 {
		t.Errorf("INCRBYFLOAT error: %v, res: %f", err, incrByFloatRes)
	}

	// Test DECR
	decrRes, err := rdb.Decr(ctx, "num1").Result()
	if err != nil || decrRes != 10 {
		t.Errorf("DECR error: %v, res: %d", err, decrRes)
	}

	// Test DECRBY
	decrByRes, err := rdb.DecrBy(ctx, "num1", 2).Result()
	if err != nil || decrByRes != 8 {
		t.Errorf("DECRBY error: %v, res: %d", err, decrByRes)
	}

	// Test APPEND
	appendRes, err := rdb.Append(ctx, "key1", "End").Result()
	if err != nil || appendRes != 7 {
		t.Errorf("APPEND error: %v, newLen: %d", err, appendRes)
	}
}

func TestSetCommands(t *testing.T) {
	rdb := setupRedisClient()
	defer teardown(rdb)

	// Test SADD and SCARD
	err := rdb.SAdd(ctx, "set1", "a", "b", "c").Err()
	if err != nil {
		t.Errorf("SADD failed, err: %v", err)
	}
	card, err := rdb.SCard(ctx, "set1").Result()
	if err != nil || card != 3 {
		t.Errorf("SCARD failed, err: %v, card: %v", err, card)
	}

	// Test SDIFF and SDIFFSTORE
	rdb.SAdd(ctx, "set2", "b", "c", "d")
	diff, err := rdb.SDiff(ctx, "set1", "set2").Result()
	if err != nil || len(diff) != 1 || diff[0] != "a" {
		t.Errorf("SDIFF failed, err: %v, diff: %v", err, diff)
	}
	rdb.SDiffStore(ctx, "set3", "set1", "set2")
	card, err = rdb.SCard(ctx, "set3").Result()
	if err != nil || card != 1 {
		t.Errorf("SDIFFSTORE failed, err: %v, card: %v", err, card)
	}

	// Test SINTER and SINTERSTORE
	inter, err := rdb.SInter(ctx, "set1", "set2").Result()
	if err != nil || len(inter) != 2 {
		t.Errorf("SINTER failed, err: %v, inter: %v", err, inter)
	}
	rdb.SInterStore(ctx, "set4", "set1", "set2")
	card, err = rdb.SCard(ctx, "set4").Result()
	if err != nil || card != 2 {
		t.Errorf("SINTERSTORE failed, err: %v, card: %v", err, card)
	}

	// Test SISMEMBER and SMEMBERS
	isMember, err := rdb.SIsMember(ctx, "set1", "a").Result()
	if err != nil || !isMember {
		t.Errorf("SISMEMBER failed, err: %v, isMember: %v", err, isMember)
	}
	members, err := rdb.SMembers(ctx, "set1").Result()
	if err != nil || len(members) != 3 {
		t.Errorf("SMEMBERS failed, err: %v, members: %v", err, members)
	}

	// Test SMOVE, SPOP, SRANDMEMBER
	rdb.SMove(ctx, "set1", "set5", "a")
	randMember, err := rdb.SRandMember(ctx, "set1").Result()
	if err != nil {
		t.Errorf("SRANDMEMBER failed, err: %v", err)
	} else {
		t.Logf("SRANDMEMBER succeeded: %v", randMember)
	}

	popMember, err := rdb.SPop(ctx, "set1").Result()
	if err != nil {
		t.Errorf("SPOP failed, err: %v", err)
	}

	// Test SREM, SUNION, SUNIONSTORE
	rdb.SRem(ctx, "set1", popMember)
	rdb.SAdd(ctx, "set6", popMember)
	union, err := rdb.SUnion(ctx, "set1", "set6").Result()
	if err != nil || len(union) != 3 {
		t.Errorf("SUNION failed, err: %v, union: %v", err, union)
	}
	rdb.SUnionStore(ctx, "set7", "set1", "set6")
	card, err = rdb.SCard(ctx, "set7").Result()
	if err != nil || card != 3 {
		t.Errorf("SUNIONSTORE failed, err: %v, card: %v", err, card)
	}

	// Test SSCAN
	var cursor uint64
	var keys []string
	for {
		var err error
		keys, cursor, err = rdb.SScan(ctx, "set1", cursor, "", 0).Result()
		if err != nil {
			t.Errorf("SSCAN failed, err: %v", err)
			break
		}
		if cursor == 0 {
			break
		}
	}
	t.Logf("SSCAN succeeded, keys: %v", keys)
}

// 测试 List 类型命令
func TestListCommands(t *testing.T) {
	rdb := setupRedisClient()
	defer teardown(rdb)

	// Setup for tests
	rdb.Del(ctx, "list1", "list2")

	// Test LPUSH and LLEN
	rdb.LPush(ctx, "list1", "a", "b", "c")
	length, err := rdb.LLen(ctx, "list1").Result()
	if err != nil || length != 3 {
		t.Errorf("Expected length 3, got %d, error: %v", length, err)
	}

	// Test LRANGE
	items, err := rdb.LRange(ctx, "list1", 0, -1).Result()
	if err != nil || len(items) != 3 {
		t.Errorf("LRANGE failed, error: %v, items: %v", err, items)
	}

	// Test LPOP and RPOP
	firstItem, err := rdb.LPop(ctx, "list1").Result()
	if err != nil || firstItem != "c" {
		t.Errorf("LPOP failed, error: %v, item: %s", err, firstItem)
	}
	lastItem, err := rdb.RPop(ctx, "list1").Result()
	if err != nil || lastItem != "a" {
		t.Errorf("RPOP failed, error: %v, item: %s", err, lastItem)
	}

	// Test RPUSH and LPUSHX (LPUSHX to existing list)
	rdb.RPush(ctx, "list2", "1")
	rdb.LPushX(ctx, "list2", "0")
	result, err := rdb.LRange(ctx, "list2", 0, -1).Result()
	if err != nil || result[0] != "0" {
		t.Errorf("LPUSHX failed, error: %v, result: %v", err, result)
	}

	// Test BRPOPLPUSH
	rdb.RPush(ctx, "list1", "first", "second")
	rdb.RPush(ctx, "list2", "third")
	poppedItem, err := rdb.BRPopLPush(ctx, "list1", "list2", 0).Result()
	if err != nil || poppedItem != "second" {
		t.Errorf("BRPOPLPUSH failed, error: %v, poppedItem: %s", err, poppedItem)
	}

	// Test LINSERT
	rdb.LInsert(ctx, "list2", "before", "third", "second-and-half")
	newList, err := rdb.LRange(ctx, "list2", 0, -1).Result()
	if err != nil || newList[1] != "second-and-half" {
		t.Errorf("LINSERT failed, error: %v, newList: %v", err, newList)
	}

	// Test LREM
	rdb.LRem(ctx, "list2", 1, "second-and-half")
	remainingItems, err := rdb.LRange(ctx, "list2", 0, -1).Result()
	if err != nil || len(remainingItems) != 3 {
		t.Errorf("LREM failed, error: %v, remainingItems: %v", err, remainingItems)
	}

	// Test LSET and LINDEX
	rdb.LSet(ctx, "list2", 1, "updated-item")
	updatedItem, err := rdb.LIndex(ctx, "list2", 1).Result()
	if err != nil || updatedItem != "updated-item" {
		t.Errorf("LSET or LINDEX failed, error: %v, updatedItem: %s", err, updatedItem)
	}

	// Test LTRIM
	rdb.LTrim(ctx, "list2", 1, 2)
	trimmedList, err := rdb.LRange(ctx, "list2", 0, -1).Result()
	if err != nil || len(trimmedList) != 2 {
		t.Errorf("LTRIM failed, error: %v, trimmedList: %v", err, trimmedList)
	}
}

// 测试 ZSet 类型命令
func TestZSetCommands(t *testing.T) {
	rdb := setupRedisClient()
	defer teardown(rdb)

	// Clear previous data
	rdb.Del(ctx, "zset1")

	// Test ZADD and ZRANGE
	rdb.ZAdd(ctx, "zset1", &redis.Z{Score: 1.0, Member: "one"})
	rdb.ZAdd(ctx, "zset1", &redis.Z{Score: 2.0, Member: "two"})
	results, err := rdb.ZRangeWithScores(ctx, "zset1", 0, -1).Result()
	if err != nil || len(results) != 2 {
		t.Errorf("Expected 2 elements, got %v, error: %v", results, err)
	}

	// Test ZCARD
	count, err := rdb.ZCard(ctx, "zset1").Result()
	if err != nil || count != 2 {
		t.Errorf("Expected count 2, got %d, error: %v", count, err)
	}

	// Test ZRANK
	rank, err := rdb.ZRank(ctx, "zset1", "one").Result()
	if err != nil || rank != 0 {
		t.Errorf("Expected rank 0, got %d, error: %v", rank, err)
	}

	// Test ZINCRBY
	newScore, err := rdb.ZIncrBy(ctx, "zset1", 1.0, "one").Result()
	if err != nil || newScore != 2.0 {
		t.Errorf("Expected new score 2.0, got %f, error: %v", newScore, err)
	}

	// Test ZREMRANGEBYRANK
	rdb.ZRemRangeByRank(ctx, "zset1", 0, 0)
	lenAfterRem, err := rdb.ZCard(ctx, "zset1").Result()
	if err != nil || lenAfterRem != 1 {
		t.Errorf("Expected length 1 after removal, got %d, error: %v", lenAfterRem, err)
	}

	// Test ZREMRANGEBYSCORE
	rdb.ZRemRangeByScore(ctx, "zset1", "1", "2")
	if count, _ = rdb.ZCard(ctx, "zset1").Result(); count != 0 {
		t.Errorf("ZREMRANGEBYSCORE expected count 0, got %d", count)
	}

	// Test ZCOUNT
	rdb.ZAdd(ctx, "zset1", &redis.Z{Score: 5.0, Member: "five"})
	countInRange, err := rdb.ZCount(ctx, "zset1", "2.1", "10.0").Result()
	if err != nil || countInRange != 1 {
		t.Errorf("Expected count in range 1, got %d, error: %v", countInRange, err)
	}

	// Test ZUNIONSTORE
	rdb.ZAdd(ctx, "zset2", &redis.Z{Score: 1.0, Member: "member1"})
	rdb.ZAdd(ctx, "zset2", &redis.Z{Score: 2.0, Member: "member2"})
	rdb.ZAdd(ctx, "zset3", &redis.Z{Score: 1.0, Member: "member1"})
	rdb.ZUnionStore(ctx, "zset1", &redis.ZStore{
		Keys:    []string{"zset2", "zset3"},
		Weights: []float64{1, 1},
	})
	unionScore, _ := rdb.ZScore(ctx, "zset1", "member1").Result()
	if unionScore != 2.0 {
		t.Errorf("ZUNIONSTORE expected score 2.0 for member1, got %f", unionScore)
	}

	// Test ZSCAN (iterate over all elements)
	iter := rdb.ZScan(ctx, "zset1", 0, "", 0).Iterator()
	for iter.Next(ctx) {
		t.Logf("ZSCAN item: %v", iter.Val())
	}
	if err := iter.Err(); err != nil {
		t.Errorf("ZSCAN encountered an error: %v", err)
	}
}

// 更多测试...
