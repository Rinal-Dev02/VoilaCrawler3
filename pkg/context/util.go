package context

import (
	"context"
	"reflect"
)

func RetrieveAllValues(ctx context.Context) map[interface{}]interface{} {
	kvs := map[interface{}]interface{}{}
	setKV := func(k, v interface{}) {
		if _, ok := kvs[k]; !ok {
			kvs[k] = v
		}
	}

	var retrive func(ctx context.Context)
	retrive = func(ctx context.Context) {
		val := reflect.ValueOf(ctx)
		if !val.IsValid() || val.IsNil() {
			return
		}
		val = val.Elem()

		switch val.Type().Name() {
		case "valueCtx":
			setKV(val.FieldByName("key").Elem().String(), val.FieldByName("val").Elem().String())
		case "valuesCtx":
			rtVals := val.FieldByName("values")
			for i := 0; i < rtVals.Len(); i += 2 {
				setKV(rtVals.Index(i).String(), rtVals.Index(i+1).String())
			}
		}

		if val.FieldByName("Context").Elem().Elem().Kind() == reflect.Struct {
			pv := val.FieldByName("Context").Interface()
			retrive(pv.(context.Context))
		}
	}
	retrive(ctx)

	return kvs
}
