package strs

import (
	"bytes"
	"reflect"
	"strconv"
	"unsafe"

	"go.shu.run/bootstrap/utils/json"
)

type V []byte

// B converts byte slice to string without a memory allocation.
func B(b []byte) V {
	return b
}

// S converts string to byte slice without a memory allocation.
func S(s string) (v V) {
	sh := *(*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&v))
	bh.Data, bh.Len, bh.Cap = sh.Data, sh.Len, sh.Len
	return v
}

func (v V) String() string {
	return *(*string)(unsafe.Pointer(&v))
}

func (v V) Bool(def bool) bool {
	b, err := strconv.ParseBool(v.String())
	if err != nil {
		return def
	}
	return b
}

func (v V) Int(def int64) int64 {
	r, err := strconv.ParseInt(v.String(), 10, 64)
	if err != nil {
		return def
	}
	return r
}

func (v V) Float(def float64) float64 {
	r, err := strconv.ParseFloat(v.String(), 64)
	if err != nil {
		return def
	}
	return r
}

func (v V) Decode(i interface{}) error {
	return json.Unmarshal(v, i)
}

func (v V) Eq(target []byte) bool {
	return bytes.Equal(v, target)
}
