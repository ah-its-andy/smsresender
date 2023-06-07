package typeconv

import (
	"fmt"
	"time"
)

func MustInt(v interface{}) int {
	if ret, ok := Int(v); ok {
		return ret
	} else {
		panic("typeconv.MustInt: '" + fmt.Sprintf("%v", v) + "' could not be converted to int")
	}
}

func MustInt8(v interface{}) int8 {
	if ret, ok := Int8(v); ok {
		return ret
	} else {
		panic("typeconv.MustInt8: '" + fmt.Sprintf("%v", v) + "' could not be converted to int8")
	}
}

func MustInt16(v interface{}) int16 {
	if ret, ok := Int16(v); ok {
		return ret
	} else {
		panic("typeconv.MustInt16: '" + fmt.Sprintf("%v", v) + "' could not be converted to int16")
	}
}

func MustInt32(v interface{}) int32 {
	if ret, ok := Int32(v); ok {
		return ret
	} else {
		panic("typeconv.MustInt32: '" + fmt.Sprintf("%v", v) + "' could not be converted to int32")
	}
}

func MustInt64(v interface{}) int64 {
	if ret, ok := Int64(v); ok {
		return ret
	} else {
		panic("typeconv.MustInt64: '" + fmt.Sprintf("%v", v) + "' could not be converted to int64")
	}
}

func MustUint(v interface{}) uint {
	if ret, ok := Uint(v); ok {
		return ret
	} else {
		panic("typeconv.MustUint: '" + fmt.Sprintf("%v", v) + "' could not be converted to uint")
	}
}

func MustUint8(v interface{}) uint8 {
	if ret, ok := Uint8(v); ok {
		return ret
	} else {
		panic("typeconv.MustUint8: '" + fmt.Sprintf("%v", v) + "' could not be converted to uint8")
	}
}

func MustUint16(v interface{}) uint16 {
	if ret, ok := Uint16(v); ok {
		return ret
	} else {
		panic("typeconv.MustUint16: '" + fmt.Sprintf("%v", v) + "' could not be converted to uint16")
	}
}

func MustUint32(v interface{}) uint32 {
	if ret, ok := Uint32(v); ok {
		return ret
	} else {
		panic("typeconv.MustUint32: '" + fmt.Sprintf("%v", v) + "' could not be converted to uint32")
	}
}

func MustUint64(v interface{}) uint64 {
	if ret, ok := Uint64(v); ok {
		return ret
	} else {
		panic("typeconv.MustUint64: '" + fmt.Sprintf("%v", v) + "' could not be converted to uint64")
	}
}

func MustFloat32(v interface{}) float32 {
	if ret, ok := Float32(v); ok {
		return ret
	} else {
		panic("typeconv.MustFloat32: '" + fmt.Sprintf("%v", v) + "' could not be converted to float32")
	}
}

func MustFloat64(v interface{}) float64 {
	if ret, ok := Float64(v); ok {
		return ret
	} else {
		panic("typeconv.MustFloat64: '" + fmt.Sprintf("%v", v) + "' could not be converted to float64")
	}
}

func MustString(v interface{}) string {
	if ret, ok := String(v); ok {
		return ret
	} else {
		panic("typeconv.MustString: '" + fmt.Sprintf("%v", v) + "' could not be converted to string")
	}
}

func MustBoolean(v interface{}) bool {
	if ret, ok := Boolean(v); ok {
		return ret
	} else {
		panic("typeconv.MustBoolean: '" + fmt.Sprintf("%v", v) + "' could not be converted to bool")
	}
}

func MustTime(v interface{}) time.Time {
	if ret, ok := Time(v); ok {
		return ret
	} else {
		panic("typeconv.MustTime: '" + fmt.Sprintf("%v", v) + "' could not be converted to time.Time")
	}
}
