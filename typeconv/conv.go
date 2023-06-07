package typeconv

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func Convert(t reflect.Type, v interface{}) (interface{}, bool) {
	if v == nil {
		return reflect.Zero(t).Interface(), false
	}

	if reflect.TypeOf(v).ConvertibleTo(t) {
		return reflect.ValueOf(v).Convert(t).Interface(), true
	}

	if t.Kind() == reflect.Int ||
		t.Kind() == reflect.Int8 ||
		t.Kind() == reflect.Int16 ||
		t.Kind() == reflect.Int32 ||
		t.Kind() == reflect.Int64 {
		if i, ok := v.(time.Time); ok {
			ret := i.Unix()
			return reflect.ValueOf(ret).Convert(t).Interface(), true
		} else {
			strV := fmt.Sprintf("%v", v)
			ret, err := strconv.ParseInt(strV, 10, 64)
			if err != nil {
				return reflect.Zero(t).Interface(), false
			}
			return reflect.ValueOf(ret).Convert(t).Interface(), true
		}
	} else if t.Kind() == reflect.Uint ||
		t.Kind() == reflect.Uint8 ||
		t.Kind() == reflect.Uint16 ||
		t.Kind() == reflect.Uint32 ||
		t.Kind() == reflect.Uint64 {
		if i, ok := v.(time.Time); ok {
			ret := i.Unix()
			return reflect.ValueOf(ret).Convert(t).Interface(), true
		} else {
			strV := fmt.Sprintf("%v", v)
			ret, err := strconv.ParseUint(strV, 10, 64)
			if err != nil {
				return reflect.Zero(t).Interface(), false
			}
			return reflect.ValueOf(ret).Convert(t).Interface(), true
		}
	} else if t.Kind() == reflect.Float32 ||
		t.Kind() == reflect.Float64 {
		if i, ok := v.(time.Time); ok {
			ret := i.Unix()
			return reflect.ValueOf(ret).Convert(t).Interface(), true
		} else {
			strV := fmt.Sprintf("%v", v)
			ret, err := strconv.ParseFloat(strV, 64)
			if err != nil {
				return reflect.Zero(t).Interface(), false
			}
			return reflect.ValueOf(ret).Convert(t).Interface(), true
		}
	}
	return reflect.Zero(t).Interface(), false
}
