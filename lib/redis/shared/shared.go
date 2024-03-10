package shared

import (
	"github.com/cinea4678/resp3"
	"redis-go/lib/redis/core"
)

var (
	Shared *ValuesStruct
)

// ValuesStruct 共享对象
type ValuesStruct struct {
	Ok        *resp3.Value
	err       *resp3.Value
	syntaxErr *resp3.Value
	nullBulk  *resp3.Value
	cZero     *resp3.Value
	cOne      *resp3.Value
	oomErr    *resp3.Value
}

func CreateSharedValues() {
	Shared = &ValuesStruct{
		Ok:        &resp3.Value{Type: resp3.TypeSimpleString, Str: "OK"},
		err:       &resp3.Value{Type: resp3.TypeSimpleError, Str: "ERR"},
		syntaxErr: &resp3.Value{Type: resp3.TypeSimpleError, Str: "-ERR syntax error"},
		cZero:     &resp3.Value{Type: resp3.TypeNumber, Integer: 0},
		cOne:      &resp3.Value{Type: resp3.TypeNumber, Integer: 1},
		oomErr:    &resp3.Value{Type: resp3.TypeSimpleError, Str: "-OOM command not allowed when used memory > 'maxmemory'"},
	}
}

var Server = &core.RedisServer{}
