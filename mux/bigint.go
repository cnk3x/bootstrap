package mux

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
)

//BigInt BigInt
type BigInt int64

//Scan can scan
func (b *BigInt) Scan(src interface{}) error {
	switch v := src.(type) {
	case string:
		var i int64
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		*b = BigInt(i)
	case []byte:
		var i int64
		i, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return err
		}
		*b = BigInt(i)
	case int64:
		*b = BigInt(v)
	default:
		return fmt.Errorf("not bigint: %v", src)
	}
	return nil
}

//Value Value
func (b BigInt) Value() (driver.Value, error) {
	return int64(b), nil
}

//UnmarshalJSON UnmarshalJSON
func (b *BigInt) UnmarshalJSON(bytes []byte) error {
	var s string
	err := json.Unmarshal(bytes, &s)
	if err == nil {
		var i int64
		i, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*b = BigInt(i)
	}
	return err
}

//MarshalJSON MarshalJSON
func (b BigInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(b), 10))
}

func (b BigInt) String() string {
	return strconv.FormatInt(int64(b), 10)
}

//Int Int
func (b BigInt) Int() int64 {
	return int64(b)
}

var _ fmt.Stringer = BigInt(0)
var _ json.Marshaler = BigInt(0)
var _ json.Unmarshaler = (*BigInt)(nil)
var _ driver.Valuer = BigInt(0)
var _ sql.Scanner = (*BigInt)(nil)
