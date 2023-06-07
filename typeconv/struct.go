package typeconv

import (
	"fmt"
	"reflect"
	"strings"
)

type MarshalType uint

const (
	MarshalAsText MarshalType = iota
	MarshalAsNumber
	MarshalAsDecimal
	MarshalAsBoolean
)

type WeakType struct {
	v           any
	marshalType MarshalType
}

func NewWeakType(v any, marshalAs ...MarshalType) *WeakType {
	wt := WeakType{
		v: v,
	}
	if len(marshalAs) > 0 {
		wt.marshalType = marshalAs[0]
		return &wt
	}
	if v != nil {
		t := reflect.TypeOf(v)
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t.Kind() == reflect.Int ||
			t.Kind() == reflect.Int8 ||
			t.Kind() == reflect.Int16 ||
			t.Kind() == reflect.Int32 ||
			t.Kind() == reflect.Int64 ||
			t.Kind() == reflect.Uint ||
			t.Kind() == reflect.Uint8 ||
			t.Kind() == reflect.Uint16 ||
			t.Kind() == reflect.Uint32 ||
			t.Kind() == reflect.Uint64 {
			wt.marshalType = MarshalAsNumber
		} else if t.Kind() == reflect.Float32 ||
			t.Kind() == reflect.Float64 {
			wt.marshalType = MarshalAsDecimal
		} else if t.Kind() == reflect.Bool {
			wt.marshalType = MarshalAsBoolean
		} else {
			wt.marshalType = MarshalAsText
		}
	} else {
		wt.marshalType = MarshalAsText
	}
	return &wt
}

func (v *WeakType) MarshalAs(marshalType MarshalType) {
	v.marshalType = marshalType
}

func (v *WeakType) Int() (int, bool) {
	return Int(v.v)
}

func (v *WeakType) Int8() (int8, bool) {
	return Int8(v.v)
}

func (v *WeakType) Int16() (int16, bool) {
	return Int16(v.v)
}

func (v *WeakType) Int32() (int32, bool) {
	return Int32(v.v)
}

func (v *WeakType) Int64() (int64, bool) {
	return Int64(v.v)
}

func (v *WeakType) Uint() (uint, bool) {
	return Uint(v.v)
}

func (v *WeakType) Uint8() (uint8, bool) {
	return Uint8(v.v)
}

func (v *WeakType) Uint16() (uint16, bool) {
	return Uint16(v.v)
}

func (v *WeakType) Uint32() (uint32, bool) {
	return Uint32(v.v)
}

func (v *WeakType) Uint64() (uint64, bool) {
	return Uint64(v.v)
}

func (v *WeakType) Float32() (float32, bool) {
	return Float32(v.v)
}

func (v *WeakType) Float64() (float64, bool) {
	return Float64(v.v)
}

func (v *WeakType) Boolean() (bool, bool) {
	return Boolean(v.v)
}

func (v *WeakType) String() (string, bool) {
	return String(v.v)
}

func (v *WeakType) MustInt() int {
	return MustInt(v.v)
}

func (v *WeakType) MustInt8() int8 {
	return MustInt8(v.v)
}

func (v *WeakType) MustInt16() int16 {
	return MustInt16(v.v)
}

func (v *WeakType) MustInt32() int32 {
	return MustInt32(v.v)
}

func (v *WeakType) MustInt64() int64 {
	return MustInt64(v.v)
}

func (v *WeakType) MustUint() uint {
	return MustUint(v.v)
}

func (v *WeakType) MustUint8() uint8 {
	return MustUint8(v.v)
}

func (v *WeakType) MustUint16() uint16 {
	return MustUint16(v.v)
}

func (v *WeakType) MustUint32() uint32 {
	return MustUint32(v.v)
}

func (v *WeakType) MustUint64() uint64 {
	return MustUint64(v.v)
}

func (v *WeakType) MustFloat32() float32 {
	return MustFloat32(v.v)
}

func (v *WeakType) MustFloat64() float64 {
	return MustFloat64(v.v)
}

func (v *WeakType) MustBoolean() bool {
	return MustBoolean(v.v)
}

func (v *WeakType) MustString() string {
	return MustString(v.v)
}

func (v *WeakType) Format(formatter func(string) (any, error)) (any, error) {
	strv, ok := v.String()
	if !ok {
		return nil, fmt.Errorf("cannot format %v as string", v)
	}
	return formatter(strv)
}

func (v *WeakType) UnmarshalJSON(data []byte) error {
	content := UnescapeString(string(data))
	if len(data) == 0 || content == "" {
		return nil
	}
	wt := NewWeakType(content)
	if v == nil {
		v = &WeakType{
			v:           wt.v,
			marshalType: wt.marshalType,
		}
	} else {
		*v = *wt
	}
	return nil
}

func (v WeakType) MarshalJSON() ([]byte, error) {
	switch v.marshalType {
	case MarshalAsBoolean:
		b, ok := Boolean(v.v)
		if !ok {
			return nil, fmt.Errorf("cannot marshal %v as boolean", v)
		}
		return []byte(fmt.Sprintf("%t", b)), nil

	case MarshalAsNumber:
		i, ok := Number(v.v)
		if !ok {
			return nil, fmt.Errorf("cannot marshal %v as number", v)
		}
		return []byte(fmt.Sprintf("%d", i)), nil

	case MarshalAsDecimal:
		f, ok := Decimal(v.v)
		if !ok {
			return nil, fmt.Errorf("cannot marshal %v as decimal", v)
		}
		return []byte(fmt.Sprintf("%f", f)), nil

	case MarshalAsText:
		str, ok := String(v.v)
		if !ok {
			return nil, fmt.Errorf("cannot marshal %v as text", v)
		}
		return []byte(fmt.Sprintf("\"%s\"", str)), nil

	default:
		return nil, fmt.Errorf("cannot marshal %v", v)
	}
}

func UnescapeString(v string) string {
	return strings.Replace(v, "\"", "", -1)
}

func PadRight(v string, padChar string, length int) string {
	if len(v) >= length {
		return v
	}
	return fmt.Sprintf("%s%s", v, strings.Repeat(padChar, length-len(v)))
}

func PadLeft(v string, padChar string, length int) string {
	if len(v) >= length {
		return v
	}
	return fmt.Sprintf("%s%s", strings.Repeat(padChar, length-len(v)), v)
}
