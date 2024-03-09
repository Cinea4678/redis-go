package redis

import (
	"errors"
	"fmt"
)

func getErrorExpireTime(client *redisClient) error {
	return errors.New(fmt.Sprintf("invalid expire time in '%s' command", client.cmd.name))
}
