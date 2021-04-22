package context

import (
	"context"
	"strings"
)

type valuesCtx struct {
	context.Context
	values []string
}

func WithValues(parent context.Context, keyvals ...string) context.Context {
	if parent == nil {
		panic("cannot create context from nil parent")
	}

	if len(keyvals)%2 != 0 {
		panic("key value pair not match")
	}

	vals := make([]string, len(keyvals))
	vals = append(vals[0:0], keyvals...)

	return &valuesCtx{parent, vals}
}

func (c *valuesCtx) String() string {
	return "valuesCtx.WithValue([" + strings.Join(c.values, ", ") + "])"
}

func (c *valuesCtx) Value(key interface{}) interface{} {
	if ks, ok := key.(string); ok {
		for i := 0; i < len(c.values); i += 2 {
			if c.values[i] == ks {
				return c.values[i+1]
			}
		}
	}
	return c.Context.Value(key)
}
