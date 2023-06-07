package typeconv

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func Int(v interface{}) (int, bool) {
	ret, ok := Convert(reflect.TypeOf(int(0)), v)
	if !ok {
		return 0, false
	}
	return ret.(int), true
}

func Int8(v interface{}) (int8, bool) {
	ret, ok := Convert(reflect.TypeOf(int8(0)), v)
	if !ok {
		return 0, false
	}
	return ret.(int8), true
}

func Int16(v interface{}) (int16, bool) {
	ret, ok := Convert(reflect.TypeOf(int16(0)), v)
	if !ok {
		return 0, false
	}
	return ret.(int16), true
}

func Int32(v interface{}) (int32, bool) {
	ret, ok := Convert(reflect.TypeOf(int32(0)), v)
	if !ok {
		return 0, false
	}
	return ret.(int32), true
}

func Number(v interface{}) (int64, bool) {
	if iv, ok := Int(v); ok {
		return int64(iv), true
	} else if iv, ok := Int8(v); ok {
		return int64(iv), true
	} else if iv, ok := Int16(v); ok {
		return int64(iv), true
	} else if iv, ok := Int32(v); ok {
		return int64(iv), true
	} else if iv, ok := Int64(v); ok {
		return int64(iv), true
	} else if iv, ok := Uint(v); ok {
		return int64(iv), true
	} else if iv, ok := Uint8(v); ok {
		return int64(iv), true
	} else if iv, ok := Uint16(v); ok {
		return int64(iv), true
	} else if iv, ok := Uint32(v); ok {
		return int64(iv), true
	} else if iv, ok := Uint64(v); ok {
		return int64(iv), true
	} else if iv, ok := Float32(v); ok {
		return int64(iv), true
	} else if iv, ok := Float64(v); ok {
		return int64(iv), true
	} else {
		return 0, false
	}
}

func Decimal(v any) (float64, bool) {
	if iv, ok := Int(v); ok {
		return float64(iv), true
	} else if iv, ok := Int8(v); ok {
		return float64(iv), true
	} else if iv, ok := Int16(v); ok {
		return float64(iv), true
	} else if iv, ok := Int32(v); ok {
		return float64(iv), true
	} else if iv, ok := Int64(v); ok {
		return float64(iv), true
	} else if iv, ok := Uint(v); ok {
		return float64(iv), true
	} else if iv, ok := Uint8(v); ok {
		return float64(iv), true
	} else if iv, ok := Uint16(v); ok {
		return float64(iv), true
	} else if iv, ok := Uint32(v); ok {
		return float64(iv), true
	} else if iv, ok := Uint64(v); ok {
		return float64(iv), true
	} else if iv, ok := Float32(v); ok {
		return float64(iv), true
	} else if iv, ok := Float64(v); ok {
		return float64(iv), true
	} else {
		return 0, false
	}
}

func Int64(v interface{}) (int64, bool) {
	ret, ok := Convert(reflect.TypeOf(int64(0)), v)
	if !ok {
		return 0, false
	}
	return ret.(int64), true
}

func Uint(v interface{}) (uint, bool) {
	ret, ok := Convert(reflect.TypeOf(uint(0)), v)
	if !ok {
		return 0, false
	}
	return ret.(uint), true
}

func Uint8(v interface{}) (uint8, bool) {
	ret, ok := Convert(reflect.TypeOf(uint8(0)), v)
	if !ok {
		return 0, false
	}
	return ret.(uint8), true
}

func Uint16(v interface{}) (uint16, bool) {
	ret, ok := Convert(reflect.TypeOf(uint16(0)), v)
	if !ok {
		return 0, false
	}
	return ret.(uint16), true
}

func Uint32(v interface{}) (uint32, bool) {
	ret, ok := Convert(reflect.TypeOf(uint32(0)), v)
	if !ok {
		return 0, false
	}
	return ret.(uint32), true
}

func Uint64(v interface{}) (uint64, bool) {
	ret, ok := Convert(reflect.TypeOf(uint64(0)), v)
	if !ok {
		return 0, false
	}
	return ret.(uint64), true
}

func Float32(v interface{}) (float32, bool) {
	ret, ok := Convert(reflect.TypeOf(float32(0)), v)
	if !ok {
		return 0, false
	}
	return ret.(float32), true
}

func Float64(v interface{}) (float64, bool) {
	ret, ok := Convert(reflect.TypeOf(float64(0)), v)
	if !ok {
		return 0, false
	}
	return ret.(float64), true
}

func Time(v interface{}) (time.Time, bool) {
	if v == nil {
		return time.Time{}, false
	}
	if tv, ok := v.(time.Time); ok {
		return tv, true
	} else if iv, ok := Int(v); ok {
		return time.Unix(int64(iv), 0), true
	} else if tv, err := time.ParseInLocation(`2006-01-02 15:04:05`, fmt.Sprintf("%v", v), time.Local); err == nil {
		return tv, true
	} else {
		return time.Time{}, false
	}
}

func String(v interface{}) (string, bool) {
	if v == nil {
		return "", false
	}
	if ret, ok := v.(string); ok {
		return ret, true
	} else if iv, ok := v.(int); ok {
		return strconv.FormatInt(int64(iv), 10), true
	} else if iv, ok := v.(int8); ok {
		return strconv.FormatInt(int64(iv), 10), true
	} else if iv, ok := v.(int16); ok {
		return strconv.FormatInt(int64(iv), 10), true
	} else if iv, ok := v.(int32); ok {
		return strconv.FormatInt(int64(iv), 10), true
	} else if iv, ok := v.(int64); ok {
		return strconv.FormatInt(int64(iv), 10), true
	} else if iv, ok := v.(uint); ok {
		return strconv.FormatUint(uint64(iv), 10), true
	} else if iv, ok := v.(uint8); ok {
		return strconv.FormatUint(uint64(iv), 10), true
	} else if iv, ok := v.(uint16); ok {
		return strconv.FormatUint(uint64(iv), 10), true
	} else if iv, ok := v.(uint32); ok {
		return strconv.FormatUint(uint64(iv), 10), true
	} else if iv, ok := v.(uint64); ok {
		return strconv.FormatUint(uint64(iv), 10), true
	} else if iv, ok := v.(float32); ok {
		return strconv.FormatFloat(float64(iv), 'f', -1, 32), true
	} else if iv, ok := v.(float64); ok {
		return strconv.FormatFloat(float64(iv), 'f', -1, 64), true
	} else if bv, ok := v.(bool); ok {
		return strconv.FormatBool(bv), true
	} else {
		return fmt.Sprintf("%v", v), true
	}
}

func Boolean(v interface{}) (bool, bool) {
	if iv, ok := Int(v); ok {
		return iv > 0, true
	} else if bv, err := strconv.ParseBool(fmt.Sprintf("%v", v)); err == nil {
		return bv, true
	} else {
		return false, false
	}
}
