package json

import (
	"errors"
	"redis-go/lib/redis/core"
	"redis-go/lib/redis/io"
	"reflect"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

var (
	errNotEnoughArgs = errors.New("not enough args")
	errNoSuchElement = errors.New("no such element")
	errNotJson       = errors.New("element not json")
)

func JGet(client *core.RedisClient) (err error) {
	req := client.ReqValue.Elems[1:]
	if len(req) < 2 {
		return errNotEnoughArgs
	}

	key := req[0].Str
	jpath := req[1].Str

	value := client.Db.LookupKey(key)
	if value == nil {
		return errNoSuchElement
	}

	vs, err := value.GetString()
	if err != nil {
		return err
	}

	if !gjson.Valid(vs) {
		return errNotJson
	}
	res := gjson.Get(vs, jpath)

	// 处理返回. 策略: 对简单对象,使用RESP3直接返回
	// 对复杂对象, 序列化为json后返回
	if res.IsArray() || res.IsObject() {
		json, err := jsoniter.MarshalToString(res.Value())
		if err != nil {
			return err
		}

		io.AddReplyString(client, json)
	} else if res.IsBool() {
		io.AddReplyBool(client, res.Bool())
	} else {
		switch v := res.Value().(type) {
		case float64:
			io.AddReplyFloat(client, v)
		case string:
			io.AddReplyString(client, v)
		case nil:
			io.AddReplyNull(client)
		default:
			log.Warn().Str("type", reflect.TypeOf(v).Name()).Msg("jget unknown type")
		}
	}

	return nil
}
