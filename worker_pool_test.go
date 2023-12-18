package work

import (
	"bytes"
	"reflect"
)

var tstCtxType = reflect.TypeOf(tstCtx{})

type tstCtx struct {
	a int
	bytes.Buffer
}

func (c *tstCtx) record(s string) {
	_, _ = c.WriteString(s)
}
