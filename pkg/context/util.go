package context

import (
	"context"
	"reflect"
	"unsafe"

	"github.com/voiladev/go-framework/strconv"
)

// GetString
func GetString(ctx context.Context, key interface{}) string {
	if ctx == nil || key == nil {
		return ""
	}
	val := ctx.Value(key)
	if val == nil {
		return ""
	}
	v, _ := val.(string)
	return v
}

// GetInt
func GetInt(ctx context.Context, key interface{}) int64 {
	if ctx == nil || key == nil {
		return 0
	}
	val := ctx.Value(key)
	if val == nil {
		return 0
	}
	v, _ := strconv.ParseInt(val)
	return v
}

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
			kVal := val.FieldByName("key")
			vVal := val.FieldByName("val")

			setKV(
				reflect.NewAt(kVal.Type(), unsafe.Pointer(kVal.UnsafeAddr())).Elem().Interface(),
				reflect.NewAt(vVal.Type(), unsafe.Pointer(vVal.UnsafeAddr())).Elem().Interface())
		case "valuesCtx":
			rtVals := val.FieldByName("values")
			for i := 0; i < rtVals.Len(); i += 2 {
				setKV(rtVals.Index(i).String(), rtVals.Index(i+1).String())
			}
		}

		if val.Kind() == reflect.Struct || val.Kind() == reflect.Ptr {
			if val.FieldByName("Context").Elem().Elem().Kind() == reflect.Struct {
				pv := val.FieldByName("Context").Interface()
				retrive(pv.(context.Context))
			}
		}
	}
	retrive(ctx)

	return kvs
}
