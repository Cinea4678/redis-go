package core

import (
	"github.com/cinea4678/resp3"
	"github.com/thoas/go-funk"
)

// RedisCommand redis命令结构
type RedisCommand struct {
	Name            string                          // 命令名称
	RedisClientFunc func(client *RedisClient) error // 命令处理函数
}

// RedisCommandInfo 用于反馈Command命令的命令信息
type RedisCommandInfo struct {
	Name             string   // 命令名称
	Arity            int16    // 参数数量(固定数量使用正数,有可选参数则是最小数量的相反数,数量包含命令本身)
	Flags            []string // 参数标志,具体看https://t.ly/aqXqb (可以偷懒找个Redis输入command命令看看)
	FirstKeyPosition int16    // 第一个键参数的位置 (相关解释请看:https://t.ly/QT-_K)
	LastKeyPosition  int16    // 最后一个键参数的位置 (相关解释请看:https://t.ly/QT-_K)
	StepCount        int16    // 步长 (相关解释请看:https://t.ly/QT-_K)
}

// NewRedisCommandInfo 工厂函数
func NewRedisCommandInfo(name string, arity int16, flags []string, firstKeyPosition int16, lastKeyPosition int16, stepCount int16) *RedisCommandInfo {
	return &RedisCommandInfo{
		Name:             name,
		Arity:            arity,
		Flags:            flags,
		FirstKeyPosition: firstKeyPosition,
		LastKeyPosition:  lastKeyPosition,
		StepCount:        stepCount,
	}
}

func RedisCommandInfoToValue(commands []*RedisCommandInfo) *resp3.Value {
	command_value := make([]*resp3.Value, 0, len(commands))
	for _, c := range commands {
		flags := funk.Map(c.Flags, func(f string) *resp3.Value {
			return resp3.NewSimpleStringValue(f)
		}).([]*resp3.Value)

		values := []*resp3.Value{
			resp3.NewSimpleStringValue(c.Name),
			resp3.NewNumberValue(int64(c.Arity)),
			resp3.NewArrayValue(flags),
			resp3.NewNumberValue(int64(c.FirstKeyPosition)),
			resp3.NewNumberValue(int64(c.LastKeyPosition)),
			resp3.NewNumberValue(int64(c.StepCount)),
		}

		v := resp3.NewArrayValue(values)
		command_value = append(command_value, v)
	}

	return resp3.NewArrayValue(command_value)
}
