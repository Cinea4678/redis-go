package redis

// 工具类的小命令

func cmdPing(client *redisClient) error {
	req := client.reqValue
	if len(req.Elems) < 2 {
		addReplyString(client, "PONG")
	} else {
		addReplyString(client, req.Elems[1].Str)
	}
	return nil
}
