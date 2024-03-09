package redis

/*************
	String 操作&命令
 *************/

const (
	objNoFlags = 0
	objSetNX   = 1 << (iota - 1) // Set if key not exists
	objSetXX                     // Set if key exists
	objEX                        // Set if time in seconds is given
	objPX                        // Set if time in ms in given
	objKeepTtl                   // Set and keep the ttl
	objSetGet                    // Set if want to get key before set
	objEXAT                      // Set if timestamp in second is given
	objPXAT                      // Set if timestamp in ms is given
	objPersist                   // Set if we need to remove the ttl
)

// 通用的Set命令处理工具
//func setGenericCommand(client *redisClient, flags int, key *robj, val *robj, expire *robj, unit int, ok_reply *robj, abort_reply *robj) {
//	milliseconds := 0
//	found := 0
//	setkey_falgs := 0
//
//	//if expire != nil &&
//}
//
//func getExpireMillisecondsOrReply(client *redisClient, expire *robj, flags int, unit int) (milliseconds uint64, err error) {
//	milliseconds, err = expire.getUint64()
//	if err != nil {
//		return
//	}
//
//	if unit == unitSeconds && milliseconds > math.MaxUint64/1000 {
//		addReplyErrorExpireTime(client)
//		return 0,
//	}
//
//}
